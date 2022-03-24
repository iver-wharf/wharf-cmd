package aggregator

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/request"
	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/response"
	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/wharfapi"
	"github.com/iver-wharf/wharf-cmd/pkg/aggregator/relayer"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerclient"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/net"
	k8sRuntime "k8s.io/apimachinery/pkg/util/runtime"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var log = logger.NewScoped("AGGREGATOR")

const (
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

func (a k8sAggregator) Serve() error {
	log.Info().Message("Aggregator started")

	// Silences the output of error messages from internal k8s code to console.
	//
	// The console was clogged with forwarding errors when attempting to ping
	// a worker while its server wasn't running.
	k8sRuntime.ErrorHandlers = []func(error){}

	inProgress := sync.Map{}
	for {
		// TODO: Healthcheck for Wharf API, back off if down.

		time.Sleep(time.Second)
		podList, err := a.listMatchingPods(context.Background())
		if err != nil {
			log.Error().WithError(err).Message("error listing pods")
			continue
		}
		for _, pod := range podList.Items {
			if pod.Status.Phase != v1.PodRunning {
				continue
			}

			if _, ok := inProgress.Load(string(pod.UID)); ok {
				continue
			}

			log.Debug().WithString("podName", pod.Name).Message("Pod found")
			go func(p v1.Pod) {
				inProgress.Store(string(p.UID), true)
				a.streamToWharfDB(&p)
				inProgress.Delete(string(p.UID))
			}(pod)
		}
	}
}

func (a k8sAggregator) listMatchingPods(ctx context.Context) (*v1.PodList, error) {
	return a.Pods.List(ctx, listOptionsMatchLabels)
}

type connectionCloser func()

func (a k8sAggregator) establishTunnel(namespace, podName string) (*portforward.ForwardedPort, connectionCloser, error) {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
		namespace, podName)

	log.Debug().WithString("path", path)

	hostBase := strings.TrimLeft(a.restConfig.Host, "htps:/")
	hostSplit := strings.Split(hostBase, ":")
	hostIP := hostSplit[0]
	hostPort, err := strconv.Atoi(hostSplit[1])

	if err != nil {
		return nil, nil, err
	}

	url := net.FormatURL("https", hostIP, hostPort, path)
	dialer := spdy.NewDialer(a.upgrader, a.httpClient, http.MethodGet, url)
	stopCh, readyCh := make(chan struct{}, 1), make(chan struct{}, 1)
	forwarder, err := portforward.New(dialer,
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

	for range readyCh {
		if forwarderErr != nil {
			return nil, nil, forwarderErr
		}
	}
	if forwarderErr != nil {
		return nil, nil, forwarderErr
	}

	closerFunc := func() {
		close(stopCh)
	}

	ports, err := forwarder.GetPorts()
	if err != nil {
		log.Error().WithError(err).Message("Error getting ports.")
		closerFunc()
		return nil, nil, err
	}

	return &ports[0], closerFunc, nil
}

func (a k8sAggregator) streamToWharfDB(pod *v1.Pod) error {
	port, connCloser, err := a.establishTunnel(pod.Namespace, pod.Name)
	if err != nil {
		return err
	}

	log.Debug().WithString("name", pod.Name).
		WithUint("local", uint(port.Local)).
		WithUint("remote", uint(port.Remote)).
		Message("Tunnel opened to worker.")

	client, err := workerclient.New(fmt.Sprintf("127.0.0.1:%d", port.Local), workerclient.Options{
		InsecureSkipVerify: true,
	})
	defer func() {
		client.Close()
		connCloser()
	}()

	if err := client.Ping(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	var errs []string
	wg.Add(1)
	go func() {
		defer wg.Done()
		sender, err := relayer.AsSender[request.Log, response.CreatedLogsSummary](
			func() (any, error) {
				return a.wharfClient.CreateBuildLogStream(context.Background())
			})
		if err != nil {
			errs = append(errs, err.Error())
			return
		}

		receiver, err := relayer.AsReceiver[*workerclient.LogLine](
			func() (any, error) {
				return client.StreamLogs(context.Background(), &workerclient.LogsRequest{})
			})
		if err != nil {
			errs = append(errs, err.Error())
			return
		}

		relay := relayer.New(receiver, sender, func(v *workerclient.LogLine) request.Log {
			return request.Log{
				BuildID:      uint(v.BuildID),
				WorkerLogID:  uint(v.LogID),
				WorkerStepID: uint(v.StepID),
				Timestamp:    v.GetTimestamp().AsTime(),
				Message:      v.GetMessage(),
			}
		})

		if errs2 := relay.Relay(); errs2 != nil {
			errs = append(errs, errs2...)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		stream, err := client.StreamArtifactEvents(context.Background(), &workerclient.ArtifactEventsRequest{})
		if err != nil {
			errs = append(errs, err.Error())
			return
		}
		// No way to send to wharf DB through stream currently (that I know of)
		// so we're just printing here.
		for {
			artifactEvent, err := stream.Recv()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					errs = append(errs, err.Error())
				}
				break
			}
			log.Debug().WithStringer("value", artifactEvent).Message("Received artifact event.")
		}
		stream.CloseSend()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		stream, err := client.StreamStatusEvents(context.Background(), &workerclient.StatusEventsRequest{})
		if err != nil {
			errs = append(errs, err.Error())
			return
		}
		// No way to send to wharf DB through stream currently (that I know of)
		// so we're just printing here.
		for {
			statusEvent, err := stream.Recv()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					errs = append(errs, err.Error())
				}
				break
			}
			log.Debug().WithStringer("value", statusEvent).Message("Received status event.")
		}
		stream.CloseSend()
	}()
	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("error relaying to wharf: %s", strings.Join(errs, "; "))
	}

	if err := client.Kill(); err != nil {
		log.Error().WithError(err).Message("Failed killing worker.")
	}
	log.Info().WithString("name", pod.Name).
		WithString("id", string(pod.UID)).
		Message("Killed worker")

	return nil
}
