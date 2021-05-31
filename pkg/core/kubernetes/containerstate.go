package kubernetes

import v1 "k8s.io/api/core/v1"

type ContainerState int

const (
	ContainerUnknownState = ContainerState(iota)
	ContainerWaiting
	ContainerRunning
	ContainerTerminated
)

func mapToContainerState(state v1.ContainerState) ContainerState {
	if state.Waiting != nil {
		return ContainerWaiting
	} else if state.Running != nil {
		return ContainerRunning
	} else if state.Terminated != nil {
		return ContainerTerminated
	}
	return ContainerUnknownState
}
