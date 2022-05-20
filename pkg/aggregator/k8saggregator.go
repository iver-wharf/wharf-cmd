package aggregator

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/wharfapi"
	"github.com/iver-wharf/wharf-cmd/internal/parallel"
	"github.com/iver-wharf/wharf-cmd/pkg/config"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"gopkg.in/typ.v4/sync2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport/spdy"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8sruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
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

	restConfig *rest.Config

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

		running, failed, err := a.fetchRunningAndFailedPods(ctx)
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

		a.handleRunningPodSlice(ctx, running)
		a.handleFailedPodSlice(ctx, failed)

		time.Sleep(pollDelay)
	}
}

type workerPod struct {
	v1.Pod
	buildID uint
}

func (a k8sAggr) fetchRunningAndFailedPods(ctx context.Context) (running []workerPod, failed []workerPod, err error) {
	list, err := a.pods.List(ctx, a.listOptionsMatchLabels)
	if err != nil {
		return nil, nil, err
	}
	for _, pod := range list.Items {
		if k8sShouldSkipPod(pod) {
			continue
		}

		if !a.inProgress.Add(pod.UID) {
			continue
		}

		log.Debug().
			WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
			WithString("status", string(pod.Status.Phase)).
			Message("Pod found.")

		buildID, err := k8sParsePodBuildID(pod.ObjectMeta)
		if err != nil {
			log.Warn().
				WithError(err).
				WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
				Message("Failed parsing build ID from pod. Skipping.")
			a.inProgress.Remove(pod.UID)
			continue
		}

		p := workerPod{pod, buildID}
		if pod.Status.Phase == v1.PodRunning {
			running = append(running, p)
		} else {
			failed = append(failed, p)
		}
	}
	return running, failed, nil
}

func (a k8sAggr) handleRunningPodSlice(ctx context.Context, pods []workerPod) {
	for _, pod := range pods {
		go func(pod workerPod) {
			if err := a.handleRunningPod(ctx, pod); err != nil {
				log.Error().
					WithError(err).
					WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
					Message("Error handling running pod.")
			}
		}(pod)
	}
}

func (a k8sAggr) handleRunningPod(ctx context.Context, pod workerPod) error {
	defer a.inProgress.Remove(pod.UID)

	worker, err := newPortForwardedWorker(a, pod.Name, pod.buildID)
	if err != nil && pod.Status.Phase == v1.PodRunning {
		return err
	}
	defer worker.Close()

	if err := worker.Ping(ctx); err != nil {
		log.Debug().
			WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
			Message("Failed to ping worker pod. Assuming it's not running yet. Skipping.")
		return nil
	}

	pg := parallel.Group{}
	pg.AddFunc("logs", func(ctx context.Context) error {
		logsPiper, err := newLogsPiper(ctx, a.wharfapi, worker, pod.buildID)
		if err != nil {
			return err
		}
		return pipeAndClose(logsPiper)
	})
	pg.AddFunc("status events", func(ctx context.Context) error {
		statusEventsPiper, err := newStatusEventsPiper(ctx, a.wharfapi, worker)
		if err != nil {
			return err
		}
		return pipeAndClose(statusEventsPiper)
	})
	pg.AddFunc("artifact events", func(ctx context.Context) error {
		artifactEventsPiper, err := newArtifactEventsPiper(ctx, a.wharfapi, worker)
		if err != nil {
			return err
		}
		return pipeAndClose(artifactEventsPiper)
	})
	if err := pg.RunCancelEarly(ctx); err != nil {
		return err
	}
	return a.terminatePod(ctx, pod)
}

func (a k8sAggr) handleFailedPodSlice(ctx context.Context, pods []workerPod) {
	for _, pod := range pods {
		go func(pod workerPod) {
			if err := a.handleFailedPod(ctx, pod); err != nil {
				log.Error().
					WithError(err).
					WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
					Message("Error handling failed pod.")
			}
		}(pod)
	}
}

func (a k8sAggr) handleFailedPod(ctx context.Context, pod workerPod) error {
	defer a.inProgress.Remove(pod.UID)

	logsWriter, err := newLogsWriter(ctx, a.wharfapi, pod.buildID)
	if err != nil {
		return err
	}
	defer logsWriter.Close()

	if err := logsWriter.write(fmt.Sprintf("[aggregator] Pod '%s/%s' failed.", pod.Namespace, pod.Name)); err != nil {
		return err
	}
	if err := logsWriter.write(""); err != nil {
		return err
	}
	if err := logsWriter.write("[aggregator] Logging kubernetes events:"); err != nil {
		return err
	}
	if err := logsWriter.write(""); err != nil {
		return err
	}
	eventsReader, err := a.getEventsReader(pod)
	if err != nil {
		return fmt.Errorf("get events reader: %w", err)
	}
	if err := logsWriter.pipeAndCloseReader(eventsReader); err != nil {
		return fmt.Errorf("write events: %w", err)
	}

	writeContainerLogs := func(c v1.Container) error {
		if err := logsWriter.write(fmt.Sprintf("[aggregator] Logging kubernetes logs from container %q:", c.Name)); err != nil {
			return err
		}
		if err := logsWriter.write(""); err != nil {
			return err
		}
		logsReader, err := a.getLogsReader(ctx, pod, c)
		if err != nil {
			return fmt.Errorf("get logs reader: %w", err)
		}
		if err := logsWriter.pipeAndCloseReader(logsReader); err != nil {
			return err
		}
		return logsWriter.write("")
	}

	if err := logsWriter.write(""); err != nil {
		return err
	}

	for _, c := range pod.Spec.InitContainers {
		if err := writeContainerLogs(c); err != nil {
			return fmt.Errorf("write container logs: %w", err)
		}
	}
	for _, c := range pod.Spec.Containers {
		if err := writeContainerLogs(c); err != nil {
			return fmt.Errorf("write container logs: %w", err)
		}
	}

	return a.terminatePod(ctx, pod)
}

func (a k8sAggr) getEventsReader(pod workerPod) (io.ReadCloser, error) {
	eventsList, err := a.clientset.CoreV1().Events(pod.Namespace).Search(scheme.Scheme, &pod.Pod)
	if err != nil {
		return nil, err
	}
	return describeEvents(eventsList), nil
}

func (a k8sAggr) getLogsReader(ctx context.Context, pod workerPod, container v1.Container) (io.ReadCloser, error) {
	req := a.pods.GetLogs(pod.Name, &v1.PodLogOptions{
		Container: container.Name,
	})
	readCloser, err := req.Stream(ctx)
	if err != nil {
		return nil, err
	}
	return readCloser, nil
}

func (a k8sAggr) terminatePod(ctx context.Context, pod workerPod) error {
	log.Debug().
		WithStringf("pod", "%s/%s", a.namespace, pod.Name).
		Message("Done relaying. Terminating pod.")

	if err := a.pods.Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("terminate pod: %w", err)
	}

	log.Info().
		WithStringf("pod", "%s/%s", a.namespace, pod.Name).
		Message("Done with worker.")
	return nil
}

func pipeAndClose(p PipeCloser) error {
	defer p.Close()
	for {
		if err := p.PipeMessage(); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}
