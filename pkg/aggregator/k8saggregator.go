package aggregator

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/request"
	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/wharfapi"
	"github.com/iver-wharf/wharf-cmd/internal/parallel"
	"github.com/iver-wharf/wharf-cmd/pkg/config"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerclient"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"gopkg.in/typ.v4/sync2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport/spdy"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8sruntime "k8s.io/apimachinery/pkg/util/runtime"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var log = logger.NewScoped("AGGREGATOR")

// NewK8sAggregator returns a new Aggregator implementation that targets
// Kubernetes using a specific Kubernetes namespace and REST config.
func NewK8sAggregator(config *config.Config, restConfig *rest.Config) (Aggregator, error) {
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	roundTripper, upgrader, err := spdy.RoundTripperFor(restConfig)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{
		Transport: roundTripper,
	}
	if err != nil {
		return nil, err
	}
	return k8sAggr{
		aggrConfig: config.Aggregator,
		namespace:  config.K8s.Namespace,
		clientset:  clientset,
		pods:       clientset.CoreV1().Pods(config.K8s.Namespace),
		restConfig: restConfig,

		upgrader:   upgrader,
		httpClient: httpClient,
		instanceID: config.InstanceID,
		listOptionsMatchLabels: metav1.ListOptions{
			LabelSelector: "app.kubernetes.io/name=wharf-cmd-worker," +
				"app.kubernetes.io/managed-by=wharf-cmd-provisioner," +
				"wharf.iver.com/instance=" + config.InstanceID,
		},

		wharfapi: wharfapi.Client{
			APIURL: config.Aggregator.WharfAPIURL,
		},

		inProgress: &sync2.Set[types.UID]{},
	}, nil
}

type k8sAggr struct {
	aggrConfig config.AggregatorConfig

	namespace string
	clientset *kubernetes.Clientset
	pods      corev1.PodInterface

	restConfig             *rest.Config
	upgrader               spdy.Upgrader
	httpClient             *http.Client
	instanceID             string
	listOptionsMatchLabels metav1.ListOptions

	wharfapi wharfapi.Client

	inProgress *sync2.Set[types.UID]
}

func (a k8sAggr) Serve(ctx context.Context) error {
	const pollDelay = 5 * time.Second
	log.Info().
		WithDuration("pollDelay", pollDelay).
		Message("Aggregator started.")

	// Silences the output of error messages from internal k8s code to console.
	//
	// The console gets clogged with forwarding errors when attempting to ping
	// a worker while its server wasn't running.
	k8sruntime.ErrorHandlers = []func(error){}

	for {
		// TODO: Wait for Wharf API to be up first, with sane infinite retry logic.
		//
		// Would prevent pod listing and opening a tunnel to each pod each
		// iteration.

		pods, err := a.fetchPods(ctx)
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return ctx.Err()
			}
			log.Warn().WithError(err).
				WithDuration("pollDelay", pollDelay).
				Message("Failed to list pods. Retrying after delay.")
			time.Sleep(pollDelay)
			continue
		}
		for _, pod := range pods {
			if err := a.handlePod(ctx, pod); err != nil {
				return err
			}
		}
		time.Sleep(pollDelay)
	}
}

func (a k8sAggr) handlePod(ctx context.Context, pod v1.Pod) error {
	buildID, err := k8sParsePodBuildID(pod.ObjectMeta)
	if err != nil {
		log.Warn().WithError(err).
			WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
			Message("Failed to parse worker's build ID.")
		return nil
	}
	if !a.inProgress.Add(pod.UID) {
		// Failed to add => Already being processed
		return nil
	}

	log.Debug().
		WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
		WithString("status", string(pod.Status.Phase)).
		Message("Pod found.")

	portConn, err := newPortForwarding(a, a.namespace, pod.Name)
	if err != nil {
		return err
	}
	defer portConn.Close()

	worker, err := a.newWorkerClient(portConn, buildID)
	if err != nil {
		return err
	}
	defer worker.Close()

	apiLogStream, err := a.wharfapi.CreateBuildLogStream(ctx)
	if err != nil {
		return fmt.Errorf("open logs stream to wharf-api: %w", err)
	}
	defer func() {
		resp, err := apiLogStream.CloseAndRecv()
		if err != nil {
			log.Warn().
				WithError(err).
				Message("Unexpected error when closing log writer stream to wharf-api.")
			return
		}
		log.Debug().
			WithUint("inserted", resp.LogsInserted).
			Message("Inserted logs into wharf-api.")
	}()

	errorLogEv := log.Error().
		WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
		WithString("phase", string(pod.Status.Phase))

	switch pod.Status.Phase {
	case v1.PodRunning:
		if err := a.handleRunningPod(ctx, apiLogStream, worker, pod); err != nil {
			errorLogEv.Message("Error handling pod.")
		}
	case v1.PodFailed:
		go func(pod v1.Pod) {
			err := relay[request.Log](k8sRawLogSource{ctx, buildID, pod, a.pods}, func(v request.Log) error {
				return apiLogStream.Send(v)
			})
			if err != nil {
				errorLogEv.Message("Error handling pod.")
			}
			a.terminatePod(ctx, pod)
		}(pod)
	default:
		go func(pod v1.Pod) {
			err := relay[request.Log](k8sEventsSource{ctx, buildID, pod, a.clientset.CoreV1().Events(a.namespace)}, func(v request.Log) error {
				return apiLogStream.Send(v)
			})
			if err != nil {
				errorLogEv.Message("Error handling pod.")
			}
			a.terminatePod(ctx, pod)
		}(pod)
	}
	return nil
}

