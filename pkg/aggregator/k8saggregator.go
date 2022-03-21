package aggregator

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/request"
	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/response"
	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/wharfapi"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerclient"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	workerv1 "github.com/iver-wharf/wharf-cmd/api/workerapi/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/net"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var log = logger.NewScoped("AGGREGATOR")

const (
	wharfAPIExternalPort  = 5001
	workerAPIExternalPort = 5010
)

// Copied from pkg/provisioner/k8sprovisioner.go
var listOptionsMatchLabels = metav1.ListOptions{
	// LabelSelector: "app.kubernetes.io/name=wharf-cmd-worker," +
	// 	"app.kubernetes.io/managed-by=wharf-cmd-provisioner," +
	// 	"wharf.iver.com/instance=prod",
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
		// wharfClient: wharfapi.Client{
		// 	APIURL: "http://localhost:5001",
		// },
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
	for {
		time.Sleep(time.Second)
		podList, err := a.listMatchingPods(context.Background())
		if err != nil {
			log.Error().WithError(err).Message("listing")
			continue
		}
		for _, pod := range podList.Items {
			if err := a.streamToWharfDB(&pod); err != nil {
				log.Error().WithError(err).Message("streaming error")
				continue
			}
		}
	}
}

func (a k8sAggregator) listMatchingPods(ctx context.Context) (*v1.PodList, error) {
	return a.Pods.List(ctx, listOptionsMatchLabels)
}

type connectionCloser func()

func (a k8sAggregator) connect(namespace, podName string) (*portforward.ForwardedPort, connectionCloser, error) {
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
	out, errOut := new(bytes.Buffer), new(bytes.Buffer)
	forwarder, err := portforward.New(dialer,
		[]string{fmt.Sprintf("0:%d", workerAPIExternalPort)},
		stopCh, readyCh, out, errOut)
	if err != nil {
		return nil, nil, err
	}

	go func() {
		if err := forwarder.ForwardPorts(); err != nil {
			log.Error().WithError(err).Message("Error occurred during forwarding.")
		}
	}()

	for range readyCh {
	}
	closerFunc := func() {
		close(stopCh)
	}

	ports, err := forwarder.GetPorts()
	if err != nil {
		closerFunc()
		return nil, nil, err
	}

	return &ports[0], closerFunc, nil
}

func (a k8sAggregator) streamToWharfDB(pod *v1.Pod) error {
	port, connCloser, err := a.connect(pod.Namespace, pod.Name)
	if err != nil {
		return err
	}
	client, err := workerclient.New(fmt.Sprintf("127.0.0.1:%d", port.Local), workerclient.Options{
		InsecureSkipVerify: true,
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		outStream, err := a.wharfClient.CreateBuildLogStream(context.Background())
		stream, err := client.StreamLogs(context.Background(), &workerclient.LogsRequest{})
		if err != nil {
			log.Error().WithError(err).Message("Fetching stream failed.")
			wg.Done()
			return
		}
		relayToWharf(&wg, relayer[*workerclient.LogLine, response.CreatedLogsSummary, request.Log]{
			streamReceiver: stream,
			streamSender:   outStream,
			convert: func(v *workerclient.LogLine) request.Log {
				return request.Log{
					BuildID:      uint(v.BuildID),
					WorkerLogID:  uint(v.LogID),
					WorkerStepID: uint(v.StepID),
					Timestamp:    v.GetTimestamp().AsTime(),
					Message:      v.GetMessage(),
				}
			},
		})
	}()
	wg.Add(1)
	go func() {
		stream, err := client.StreamArtifactEvents(context.Background(), &workerclient.ArtifactEventsRequest{})
		if err != nil {
			log.Error().WithError(err).Message("Fetching stream failed.")
			wg.Done()
			return
		}
		// No way to send to wharf DB through stream currently (that I know of), so sender is nil
		relayToWharf(&wg, relayer[*workerclient.ArtifactEvent, nillable, nillable]{
			streamReceiver: stream,
		})
	}()
	wg.Add(1)
	go func() {
		stream, err := client.StreamStatusEvents(context.Background(), &workerclient.StatusEventsRequest{})
		if err != nil {
			log.Error().WithError(err).Message("Fetching stream failed.")
			wg.Done()
			return
		}
		// No way to send to wharf DB through stream currently (that I know of), so sender is nil
		relayToWharf(&wg, relayer[*workerclient.StatusEvent, nillable, nillable]{
			streamReceiver: stream,
		})
	}()
	wg.Wait()

	client.Close()
	connCloser()

	return nil
}

type nillable *int

type workerResponse interface {
	*workerv1.StreamStatusEventsResponse | *workerv1.StreamArtifactEventsResponse | *workerv1.StreamLogsResponse
	// used only for debug logging at the moment
	String() string
}

type wharfRequest interface {
	// nillable added to support nil for now
	request.Log | request.BuildStatusUpdate | nillable
}

type wharfResponse interface {
	// nillable added to support nil for now
	response.CreatedLogsSummary | nillable
}

type streamReceiver[received workerResponse] interface {
	Recv() (received, error)
}

type streamSender[sent wharfRequest, received wharfResponse] interface {
	Send(data sent) error
	CloseAndRecv() (received, error)
}

type relayer[fromWorker workerResponse, fromWharf wharfResponse, toWharf wharfRequest] struct {
	streamReceiver[fromWorker]
	streamSender[toWharf, fromWharf]
	convert func(from fromWorker) toWharf
}

func relayToWharf[T1 workerResponse, T2 wharfResponse, T3 wharfRequest](wg *sync.WaitGroup, relayer relayer[T1, T2, T3]) error {
	line, err := relayer.Recv()
	for err == nil {
		log.Debug().WithStringer("value", line).Message("Sending to Wharf")
		if relayer.streamSender != nil {
			relayer.Send(relayer.convert(line))
		}
		line, err = relayer.Recv()
	}
	if relayer.streamSender != nil {
		summary, err := relayer.CloseAndRecv()
		wg.Done()
		if err != nil {
			log.Error().WithError(err).Message("Close and Recv failed.")
			return err
		}
		log.Debug().WithString("summary", fmt.Sprintf("%v", summary)).Message("Send to Wharf DB finished.")
	}
	return nil
}
