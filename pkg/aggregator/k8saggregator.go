package aggregator

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/request"
	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/wharfapi"
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
			go func(pod v1.Pod) {
				if err := a.handlePod(ctx, pod); err != nil {
					log.Error().
						WithError(err).
						WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
						WithString("phase", string(pod.Status.Phase)).
						Message("Error handling pod.")
				}
			}(pod)
		}
		time.Sleep(pollDelay)
	}
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
	defer a.inProgress.Remove(pod.UID)

	log.Debug().
		WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
		WithString("status", string(pod.Status.Phase)).
		Message("Pod found.")

	apiLogStream, err := wharfapi.CreateBuildLogStream(ctx)
	if err != nil {
		return fmt.Errorf("open logs stream to wharf-api: %w", err)
	}
	defer closeLogStream(apiLogStream)

	switch pod.Status.Phase {
	case v1.PodRunning:
		worker, err := newPortForwardedWorker(a, pod.Name, buildID)
		if err != nil {
			return err
		}
		defer worker.Close()
		return a.relayAndTerminateOnSuccess(ctx, pod, func() error {
			return handleRunningPod(ctx, a.namespace, a.wharfapi, apiLogStream, worker, pod)
		})
	case v1.PodFailed:
		return a.relayAndTerminateOnSuccess(ctx, pod, func() error {
			source := k8sRawLogSource{ctx, buildID, pod, a.pods}
			return relay[request.Log](source, func(v request.Log) error {
				return apiLogStream.Send(v)
			})
		})
	default:
		return a.relayAndTerminateOnSuccess(ctx, pod, func() error {
			events := a.clientset.CoreV1().Events(a.namespace)
			source := k8sEventsSource{ctx, buildID, pod, events}
			return relay[request.Log](source, func(v request.Log) error {
				return apiLogStream.Send(v)
			})
		})
	}
}

func (a k8sAggr) relayAndTerminateOnSuccess(ctx context.Context, pod v1.Pod, f func() error) error {
	err := f()
	if err != nil {
		return err
	}
	if err := a.terminatePod(ctx, pod); err != nil {
		log.Error().WithError(err).Message("Failed terminating pod.")
	}
	return nil
}

func (a k8sAggr) newWorkerClient(workerPort uint16, buildID uint) (workerclient.Client, error) {
	// Intentionally "localhost" because we're port-forwarding
	return workerclient.New(fmt.Sprintf("http://localhost:%d", workerPort), workerclient.Options{
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
