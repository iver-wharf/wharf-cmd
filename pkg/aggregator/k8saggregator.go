package aggregator

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/request"
	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/wharfapi"
	"github.com/iver-wharf/wharf-cmd/pkg/config"
	"github.com/iver-wharf/wharf-cmd/pkg/workerapi/workerclient"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"github.com/iver-wharf/wharf-core/pkg/problem"
	"gopkg.in/typ.v4/sync2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/duration"
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
			buildID, err := parsePodBuildID(pod.ObjectMeta)
			if err != nil {
				log.Warn().WithError(err).
					WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
					Message("Failed to parse worker's build ID.")
				continue
			}
			if !a.inProgress.Add(pod.UID) {
				// Failed to add => Already being processed
				continue
			}

			log.Debug().
				WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
				WithString("status", string(pod.Status.Phase)).
				Message("Pod found.")
			switch pod.Status.Phase {
			case v1.PodRunning:
				go func(pod v1.Pod) {
					if err := a.relayToWharfAPI(ctx, pod.Name, buildID); err != nil {
						log.Error().WithError(err).
							WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
							Message("Relay error.")
					}
					a.terminatePod(ctx, pod)
				}(pod)
			case v1.PodFailed:
				go func(pod v1.Pod) {
					if err := a.relayRawLogs(ctx, pod, buildID); err != nil {
						log.Error().WithError(err).
							WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
							Message("Failed relaying logs for failed pod.")
					}
					a.terminatePod(ctx, pod)
				}(pod)
			default:
				go func(pod v1.Pod) {
					if err := a.relayRawLogs(ctx, pod, buildID); err != nil {
						log.Error().WithError(err).
							WithStringf("pod", "%s/%s", pod.Namespace, pod.Name).
							Message("Failed relaying events for failed pod.")
					}
					a.relayEvents(ctx, pod, buildID)
					a.terminatePod(ctx, pod)
				}(pod)
			}
		}
		time.Sleep(pollDelay)
	}
}

