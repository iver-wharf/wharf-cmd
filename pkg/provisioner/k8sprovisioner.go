package provisioner

import (
	"bufio"
	"context"
	"fmt"

	"github.com/iver-wharf/wharf-core/pkg/logger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

var listOptionsMatchLabels = metav1.ListOptions{
	LabelSelector: "app.kubernetes.io/name=wharf-cmd-worker," +
		"app.kubernetes.io/managed-by=wharf-cmd-provisioner," +
		"wharf.iver.com/instance=prod",
}

var podInitCloneArgs = []string{"git", "clone"}
var podContainerListArgs = []string{"/bin/sh", "-c", "ls -alh"}

type k8sProvisioner struct {
	Namespace  string
	Clientset  *kubernetes.Clientset
	Pods       corev1.PodInterface
	restConfig *rest.Config
	events     corev1.EventInterface
}

// NewK8sProvisioner returns a new provisioner implementation that targets
// Kubernetes using a specific Kubernetes namespace and REST config.
func NewK8sProvisioner(namespace string, restConfig *rest.Config) (Provisioner, error) {
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return k8sProvisioner{
		Namespace:  namespace,
		Clientset:  clientset,
		Pods:       clientset.CoreV1().Pods(namespace),
		restConfig: restConfig,
		events:     clientset.CoreV1().Events(namespace),
	}, nil
}

func (p k8sProvisioner) ListWorkers(ctx context.Context) ([]v1.Pod, error) {
	podList, err := p.Pods.List(ctx, listOptionsMatchLabels)
	if err != nil {
		return nil, err
	}

	return podList.Items, nil
}

func (p k8sProvisioner) DeleteWorker(ctx context.Context, workerID string) error {
	worker, err := p.getWorker(ctx, workerID)
	if err != nil {
		return err
	}

	return p.Pods.Delete(ctx, worker.Name, metav1.DeleteOptions{})
}

func (p k8sProvisioner) CreateWorker(ctx context.Context) (*v1.Pod, error) {
	podMeta := createPodMeta()
	newPod, err := p.Pods.Create(ctx, &podMeta, metav1.CreateOptions{})
	if err != nil {
		return newPod, err
	}

	err = p.waitForInitContainerDone(ctx, newPod.ObjectMeta)
	if err != nil {
		return newPod, err
	}

	log.Debug().Message("Init Container done.")

	err = p.waitForAppContainerRunning(ctx, newPod.ObjectMeta)
	if err != nil {
		return newPod, err
	}

	log.Debug().Message("App Container running.")

	err = p.streamLogsUntilCompleted(ctx, newPod.Name)
	return newPod, err
}

func (p k8sProvisioner) getWorker(ctx context.Context, workerID string) (*v1.Pod, error) {
	workers, err := p.ListWorkers(ctx)
	if err != nil {
		return nil, err
	}
	for _, pod := range workers {
		if string(pod.ObjectMeta.UID) == workerID {
			return &pod, nil
		}
	}
	return nil, fmt.Errorf("found no pod with appropriate labels matching workerID: %s", workerID)
}

func createPodMeta() v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "wharf-provisioner",
			Labels: map[string]string{
				"app":                          "wharf-cmd-worker",
				"app.kubernetes.io/name":       "wharf-cmd-worker",
				"app.kubernetes.io/part-of":    "wharf",
				"app.kubernetes.io/managed-by": "wharf-cmd-provisioner",
				"app.kubernetes.io/created-by": "wharf-cmd-provisioner",
				"wharf.iver.com/instance":      "prod",
				"wharf.iver.com/build-ref":     "123",
				"wharf.iver.com/project-id":    "456",
			},
		},
		Spec: v1.PodSpec{
			AutomountServiceAccountToken: new(bool),
			RestartPolicy:                v1.RestartPolicyNever,
			InitContainers: []v1.Container{
				{
					Name:            "init",
					Image:           "bitnami/git:2-debian-10",
					ImagePullPolicy: v1.PullIfNotPresent,
					Args:            append(podInitCloneArgs, "http://github.com/iver-wharf/wharf-cmd", "/mnt/repo"),
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "repo",
							MountPath: "/mnt/repo",
						},
					},
				},
			},
			Containers: []v1.Container{
				{
					Name:            "app",
					Image:           "ubuntu:20.04",
					ImagePullPolicy: v1.PullAlways,
					Command:         podContainerListArgs,
					WorkingDir:      "/mnt/repo",
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "repo",
							MountPath: "/mnt/repo",
						},
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "repo",
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}
}

func (p k8sProvisioner) waitForInitContainerDone(ctx context.Context, podMeta metav1.ObjectMeta) error {
	return p.waitForPodModifiedFunc(ctx, podMeta, func(p *v1.Pod) (bool, error) {
		return isInitContainerDone(p)
	})
}

func (p k8sProvisioner) waitForAppContainerRunning(ctx context.Context, podMeta metav1.ObjectMeta) error {
	return p.waitForPodModifiedFunc(ctx, podMeta, func(p *v1.Pod) (bool, error) {
		return isAppContainerRunning(p)
	})
}

func isInitContainerDone(pod *v1.Pod) (bool, error) {
	for _, c := range pod.Status.InitContainerStatuses {
		if c.State.Terminated != nil {
			if c.State.Terminated.ExitCode != 0 {
				return true, fmt.Errorf("non-zero exit code: %d", c.State.Terminated.ExitCode)
			}
			return true, nil
		}
	}
	return len(pod.Status.InitContainerStatuses) == 0, nil
}

func isAppContainerRunning(pod *v1.Pod) (bool, error) {
	for _, c := range pod.Status.ContainerStatuses {
		if c.State.Terminated != nil {
			if c.State.Terminated.ExitCode != 0 {
				return false, fmt.Errorf("non-zero exit code: %d", c.State.Terminated.ExitCode)
			}
			return true, nil
		}
		if c.State.Running != nil {
			return true, nil
		}
	}
	return false, nil
}

func (p k8sProvisioner) waitForPodModifiedFunc(ctx context.Context, podMeta metav1.ObjectMeta, f func(p *v1.Pod) (bool, error)) error {
	w, err := p.Clientset.CoreV1().Pods(p.Namespace).Watch(ctx, metav1.SingleObject(podMeta))
	if err != nil {
		return err
	}

	defer w.Stop()
	for ev := range w.ResultChan() {
		pod := ev.Object.(*v1.Pod)
		switch ev.Type {
		case watch.Modified:
			ok, err := f(pod)
			if err != nil {
				return err
			} else if ok {
				return nil
			}
		case watch.Deleted:
			return fmt.Errorf("pod was removed: %v", pod.Name)
		}
	}
	return fmt.Errorf("got no more events when watching pod: %v", podMeta.Name)
}

func (p k8sProvisioner) streamLogsUntilCompleted(ctx context.Context, podName string) error {
	req := p.Pods.GetLogs(podName, &v1.PodLogOptions{
		Follow: true,
	})
	readCloser, err := req.Stream(ctx)
	if err != nil {
		return err
	}
	defer readCloser.Close()
	podLog := logger.NewScoped(podName)
	scanner := bufio.NewScanner(readCloser)
	for scanner.Scan() {
		podLog.Info().Message(scanner.Text())
	}
	return scanner.Err()
}
