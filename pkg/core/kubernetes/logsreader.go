package kubernetes

import (
	"context"
	"fmt"
	"io"

	kubecore "k8s.io/api/core/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type ContainerLogsReader interface {
	StreamContainerLogs(podName string, containerName string) (io.ReadCloser, error)
}

type containerLogsReader struct {
	podInterface corev1.PodInterface
}

func NewContainerLogsReader(podInterface corev1.PodInterface) ContainerLogsReader {
	return &containerLogsReader{
		podInterface: podInterface,
	}
}

func (r *containerLogsReader) StreamContainerLogs(podName string, containerName string) (io.ReadCloser, error) {
	request := r.podInterface.GetLogs(podName, &kubecore.PodLogOptions{
		Follow:    true,
		Container: containerName,
	})

	if request == nil {
		return nil, fmt.Errorf("get log request is nil")
	}

	stream, err := request.Stream(context.TODO())
	if err != nil {
		log.Error().WithError(err).Message("Failed to stream logs.")
		return nil, err
	}

	return stream, nil
}
