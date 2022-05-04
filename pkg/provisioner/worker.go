package provisioner

import (
	"fmt"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/worker/workermodel"
	v1 "k8s.io/api/core/v1"
)

// Worker contains the data of a worker, an abstraction over k8s pods which is
// rougly comparable to a group of Docker containers sharing
// namespaces/volumes.
type Worker struct {
	WorkerID  string             `json:"workerId"`
	Name      string             `json:"name"`
	Status    workermodel.Status `json:"status"`
	CreatedAt time.Time          `json:"createdAt"`
}

func convertPodsToWorkers(pods []v1.Pod) []Worker {
	workers := make([]Worker, len(pods))
	for k, v := range pods {
		workers[k] = convertPodToWorker(&v)
	}
	return workers
}

func convertPodToWorker(pod *v1.Pod) Worker {
	if pod == nil {
		return Worker{}
	}
	var status workermodel.Status = workermodel.StatusUnknown
	if pod.Status.Phase == v1.PodUnknown {
		status = workermodel.StatusUnknown
	} else if pod.Status.Phase == v1.PodSucceeded {
		status = workermodel.StatusSuccess
	} else if pod.Status.Phase == v1.PodFailed {
		status = workermodel.StatusFailed
	} else if anyContainerIsRunning(pod.Status.InitContainerStatuses) {
		status = workermodel.StatusInitializing
	} else if anyContainerIsRunning(pod.Status.ContainerStatuses) {
		status = workermodel.StatusRunning
	} else if podConditionIsTrue(pod.Status.Conditions, v1.PodScheduled) {
		status = workermodel.StatusScheduling
	}

	return Worker{
		WorkerID:  string(pod.UID),
		Name:      fmt.Sprintf("%s/%s", pod.Namespace, pod.Name),
		Status:    status,
		CreatedAt: pod.CreationTimestamp.Time,
	}
}

func podConditionIsTrue(conditions []v1.PodCondition, conditionType v1.PodConditionType) bool {
	for _, v := range conditions {
		if v.Type == conditionType {
			return v.Status == v1.ConditionTrue
		}
	}
	return false
}

func anyContainerIsRunning(containers []v1.ContainerStatus) bool {
	for _, v := range containers {
		if v.State.Terminated == nil {
			return true
		}
	}
	return false
}
