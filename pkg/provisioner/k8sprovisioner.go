package provisioner

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
var podContainerListArgs = []string{"/bin/sh", "-c", "go install", "&&", "wharf-cmd run"}

type k8sProvisioner struct {
	Namespace  string
	Clientset  *kubernetes.Clientset
	Pods       corev1.PodInterface
	restConfig *rest.Config
}

// NewK8sProvisioner returns a new Provisioner implementation that targets
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
	}, nil
}

func (p k8sProvisioner) ListWorkers(ctx context.Context) ([]Worker, error) {
	podList, err := p.listPods(ctx, listOptionsMatchLabels)
	if err != nil {
		return []Worker{}, err
	}

	return convertPodsToWorkers(podList.Items), nil
}

func (p k8sProvisioner) listPods(ctx context.Context, opts metav1.ListOptions) (*v1.PodList, error) {
	return p.Pods.List(ctx, opts)
}

func (p k8sProvisioner) DeleteWorker(ctx context.Context, workerID string) error {
	pod, err := p.getPod(ctx, workerID)
	if err != nil {
		return err
	}

	return p.Pods.Delete(ctx, pod.Name, metav1.DeleteOptions{})
}

func (p k8sProvisioner) CreateWorker(ctx context.Context) (Worker, error) {
	podMeta := createPodMeta()
	newPod, err := p.Pods.Create(ctx, &podMeta, metav1.CreateOptions{})
	return convertPodToWorker(newPod), err
}

func (p k8sProvisioner) getPod(ctx context.Context, workerID string) (*v1.Pod, error) {
	podList, err := p.listPods(ctx, listOptionsMatchLabels)
	if err != nil {
		return nil, err
	}
	for _, pod := range podList.Items {
		if string(pod.UID) == workerID {
			return &pod, nil
		}
	}
	return nil, fmt.Errorf("found no worker with appropriate labels matching workerID: %s", workerID)
}

func createPodMeta() v1.Pod {
	const (
		repoVolumeName      = "repo"
		repoVolumeMountPath = "/mnt/repo"
	)
	labels := map[string]string{
		"app":                          "wharf-cmd-worker",
		"app.kubernetes.io/name":       "wharf-cmd-worker",
		"app.kubernetes.io/part-of":    "wharf",
		"app.kubernetes.io/managed-by": "wharf-cmd-provisioner",
		"app.kubernetes.io/created-by": "wharf-cmd-provisioner",
		"wharf.iver.com/instance":      "prod",
		"wharf.iver.com/build-ref":     "123",
		"wharf.iver.com/project-id":    "456",
	}
	volumeMounts := []v1.VolumeMount{
		{
			Name:      repoVolumeName,
			MountPath: repoVolumeMountPath,
		},
	}

	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "wharf-cmd-worker-",
			Labels:       labels,
		},
		Spec: v1.PodSpec{
			AutomountServiceAccountToken: new(bool),
			RestartPolicy:                v1.RestartPolicyNever,
			InitContainers: []v1.Container{
				{
					Name:            "init",
					Image:           "bitnami/git:2-debian-10",
					ImagePullPolicy: v1.PullIfNotPresent,
					Args:            append(podInitCloneArgs, "http://github.com/iver-wharf/wharf-cmd", repoVolumeMountPath),
					VolumeMounts:    volumeMounts,
				},
			},
			Containers: []v1.Container{
				{
					Name:            "app",
					Image:           "ubuntu:20.04",
					ImagePullPolicy: v1.PullAlways,
					Command:         podContainerListArgs,
					WorkingDir:      repoVolumeMountPath,
					VolumeMounts:    volumeMounts,
				},
			},
			Volumes: []v1.Volume{
				{
					Name: repoVolumeName,
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}
}
