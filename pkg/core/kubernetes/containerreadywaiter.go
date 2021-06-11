package kubernetes

import (
	v1 "k8s.io/api/core/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func NewContainerReadyWaiter(podInterface corev1.PodInterface, podName string) (ContainerWaiter, error) {
	return newContainerWaiter(podInterface, podName, newContainerReadyStateWatcher())
}

type containerReadyStateWatcher struct{}

func newContainerReadyStateWatcher() ContainerStateWatcher {
	return &containerReadyStateWatcher{}
}

func (cdsw containerReadyStateWatcher) ContainerIsInExpectedStatus(container AwaitedContainer) bool {
	return container.isReady
}

func (cdsw containerReadyStateWatcher) HasContainerStatusChanged(container AwaitedContainer, status v1.ContainerStatus) bool {
	return container.isReady != status.Ready
}