func (a k8sAggr) handleRunningPod(ctx context.Context, apiLogStream wharfapi.CreateBuildLogStream, worker workerclient.Client, pod v1.Pod) error {
	if err := worker.Ping(ctx); err != nil {
		log.Debug().
			WithStringf("pod", "%s/%s", a.namespace, pod.Name).
			Message("Failed to ping worker pod. Assuming it's not running yet. Skipping.")
		a.inProgress.Remove(pod.UID)
		return nil
	}

	pg := parallel.Group{}
	pg.AddFunc("logs", func(pod v1.Pod) parallel.Func {
		return func(ctx context.Context) error {
			err := relay[request.Log](logLineSource{ctx, worker}, func(v request.Log) error {
				return apiLogStream.Send(v)
			})
			if err != nil {
				log.Error().WithError(err).
					WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
					Message("Relay logs error.")
			}
			return err
		}
	}(pod))
	pg.AddFunc("status events", func(pod v1.Pod) parallel.Func {
		return func(ctx context.Context) error {
			previousStatus := request.BuildScheduling
			updateStatus := func(newStatus request.BuildStatus) error {
				statusUpdate := request.LogOrStatusUpdate{Status: newStatus}
				if _, err := a.wharfapi.UpdateBuildStatus(worker.BuildID(), statusUpdate); err != nil {
					return err
				}
				log.Info().
					WithString("new", string(newStatus)).
					WithString("previous", string(previousStatus)).
					Message("Updated build status.")
				previousStatus = newStatus
				return nil
			}
			err := relay[request.BuildStatus](statusSource{ctx, worker}, func(newStatus request.BuildStatus) error {
				if newStatus == request.BuildFailed || newStatus == request.BuildCompleted {
					return updateStatus(newStatus)
				}
				if newStatus == request.BuildRunning && previousStatus == request.BuildScheduling {
					if err := updateStatus(request.BuildRunning); err != nil {
						return err
					}
				}
				return nil
			})
			if previousStatus != request.BuildCompleted && previousStatus != request.BuildFailed {
				if err := updateStatus(request.BuildFailed); err != nil {
					return err
				}
			}
			if err != nil {
				log.Error().WithError(err).
					WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
					Message("Relay statuses error.")
			}
			return err
		}
	}(pod))
	pg.AddFunc("artifact events", func(pod v1.Pod) parallel.Func {
		return func(ctx context.Context) error {
			err := relay[*workerclient.ArtifactEvent](artifactSource{ctx, worker}, func(v *workerclient.ArtifactEvent) error {
				// No way to send to wharf DB through stream currently
				// so we're just logging it here.
				log.Debug().
					WithUint64("step", v.StepID).
					WithString("name", v.Name).
					WithUint64("id", v.ArtifactID).
					Message("Received artifact event.")
				return nil
			})
			return err
		}
	}(pod))
	if err := pg.RunCancelEarly(ctx); err != nil {
		return err
	}
	return a.terminatePod(ctx, pod)
}

func (a k8sAggr) fetchPods(ctx context.Context) ([]v1.Pod, error) {
	list, err := a.pods.List(ctx, a.listOptionsMatchLabels)
	if err != nil {
		return nil, err
	}
	var pods []v1.Pod
	for _, pod := range list.Items {
		// Skip terminating pods
		if pod.ObjectMeta.DeletionTimestamp != nil {
			continue
		}

		if pod.Status.Phase == v1.PodPending && k8sPodNotErrored(pod) {
			continue
		}

		pods = append(pods, pod)
	}
	return pods, nil
}

func (a k8sAggr) newWorkerClient(portConn portConnection, buildID uint) (workerclient.Client, error) {
	// Intentionally "localhost" because we're port-forwarding
	return workerclient.New(fmt.Sprintf("http://localhost:%d", portConn.Local), workerclient.Options{
		// Skipping security because we've already authenticated with Kubernetes
		// and are communicating through a secured port-forwarding tunnel.
		// Don't need to add TLS on top of TLS.
		InsecureSkipVerify: true,
		BuildID:            buildID,
	})
}

func (a k8sAggr) terminatePod(ctx context.Context, pod v1.Pod) error {
	log.Debug().
		WithStringf("pod", "%s/%s", a.namespace, pod.Name).
		Message("Done relaying. Terminating pod.")

	if err := a.pods.Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("terminate pod after done with relay build results: %w", err)
	}

	log.Info().
		WithStringf("pod", "%s/%s", a.namespace, pod.Name).
		Message("Done with worker.")

	a.inProgress.Remove(pod.UID)
	return nil
}
