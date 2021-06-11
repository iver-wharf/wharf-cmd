package kubernetes

import (
	v1 "k8s.io/api/core/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func NewContainerDoneWaiter(podInterface corev1.PodInterface, podName string) (ContainerWaiter, error) {
	return newContainerWaiter(podInterface, podName, newDoneStateWatcher())
}

type containerDoneStateWatcher struct{}

func newDoneStateWatcher() ContainerStateWatcher {
	return &containerDoneStateWatcher{}
}

func (cdsw containerDoneStateWatcher) ContainerIsInExpectedStatus(container AwaitedContainer) bool {
	return container.State == ContainerTerminated
}

func (cdsw containerDoneStateWatcher) HasContainerStatusChanged(container AwaitedContainer, status v1.ContainerStatus) bool {
	return container.State != mapToContainerState(status.State)
}
