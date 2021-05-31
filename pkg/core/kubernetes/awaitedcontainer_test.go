package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestMapToAwaitedContainer(t *testing.T) {
	podInfo := podInfo{
		namespace: "default",
		podUID:    "1234567890-poiuytr",
		name:      "super-pod",
	}

	containerStatus := v1.ContainerStatus{
		Name: "test",
		State: v1.ContainerState{
			Waiting:    nil,
			Running:    &v1.ContainerStateRunning{},
			Terminated: nil,
		},
		LastTerminationState: v1.ContainerState{},
		Ready:                false,
		RestartCount:         999,
		Image:                "",
		ImageID:              "",
		ContainerID:          "",
	}

	awaitedContainer := mapToAwaitedContainer(containerStatus, ContainerTypeInit, podInfo)

	assert.Equal(t, ContainerTypeInit, awaitedContainer.Type)

	assert.Equal(t, false, awaitedContainer.isReady)
	assert.Equal(t, 999, awaitedContainer.Restarts)
	assert.Equal(t, "test", awaitedContainer.Name)

	assert.Equal(t, ContainerRunning, awaitedContainer.State)

	assert.Equal(t, podInfo.namespace, awaitedContainer.Namespace)
	assert.Equal(t, podInfo.name, awaitedContainer.PodName)
	assert.Equal(t, podInfo.podUID, awaitedContainer.PodUID)
}
