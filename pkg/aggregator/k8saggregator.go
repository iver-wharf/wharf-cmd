package aggregator

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/request"
	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/wharfapi"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerclient"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"gopkg.in/typ.v3/pkg/sync2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8sruntime "k8s.io/apimachinery/pkg/util/runtime"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var log = logger.NewScoped("AGGREGATOR")

const (
	wharfAPIExternalPort  = 5011
	workerAPIExternalPort = 5010
)

// Copied from pkg/provisioner/k8sprovisioner.go
var listOptionsMatchLabels = metav1.ListOptions{
	LabelSelector: "app.kubernetes.io/name=wharf-cmd-worker," +
		"app.kubernetes.io/managed-by=wharf-cmd-provisioner," +
		"wharf.iver.com/instance=prod",
}

// NewK8sAggregator returns a new Aggregator implementation that targets
// Kubernetes using a specific Kubernetes namespace and REST config.
func NewK8sAggregator(namespace string, restConfig *rest.Config) (Aggregator, error) {
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
	return k8sAggregator{
		Namespace:  namespace,
		Clientset:  clientset,
		Pods:       clientset.CoreV1().Pods(namespace),
		restConfig: restConfig,

		upgrader:   upgrader,
		httpClient: httpClient,
		wharfClient: wharfapi.Client{
			APIURL: "http://localhost:5001",
		},
	}, nil
}

type k8sAggregator struct {
	Namespace string
	Clientset *kubernetes.Clientset
	Pods      corev1.PodInterface

	restConfig *rest.Config
	upgrader   spdy.Upgrader
	httpClient *http.Client

	wharfClient wharfapi.Client
}

func (a k8sAggregator) Serve(ctx context.Context) error {
	log.Info().Message("Aggregator started")

	// Silences the output of error messages from internal k8s code to console.
	//
	// The console was clogged with forwarding errors when attempting to ping
	// a worker while its server wasn't running.
	k8sruntime.ErrorHandlers = []func(error){}

	inProgress := sync2.Map[types.UID, bool]{}
	for {
		// TODO: Wait for Wharf API to be up first, with sane infinite retry logic.
		//
		// Would prevent pod listing and opening a tunnel to each pod each
		// iteration.

		podList, err := a.listMatchingPods(ctx)
		if err != nil {
			continue
		}
		for _, pod := range podList.Items {
			if pod.Status.Phase != v1.PodRunning {
				continue
			}
			if _, ok := inProgress.Load(pod.UID); ok {
				continue
			}

			log.Debug().WithString("pod", pod.Name).Message("Pod found.")
			go func(p v1.Pod) {
				inProgress.Store(p.UID, true)
				if err := a.relayToWharfDB(ctx, &p); err != nil {
					log.Error().WithError(err).Message("Relay error.")
				}
				inProgress.Delete(p.UID)
			}(pod)
		}
		time.Sleep(5 * time.Second)
	}
}

func (a k8sAggregator) listMatchingPods(ctx context.Context) (*v1.PodList, error) {
	return a.Pods.List(ctx, listOptionsMatchLabels)
}

func (a k8sAggregator) relayToWharfDB(ctx context.Context, pod *v1.Pod) error {
	port, connCloser, err := a.establishTunnel(pod.Namespace, pod.Name)
	if err != nil {
		return err
	}
	defer connCloser.Close()

	log.Info().WithString("pod", pod.Name).
		WithUint("local", uint(port.Local)).
		WithUint("remote", uint(port.Remote)).
		Message("Connected to worker. Port-forwarding from pod.")

	client, err := workerclient.New(fmt.Sprintf("127.0.0.1:%d", port.Local), workerclient.Options{
		// Skipping security because we've already authenticated with Kubernetes
		// and are communicating through a secured port-forwarding tunnel.
		// Don't need to add TLS on top of TLS.
		InsecureSkipVerify: true,
	})
	defer client.Close()

	if err := client.Ping(ctx); err != nil {
		return err
	}

	var cg cancelGroup
	cg.add(func(ctx context.Context) error {
		return a.relayLogs(ctx, client)
	})
	cg.add(func(ctx context.Context) error {
		return a.relayArtifactEvents(ctx, client)
	})
	cg.add(func(ctx context.Context) error {
		return a.relayStatusEvents(ctx, client)
	})

	if err := cg.runInParallelFailFast(ctx); err != nil {
		// This will not show all the errors, but that's fine.
		return fmt.Errorf("relaying to wharf: %w", err)
	}

	if err := client.Kill(ctx); err != nil {
		log.Error().WithError(err).WithString("pod", pod.Name).
			Message("Failed to kill worker.")
		return err
	}
	log.Info().WithString("pod", pod.Name).
		Message("Done with worker. Killed pod.")

	return nil
}

