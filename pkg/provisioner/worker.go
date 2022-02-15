package provisioner

import (
	"fmt"

	"github.com/iver-wharf/wharf-cmd/pkg/worker"
	v1 "k8s.io/api/core/v1"
)

// Worker contains the data of a worker, an abstraction over k8s pods which is
// rougly comparable to a group of Docker containers sharing
// namespaces/volumes.
type Worker struct {
	ID     string
	Name   string
	Status worker.Status
}

// WorkerList contains a slice of workers and a count field for the size of the
// array.
type WorkerList struct {
	Items []Worker
	Count int
}

func convertPodListToWorkerList(podList *v1.PodList) WorkerList {
	if podList == nil {
		return WorkerList{}
	}
	return WorkerList{
		Items: convertPodsToWorkers(podList.Items),
		Count: len(podList.Items),
	}
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
	return Worker{
		ID:     string(pod.UID),
		Name:   fmt.Sprintf("%s/%s", pod.Namespace, pod.Name),
		Status: worker.StatusUnknown,
	}
}
