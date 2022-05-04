package aggregator

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/wharfapi"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerclient"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"gopkg.in/typ.v4/sync2"
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
	wharfAPIExternalPort  = 5001
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
	const pollDelay = 5 * time.Second
	log.Info().
		WithDuration("pollDelay", pollDelay).
		Message("Aggregator started.")

	// Silences the output of error messages from internal k8s code to console.
	//
	// The console gets clogged with forwarding errors when attempting to ping
	// a worker while its server wasn't running.
	k8sruntime.ErrorHandlers = []func(error){}

	var inProgress sync2.Set[types.UID]
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
			if pod.Status.Phase != v1.PodRunning {
				continue
			}
			buildID, err := parsePodBuildID(pod.ObjectMeta)
			if err != nil {
				log.Warn().WithError(err).
					WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
					Message("Failed to parse worker's build ID.")
				continue
			}
			if !inProgress.Add(pod.UID) {
				// Failed to add => Already beeing processed
				continue
			}
			log.Debug().
				WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
				Message("Pod found.")

			go func(pod v1.Pod) {
				if err := a.relayToWharfAPI(ctx, pod.Name, buildID); err != nil {
					log.Error().WithError(err).
						WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
						Message("Relay error.")
				}
				inProgress.Remove(pod.UID)
			}(pod)
		}
		time.Sleep(pollDelay)
	}
}

func parsePodBuildID(podMeta metav1.ObjectMeta) (uint, error) {
	buildRef, ok := podMeta.Labels["wharf.iver.com/build-ref"]
	if !ok {
		return 0, errors.New("missing label 'wharf.iver.com/build-ref'")
	}
	buildID, err := strconv.ParseUint(buildRef, 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(buildID), nil
}

func (a k8sAggr) fetchPods(ctx context.Context) ([]v1.Pod, error) {
	list, err := a.pods.List(ctx, listOptionsMatchLabels)
	if err != nil {
		return nil, err
	}
	var pods []v1.Pod
	for _, pod := range list.Items {
		// Skip terminating pods
		if pod.ObjectMeta.DeletionTimestamp != nil {
			continue
		}
		// Skip failed or pending pods
		if pod.Status.Phase != "Running" {
			continue
		}
		pods = append(pods, pod)
	}
	return pods, nil
}

func (a k8sAggr) relayToWharfAPI(ctx context.Context, podName string, buildID uint) error {
	portConn, err := a.newPortForwarding(a.namespace, podName)
	if err != nil {
		return err
	}
	defer portConn.Close()

	worker, err := a.newWorkerClient(portConn, buildID)
	if err != nil {
		return err
	}
	defer worker.Close()

	if err := worker.Ping(ctx); err != nil {
		log.Debug().
			WithStringf("pod", "%s/%s", a.namespace, podName).
			Message("Failed to ping worker pod. Assuming it's not running yet. Skipping.")
		return nil
	}

	if err := relayAll(ctx, a.wharfapi, worker); err != nil {
		// This will not show all the errors, but that's fine.
		return fmt.Errorf("relaying to wharf: %w", err)
	}

	log.Debug().
		WithStringf("pod", "%s/%s", a.namespace, podName).
		Message("Done relaying. Terminating pod.")

	if err := a.pods.Delete(ctx, podName, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("terminate pod after done with relay build results: %w", err)
	}

	log.Info().
		WithStringf("pod", "%s/%s", a.namespace, podName).
		Message("Done with worker.")
	return nil
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

type portConnection struct {
	portforward.ForwardedPort
	stopCh chan struct{}
}

func (pc portConnection) Close() error {
	close(pc.stopCh)
	return nil
}

func (a k8sAggr) newPortForwarding(namespace, podName string) (portConnection, error) {
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
	port := ports[0]

	log.Debug().
		WithStringf("pod", "%s/%s", a.namespace, podName).
		WithUint("local", uint(port.Local)).
		WithUint("remote", uint(port.Remote)).
		Message("Connected to worker. Port-forwarding from pod.")

	return portConnection{
		ForwardedPort: port,
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
