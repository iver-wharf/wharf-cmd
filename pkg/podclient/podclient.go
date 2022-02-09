package podclient

import (
	"context"

	v1 "k8s.io/api/core/v1"
)

// Client contains the method set we expect from things dealing with
// pods.
type Client interface {
	// WaitForPodModifiedFunc waits until the pod fulfills the requirements
	// defined in the passed in function.
	WaitForPodModifiedFunc(ctx context.Context, pod *v1.Pod, f func(p *v1.Pod) (bool, error)) error
	// StreamLogsUntilCompleted prints the logs of a pod until the stream has
	// terminated.
	StreamLogsUntilCompleted(ctx context.Context, podName string) error
}
