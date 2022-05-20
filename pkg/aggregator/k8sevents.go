// Copyright 2014 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aggregator

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
)

// Modified copy from:
//   https://github.com/kubernetes/kubernetes/blob/b6a0718858876bbf8cedaeeb47e6de7e650a6c5b/pkg/kubectl/describe/versioned/describe.go#L3247
func describeEvents(el *v1.EventList) io.ReadCloser {
	var b bytes.Buffer
	if len(el.Items) == 0 {
		b.WriteString("<none>\n")
		return io.NopCloser(&b)
	}

	tw := tabwriter.NewWriter(&b, 1, 4, 2, ' ', 0)
	fmt.Fprintln(tw)
	fmt.Fprintln(tw, "Type\tReason\tAge\tFrom\tMessage")
	fmt.Fprintln(tw, "----\t------\t---\t----\t-------")
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

	return io.NopCloser(&b)
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