func (a k8sAggr) relayRawLogs(ctx context.Context, pod v1.Pod, buildID uint) error {
	logs := make(chan string)
	go func() {
		logs <- fmt.Sprintf("The %s pod failed to start.", pod.Name)
		logs <- fmt.Sprintf("Logs from %s:\n", pod.Name)
		if err := a.appendRawLogsFromPodToChannel(ctx, pod.Name, logs); err != nil {
			logs <- fmt.Sprintf("Failed reading logs: %v", err)
		}
		close(logs)
	}()
	defer func() {
		if _, err := a.wharfapi.UpdateBuildStatus(buildID, request.LogOrStatusUpdate{
			Message:   "Build failed",
			Timestamp: time.Now(),
			Status:    request.BuildFailed,
		}); err != nil {
			var prob problem.Response
			if errors.As(err, &prob) {
				log.Error().WithFunc(logFuncFromProb(prob)).Message("Failed updating build status after logs streaming.")
			} else {
				log.Error().WithError(err).Message("Failed updating build status after logs streaming.")
			}
		}
	}()
	return a.relayLogsFromChannelToWharfAPI(ctx, pod.Name, logs, buildID)
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

		if pod.Status.Phase == v1.PodPending && podNotErrored(pod) {
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
	return nil
}

func (a k8sAggr) relayLogsFromChannelToWharfAPI(ctx context.Context, podName string, logs <-chan string, buildID uint) error {
	var sentLogs uint
	writer, err := a.wharfapi.CreateBuildLogStream(ctx)
	if err != nil {
		return fmt.Errorf("open logs stream to wharf-api: %w", err)
	}
	defer func() {
		resp, err := writer.CloseAndRecv()
		if err != nil {
			log.Warn().
				WithError(err).
				Message("Unexpected error when closing log writer stream to wharf-api.")
			return
		}
		log.Debug().
			WithUint("sent", sentLogs).
			WithUint("inserted", resp.LogsInserted).
			Message("Inserted logs into wharf-api.")
	}()
	for logLine := range logs {
		if err := writer.Send(request.Log{
			BuildID:   buildID,
			Timestamp: time.Now(),
			Message:   logLine,
		}); err != nil {
			log.Error().WithError(err).Message("Sending logs to Wharf API failed.")
		}
		sentLogs++
	}
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

func (a k8sAggr) appendRawLogsFromPodToChannel(ctx context.Context, podName string, ch chan<- string) error {
	req := a.pods.GetLogs(podName, &v1.PodLogOptions{})
	readCloser, err := req.Stream(ctx)
	if err != nil {
		return err
	}
	defer readCloser.Close()
	scanner := bufio.NewScanner(readCloser)
	for scanner.Scan() {
		txt := scanner.Text()
		log.Debug().Message(txt)
		idx := strings.LastIndexByte(txt, '\r')
		if idx != -1 {
			txt = txt[idx+1:]
		}
		ch <- txt
	}
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

func (a k8sAggr) newPortForwarding(namespace, podName string) (portConnection, error) {
	portForwardURL, err := newPortForwardURL(a.restConfig.Host, namespace, podName)
	if err != nil {
		return portConnection{}, err
	}

	dialer := spdy.NewDialer(a.upgrader, a.httpClient, http.MethodGet, portForwardURL)
	stopCh, readyCh := make(chan struct{}, 1), make(chan struct{}, 1)
	forwarder, err := portforward.New(dialer,
		// From random unused local port (port 0) to the worker HTTP API port.
		[]string{fmt.Sprintf("0:%d", a.aggrConfig.WorkerAPIExternalPort)},
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

// Modified version of DescribeEvents found at:
//   https://github.com/kubernetes/kubernetes/blob/b6a0718858876bbf8cedaeeb47e6de7e650a6c5b/pkg/kubectl/describe/versioned/describe.go#L3247
func describeEvents(el *v1.EventList) []string {
	if len(el.Items) == 0 {
		return []string{"<none>"}
	}

	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 1, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "\nType\tReason\tAge\tFrom\tMessage")
	fmt.Fprintln(tw, "----\t------\t----\t----\t-------")
	for _, e := range el.Items {
		var interval string
		if e.Count > 1 {
			interval = fmt.Sprintf("%s (x%d over %s)", translateTimestampSince(e.LastTimestamp), e.Count, translateTimestampSince(e.FirstTimestamp))
		} else {
			interval = translateTimestampSince(e.FirstTimestamp)
		}
		fmt.Fprintf(tw, "%v\t%v\t%s\t%v\t%v\n",
			e.Type,
			e.Reason,
			interval,
			formatEventSource(e.Source),
			strings.TrimSpace(e.Message),
		)
	}
	tw.Flush()
	return strings.Split(sb.String(), "\n")
}

// translateTimestampSince returns the elapsed time since timestamp in
// human-readable approximation.
//
// Copied from:
//   https://github.com/kubernetes/kubernetes/blob/b6a0718858876bbf8cedaeeb47e6de7e650a6c5b/pkg/kubectl/describe/versioned/describe.go#L4299
func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}

// formatEventSource formats EventSource as a comma separated string excluding Host when empty
//
// Copied from:
//   https://github.com/kubernetes/kubernetes/blob/b6a0718858876bbf8cedaeeb47e6de7e650a6c5b/pkg/kubectl/describe/versioned/describe.go#L4308
func formatEventSource(es v1.EventSource) string {
	EventSourceString := []string{es.Component}
	if len(es.Host) > 0 {
		EventSourceString = append(EventSourceString, es.Host)
	}
	return strings.Join(EventSourceString, ", ")
}

func (a k8sAggr) appendEventsFromPodToChannel(ctx context.Context, pod *v1.Pod, ch chan<- string) error {
	eventsList, err := a.clientset.CoreV1().Events(a.namespace).Search(scheme.Scheme, pod)
	if err != nil {
		return err
	}

	lines := describeEvents(eventsList)
	for _, line := range lines {
		log.Debug().Message(line)
		idx := strings.LastIndexByte(line, '\r')
		if idx != -1 {
			line = line[idx+1:]
		}
		ch <- line
	}

	return nil
}

func (a k8sAggr) relayEvents(ctx context.Context, pod v1.Pod, buildID uint) error {
	logs := make(chan string)
	go func() {
		logs <- fmt.Sprintf("The %s pod failed to start. Kubernetes events:", pod.Name)
		if err := a.appendEventsFromPodToChannel(ctx, &pod, logs); err != nil {
			logs <- fmt.Sprintf("Failed reading events: %v", err)
		}
		logs <- fmt.Sprintf("No logs from %s pod.", pod.Name)
		close(logs)
	}()
	defer func() {
		if _, err := a.wharfapi.UpdateBuildStatus(buildID, request.LogOrStatusUpdate{
			Message:   "Build failed",
			Timestamp: time.Now(),
			Status:    request.BuildFailed,
		}); err != nil {
			var prob problem.Response
			if errors.As(err, &prob) {
				log.Error().WithFunc(logFuncFromProb(prob)).Message("Failed updating build status after event logs streaming.")
			} else {
				log.Error().WithError(err).Message("Failed updating build status after event logs streaming.")
			}
		}
	}()
	return a.relayLogsFromChannelToWharfAPI(ctx, pod.Name, logs, buildID)
}

func logFuncFromProb(prob problem.Response) func(ev logger.Event) logger.Event {
	return func(ev logger.Event) logger.Event {
		return ev.
			WithInt("httpStatus", prob.Status).
			WithString("docs", prob.Type).
			WithString("title", prob.Title).
			WithString("detail", prob.Detail).
			WithString("instance", prob.Instance).
			WithString("error", prob.Error())
	}
}

func podNotErrored(pod v1.Pod) bool {
	for _, s := range pod.Status.InitContainerStatuses {
		if s.State.Waiting != nil {
			switch s.State.Waiting.Reason {
			case "CrashLoopBackOff", "ErrImagePull",
				"ImagePullBackOff", "CreateContainerConfigError",
				"InvalidImageName", "CreateContainerError":
				return false
			}
		}
	}
	for _, s := range pod.Status.ContainerStatuses {
		if s.State.Waiting != nil {
			switch s.State.Waiting.Reason {
			case "CrashLoopBackOff", "ErrImagePull",
				"ImagePullBackOff", "CreateContainerConfigError",
				"InvalidImageName", "CreateContainerError":
				return false
			}
		}
	}
	return true
}
