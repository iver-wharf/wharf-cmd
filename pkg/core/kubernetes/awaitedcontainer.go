package kubernetes

import v1 "k8s.io/api/core/v1"

type AwaitedContainer struct {
	Namespace string
	PodUID    string
	PodName   string
	Name      string
	Type      ContainerType
	State     ContainerState
	Restarts  int
	isReady   bool
}

func mapToAwaitedContainer(containerStatus v1.ContainerStatus, containerType ContainerType, podInfo podInfo) AwaitedContainer {
	return AwaitedContainer{
		Name:      containerStatus.Name,
		PodName:   podInfo.name,
		PodUID:    podInfo.podUID,
		Namespace: podInfo.namespace,
		Type:      containerType,
		State:     mapToContainerState(containerStatus.State),
		Restarts:  int(containerStatus.RestartCount),
		isReady:   containerStatus.Ready,
	}
}
