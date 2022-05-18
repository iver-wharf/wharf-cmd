package aggregator

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
)

// Modified copy from:
//   https://github.com/kubernetes/kubernetes/blob/b6a0718858876bbf8cedaeeb47e6de7e650a6c5b/pkg/kubectl/describe/versioned/describe.go#L3247
func describeEvents(el *v1.EventList) string {
	if len(el.Items) == 0 {
		return "<none>"
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
	return sb.String()
}

// translateTimestampSince returns the elapsed time since timestamp in
// human-readable approximation.
//
// Verbatim copy from:
//   https://github.com/kubernetes/kubernetes/blob/b6a0718858876bbf8cedaeeb47e6de7e650a6c5b/pkg/kubectl/describe/versioned/describe.go#L4299
func translateTimestampSince(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}

// formatEventSource formats EventSource as a comma separated string excluding Host when empty
//
// Verbatim copy from:
//   https://github.com/kubernetes/kubernetes/blob/b6a0718858876bbf8cedaeeb47e6de7e650a6c5b/pkg/kubectl/describe/versioned/describe.go#L4308
func formatEventSource(es v1.EventSource) string {
	EventSourceString := []string{es.Component}
	if len(es.Host) > 0 {
		EventSourceString = append(EventSourceString, es.Host)
	}
	return strings.Join(EventSourceString, ", ")
}
