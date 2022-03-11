package aggregator

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/iver-wharf/wharf-core/pkg/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/net"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var log = logger.NewScoped("AGGREGATOR")

const (
	wharfAPIExternalPort  = 5001
	workerAPIExternalPort = 27017
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
	}, nil
}

type k8sAggregator struct {
	Namespace string
	Clientset *kubernetes.Clientset
	Pods      corev1.PodInterface

	restConfig *rest.Config
	upgrader   spdy.Upgrader
	httpClient *http.Client
}

func (a k8sAggregator) Serve() error {
	log.Info().Message("Serve started")
	var err error
	var podList *v1.PodList
	podList, err = a.listMatchingPods(context.Background())
	if err != nil {
		return err
	}

	for _, item := range podList.Items {
		log.Info().WithString("name", item.Name).Message("")
	}

	body, err := a.do(http.MethodGet, "/api")
	if err != nil {
		return err
	}

	bytes, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	log.Info().WithString("body", string(bytes)).Message("Serve ended")
	return nil
}

func (a k8sAggregator) listMatchingPods(ctx context.Context) (*v1.PodList, error) {
	return a.Pods.List(ctx, listOptionsMatchLabels)
}

func (a k8sAggregator) do(method, address string) (io.ReadCloser, error) {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
		"default", "wharf-build-container-test-wharfable-w294b")
	log.Info().WithString("path", path).Message("starting request")
	hostBase := strings.TrimLeft(a.restConfig.Host, "htps:/")
	hostSplit := strings.Split(hostBase, ":")
	hostIP := hostSplit[0]
	hostPort, err := strconv.Atoi(hostSplit[1])
	if err != nil {
		return nil, err
	}

	url := net.FormatURL("https", hostIP, hostPort, path)
	log.Info().WithStringer("url", url).Message("starting request")

	dialer := spdy.NewDialer(a.upgrader, a.httpClient, method, url)
	stopCh, readyCh := make(chan struct{}, 1), make(chan struct{}, 1)
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)
	forwarder, err := portforward.New(dialer,
		[]string{fmt.Sprintf("0:%d", 10000)},
		stopCh, readyCh, out, errOut)
	if err != nil {
		return nil, err
	}

	go func() {
		for range readyCh {
		}
		ports, err := forwarder.GetPorts()
		if err == nil {
			log.Info().WithInt("local", int(ports[0].Local)).WithInt("remote", int(ports[0].Remote)).Message("PORTS")

			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/api", ports[0].Local))
			if err != nil {
				log.Error().WithError(err).Message("request error")
				return
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Error().WithError(err).Message("request error")
				return
			}
			log.Info().WithString("body", string(body)).Message("success")
		}

		close(stopCh)
	}()
	if err := forwarder.ForwardPorts(); err != nil {
		log.Info().Message("forward err")
		return nil, err
	}

	if len(errOut.String()) != 0 {
		log.Info().Message("err")
		return nil, errors.New(errOut.String())
	} else if len(out.String()) != 0 {
		log.Info().Message("out")
		log.Info().Message(out.String())
		return io.NopCloser(bufio.NewReader(out)), nil
	}
	log.Info().Message("no response")
	return nil, errors.New("no response")
}

type forwardedConnection interface {
	Do(method, address string) (*http.Response, error)
}

type connection struct {
	forwarder *portforward.PortForwarder
	ports     *portforward.ForwardedPort
	stopCh    chan struct{}
}

func (c *connection) Do(method, address string) (*http.Response, error) {
	return nil, nil
}
