package podclient

import (
	"bufio"
	"context"
	"fmt"

	"github.com/iver-wharf/wharf-core/pkg/logger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// BaseClient is a basic implementation of the Client interface.
type BaseClient struct {
	Namespace string
	// restConfig *rest.Config
	Clientset *kubernetes.Clientset
	Pods      corev1.PodInterface
	// events     corev1.EventInterface
}

// WaitForPodModifiedFunc is used to wait until a pod is modified and the
// passed function evaluates to true when the resulting pod is passed to
// it.
func (c BaseClient) WaitForPodModifiedFunc(ctx context.Context, podMeta metav1.ObjectMeta, f func(p *v1.Pod) (bool, error)) error {
	w, err := c.Clientset.CoreV1().Pods(c.Namespace).Watch(ctx, metav1.SingleObject(podMeta))
	if err != nil {
		return err
	}

	defer w.Stop()
	for ev := range w.ResultChan() {
		pod := ev.Object.(*v1.Pod)
		switch ev.Type {
		case watch.Modified:
			ok, err := f(pod)
			if err != nil {
				return err
			} else if ok {
				return nil
			}
		case watch.Deleted:
			return fmt.Errorf("pod was removed: %v", pod.Name)
		}
	}
	return fmt.Errorf("got no more events when watching pod: %v", podMeta.Name)
}

// StreamLogsUntilCompleted prints the logs of a pod until the stream has
// terminated.
func (c BaseClient) StreamLogsUntilCompleted(ctx context.Context, podName string) error {
	req := c.Pods.GetLogs(podName, &v1.PodLogOptions{
		Follow: true,
	})
	readCloser, err := req.Stream(ctx)
	if err != nil {
		return err
	}
	defer readCloser.Close()
	podLog := logger.NewScoped(podName)
	scanner := bufio.NewScanner(readCloser)
	for scanner.Scan() {
		podLog.Info().Message(scanner.Text())
	}
	return scanner.Err()
}
