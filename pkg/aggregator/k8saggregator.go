package aggregator

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

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
	// TODO: Get these ports from params
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
	return k8sAggr{
		namespace:  namespace,
		clientset:  clientset,
		pods:       clientset.CoreV1().Pods(namespace),
		restConfig: restConfig,

		upgrader:   upgrader,
		httpClient: httpClient,
		wharfapi: wharfapi.Client{
			// TODO: Get from params
			APIURL: "http://localhost:5001",
		},
	}, nil
}

type k8sAggr struct {
	namespace string
	clientset *kubernetes.Clientset
	pods      corev1.PodInterface

	restConfig *rest.Config
	upgrader   spdy.Upgrader
	httpClient *http.Client

	wharfapi wharfapi.Client
}

func (a k8sAggr) Serve(ctx context.Context) error {
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

func (a k8sAggr) listMatchingPods(ctx context.Context) (*v1.PodList, error) {
	return a.pods.List(ctx, listOptionsMatchLabels)
}

func (a k8sAggr) relayToWharfDB(ctx context.Context, pod *v1.Pod) error {
	portConn, err := a.establishTunnel(pod.Namespace, pod.Name)
	if err != nil {
		return err
	}
	defer portConn.Close()

	log.Info().WithString("pod", pod.Name).
		WithUint("local", uint(portConn.Local)).
		WithUint("remote", uint(portConn.Remote)).
		Message("Connected to worker. Port-forwarding from pod.")

	worker, err := workerclient.New(fmt.Sprintf("127.0.0.1:%d", portConn.Local), workerclient.Options{
		// Skipping security because we've already authenticated with Kubernetes
		// and are communicating through a secured port-forwarding tunnel.
		// Don't need to add TLS on top of TLS.
		InsecureSkipVerify: true,
	})
	if err != nil {
		return err
	}
	defer worker.Close()

	if err := worker.Ping(ctx); err != nil {
		return err
	}

	if err := relayAll(ctx, a.wharfapi, worker); err != nil {
		// This will not show all the errors, but that's fine.
		return fmt.Errorf("relaying to wharf: %w", err)
	}

	// TODO: Check build results from already-streamed status events if the
	// build is actually done. If not, then handle that as an error

	if err := worker.Kill(ctx); err != nil {
		log.Error().WithError(err).WithString("pod", pod.Name).
			Message("Failed to kill worker.")
		return err
	}
	log.Info().WithString("pod", pod.Name).
		Message("Done with worker. Killed pod.")

	return nil
}

type portConnection struct {
	portforward.ForwardedPort
	stopCh chan struct{}
}

func (pc portConnection) Close() error {
	close(pc.stopCh)
	return nil
}

func (a k8sAggr) establishTunnel(namespace, podName string) (portConnection, error) {
	portForwardURL, err := newPortForwardURL(a.restConfig.Host, namespace, podName)
	if err != nil {
		return portConnection{}, err
	}

	dialer := spdy.NewDialer(a.upgrader, a.httpClient, http.MethodGet, portForwardURL)
	stopCh, readyCh := make(chan struct{}, 1), make(chan struct{}, 1)
	forwarder, err := portforward.New(dialer,
		// From random unused local port (port 0) to the worker HTTP API port.
		[]string{fmt.Sprintf("0:%d", workerAPIExternalPort)},
		stopCh, readyCh, nil, nil)
	if err != nil {
		return portConnection{}, err
	}

	var forwarderErr error
	go func() {
		if forwarderErr = forwarder.ForwardPorts(); forwarderErr != nil {
			log.Error().WithError(forwarderErr).Message("Error occurred when forwarding ports.")
			close(stopCh)
		}
	}()

	select {
	case <-readyCh:
	case <-stopCh:
	}
	if forwarderErr != nil {
		return portConnection{}, forwarderErr
	}

	ports, err := forwarder.GetPorts()
	if err != nil {
		log.Error().WithError(err).Message("Error getting ports.")
		close(stopCh)
		return portConnection{}, err
	}

	return portConnection{
		ForwardedPort: ports[0],
		stopCh:        stopCh,
	}, nil
}

func newPortForwardURL(apiURL, namespace, podName string) (*url.URL, error) {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
		url.PathEscape(namespace), url.PathEscape(podName))

	portForwardURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("parse URL from kubeconfig: %w", err)
	}
	portForwardURL.Path += path
	return portForwardURL, nil
}