type closerFunc func()

func (f closerFunc) Close() error {
	f()
	return nil
}

func (a k8sAggregator) establishTunnel(namespace, podName string) (*portforward.ForwardedPort, io.Closer, error) {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
		url.PathEscape(namespace), url.PathEscape(podName))

	portForwardUrl, err := url.Parse(a.restConfig.Host)
	if err != nil {
		return nil, nil, fmt.Errorf("parse URL from kubeconfig: %w", err)
	}
	// rest.Config.Host can look something like one of these:
	//   https://172.50.123.3:6443
	//   https://rancher.example.com/k8s/clusters/c-m-13mz8a32
	//
	// We add the path to that, to produce the correct results:
	//   https://172.50.123.3:6443/api/v1/namespaces/my-ns/pods/my-pod/portforward
	//   https://rancher.example.com/k8s/clusters/c-m-13mz8a32/api/v1/namespaces/my-ns/pods/my-pod/port-forward
	portForwardUrl.Path += path

	dialer := spdy.NewDialer(a.upgrader, a.httpClient, http.MethodGet, portForwardUrl)
	stopCh, readyCh := make(chan struct{}, 1), make(chan struct{}, 1)
	forwarder, err := portforward.New(dialer,
		// From random unused local port (port 0) to the worker HTTP API port.
		[]string{fmt.Sprintf("0:%d", workerAPIExternalPort)},
		stopCh, readyCh, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	var forwarderErr error
	go func() {
		if forwarderErr = forwarder.ForwardPorts(); forwarderErr != nil {
			log.Error().WithError(forwarderErr).Message("Error occurred during tunneling.")
			close(readyCh)
			close(stopCh)
			forwarder.Close()
		}
	}()

	<-readyCh
	if forwarderErr != nil {
		return nil, nil, forwarderErr
	}

	closePortForward := closerFunc(func() {
		close(stopCh)
	})

	ports, err := forwarder.GetPorts()
	if err != nil {
		log.Error().WithError(err).Message("Error getting ports.")
		closePortForward()
		return nil, nil, err
	}

	return &ports[0], closePortForward, nil
}

func (a k8sAggregator) relayLogs(ctx context.Context, client workerclient.Client) error {
	reader, err := client.StreamLogs(ctx, &workerclient.StreamLogsRequest{})
	if err != nil {
		return fmt.Errorf("open logs stream from wharf-cmd-worker: %w", err)
	}
	defer reader.CloseSend()

	writer, err := a.wharfClient.CreateBuildLogStream(ctx)
	if err != nil {
		return fmt.Errorf("open logs stream to wharf-api: %w", err)
	}
	defer writer.CloseAndRecv()

	for {
		logLine, err := reader.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			break
		}
		writer.Send(request.Log{
			BuildID:      uint(logLine.BuildID),
			WorkerLogID:  uint(logLine.LogID),
			WorkerStepID: uint(logLine.StepID),
			Timestamp:    logLine.Timestamp.AsTime(),
			Message:      logLine.Message,
		})
	}
	return nil
}

func (a k8sAggregator) relayArtifactEvents(ctx context.Context, client workerclient.Client) error {
	stream, err := client.StreamArtifactEvents(ctx, &workerclient.ArtifactEventsRequest{})
	if err != nil {
		return fmt.Errorf("open artifact events stream from wharf-cmd-worker: %w", err)
	}
	defer stream.CloseSend()

	for {
		artifactEvent, err := stream.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			break
		}
		// No way to send to wharf DB through stream currently
		// so we're just logging it here.
		log.Debug().
			WithString("name", artifactEvent.Name).
			WithUint64("id", artifactEvent.ArtifactID).
			Message("Received artifact event.")
	}
	return nil
}

func (a k8sAggregator) relayStatusEvents(ctx context.Context, client workerclient.Client) error {
	stream, err := client.StreamStatusEvents(ctx, &workerclient.StatusEventsRequest{})
	if err != nil {
		return fmt.Errorf("open status events stream from wharf-cmd-worker: %w", err)
	}
	defer stream.CloseSend()

	// TODO: Update build status based on statuses
	for {
		statusEvent, err := stream.Recv()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			break
		}
		log.Debug().
			WithStringer("status", statusEvent.Status).
			Message("Received status event.")
	}
	return nil
}
