package kubernetes

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type ContainerStateWatcher interface {
	ContainerIsInExpectedStatus(container AwaitedContainer) bool
	HasContainerStatusChanged(container AwaitedContainer, status v1.ContainerStatus) bool
}

type ContainerWaiter interface {
	AnyRemaining() bool
	WaitNext() (AwaitedContainer, error)
}

type containerWaiter struct {
	podInterface          corev1.PodInterface
	containers            map[string]AwaitedContainer
	podInfo               podInfo
	containersFetched     bool
	containerStateWatcher ContainerStateWatcher
}

const waitBetweenChecks = 100 * time.Millisecond

func newContainerWaiter(
	podInterface corev1.PodInterface,
	podName string,
	containerStateWatcher ContainerStateWatcher) (ContainerWaiter, error) {
	c := &containerWaiter{
		podInterface: podInterface,
		containers:   map[string]AwaitedContainer{},
		podInfo: podInfo{
			namespace: "",
			podUID:    "",
			name:      podName,
		},
		containersFetched:     false,
		containerStateWatcher: containerStateWatcher,
	}

	pod, err := c.fetchPod()
	if err != nil {
		return nil, err
	}

	c.podInfo = mapToPodInfo(pod)

	return c, nil
}

func (cws *containerWaiter) AnyRemaining() bool {
	if !cws.containersFetched {
		return true
	}

	for _, c := range cws.containers {
		if !cws.containerStateWatcher.ContainerIsInExpectedStatus(c) {
			return true
		}
	}

	return false
}

func (cws *containerWaiter) WaitNext() (AwaitedContainer, error) {
	for cws.AnyRemaining() {
		container, err := cws.fetchNextContainer()
		if err != nil {
			return container, err
		}

		empty := AwaitedContainer{}
		if container != empty {
			return container, nil
		}

		time.Sleep(waitBetweenChecks)
	}

	return AwaitedContainer{}, fmt.Errorf("no more containers to wait")
}

func (cws *containerWaiter) fetchNextContainer() (AwaitedContainer, error) {
	pod, err := cws.fetchPod()
	if err != nil {
		return AwaitedContainer{}, err
	}

	defer func() { cws.containersFetched = true }()

	container, ok := cws.nextContainerInExpectedState(pod.Status.InitContainerStatuses, ContainerTypeInit)
	if ok {
		if !cws.containersFetched {
			_, _ = cws.nextContainerInExpectedState(pod.Status.ContainerStatuses, ContainerTypeApp)
		}
		return container, nil
	}

	container, ok = cws.nextContainerInExpectedState(pod.Status.ContainerStatuses, ContainerTypeApp)
	if ok {
		return container, nil
	}

	return AwaitedContainer{}, nil
}

func (cws *containerWaiter) fetchPod() (*v1.Pod, error) {
	pod, err := cws.podInterface.Get(cws.podInfo.name, metav1.GetOptions{})
	if err != nil {
		log.Error().
			WithString("pod", cws.podInfo.name).
			WithString("namespace", cws.podInfo.namespace).
			WithString("uid", cws.podInfo.podUID).
			Message("Failed to get pod.")
		return nil, err
	}
	return pod, nil
}

func (cws *containerWaiter) nextContainerInExpectedState(containerStatuses []v1.ContainerStatus, containerType ContainerType) (AwaitedContainer, bool) {
	for _, s := range containerStatuses {
		awaitedContainer, ok := cws.containers[s.Name]
		if !ok || cws.containerStateWatcher.HasContainerStatusChanged(awaitedContainer, s) {
			cws.containers[s.Name] = mapToAwaitedContainer(s, containerType, cws.podInfo)
			if cws.containerStateWatcher.ContainerIsInExpectedStatus(cws.containers[s.Name]) {
				return cws.containers[s.Name], true
			}
		}
	}
	return AwaitedContainer{}, false
}
