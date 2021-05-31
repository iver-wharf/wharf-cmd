package kubernetes

import v1 "k8s.io/api/core/v1"

type podInfo struct {
	namespace string
	podUID    string
	name      string
}

func mapToPodInfo(pod *v1.Pod) podInfo {
	if pod == nil {
		return podInfo{}
	}
	
	return podInfo{
		namespace: pod.Namespace,
		podUID:    string(pod.UID),
		name:      pod.Name,
	}
}