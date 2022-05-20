package aggregator

import (
	"errors"
	"strconv"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func k8sParsePodBuildID(podMeta metav1.ObjectMeta) (uint, error) {
	buildRef, ok := podMeta.Labels["wharf.iver.com/build-ref"]
	if !ok {
		return 0, errors.New("missing label 'wharf.iver.com/build-ref'")
	}
	buildID, err := strconv.ParseUint(buildRef, 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(buildID), nil
}

func k8sShouldSkipPod(pod v1.Pod) bool {
	if pod.ObjectMeta.DeletionTimestamp != nil {
		return true
	}
	if pod.Status.Phase == v1.PodPending && k8sPodNotErrored(pod) {
		return true
	}
	return false
}

func k8sPodNotErrored(pod v1.Pod) bool {
	for _, s := range pod.Status.InitContainerStatuses {
		if isStuck(s.State.Waiting) {
			return false
		}
	}
	for _, s := range pod.Status.ContainerStatuses {
		if isStuck(s.State.Waiting) {
			return false
		}
	}
	return true
}

func isStuck(s *v1.ContainerStateWaiting) bool {
	if s == nil {
		return false
	}
	switch s.Reason {
	case "CrashLoopBackOff", "ErrImagePull",
		"ImagePullBackOff", "CreateContainerConfigError",
		"InvalidImageName", "CreateContainerError":
		return true
	default:
		return false
	}
}
