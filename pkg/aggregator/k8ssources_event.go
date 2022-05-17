package aggregator

import (
	"context"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/iver-wharf/wharf-api-client-go/v2/pkg/model/request"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type k8sEventsSource struct {
	ctx     context.Context
	buildID uint
	pod     v1.Pod
	events  corev1.EventInterface
}

func (s k8sEventsSource) pushInto(dst chan<- request.Log) error {
	dst <- s.stringToLog(fmt.Sprintf("The %s pod failed to start. Kubernetes events:", s.pod.Name))
	eventsList, err := s.events.Search(scheme.Scheme, &s.pod)
	if err != nil {
		dst <- s.stringToLog(fmt.Sprintf("Failed reading events: %v", err))
		return err
	}
	lines := describeEvents(eventsList)
	for _, line := range lines {
		log.Debug().Message(line)
		idx := strings.LastIndexByte(line, '\r')
		if idx != -1 {
			line = line[idx+1:]
		}
		dst <- s.stringToLog(line)
	}
	dst <- s.stringToLog(fmt.Sprintf("No logs from %s pod.", s.pod.Name))
	return nil
}

func (s k8sEventsSource) stringToLog(str string) request.Log {
	return request.Log{
		BuildID: s.buildID,
		// WorkerLogID:  0,
		// WorkerStepID: 0,
		Timestamp: time.Now(),
		Message:   str,
	}
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
