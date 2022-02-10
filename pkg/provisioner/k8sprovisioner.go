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
var podContainerListArgs = []string{"/bin/sh", "-c", "ls -alh"}

type k8sProvisioner struct {
	Namespace  string
	Clientset  *kubernetes.Clientset
	Pods       corev1.PodInterface
	restConfig *rest.Config
	events     corev1.EventInterface
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
	const (
		gitCloneContainerName = "init"
		gitCloneImage         = "bitnami/git:2-debian-10"
		gitURL                = "http://github.com/iver-wharf/wharf-cmd"
		repoVolumeName        = "repo"
		repoVolumeMountPath   = "/mnt/repo"

		workerContainerName = "app"
		workerImage         = "ubuntu:20.04"

		labelWorkerInstance  = "prod"
		labelWorkerBuildRef  = "123"
		labelWorkerProjectID = "456"
	)
	extraLabels := make(map[string]string)
	labels := map[string]string{
		"app":                          "wharf-cmd-worker",
		"app.kubernetes.io/name":       "wharf-cmd-worker",
		"app.kubernetes.io/part-of":    "wharf",
		"app.kubernetes.io/managed-by": "wharf-cmd-provisioner",
		"app.kubernetes.io/created-by": "wharf-cmd-provisioner",
		"wharf.iver.com/instance":      labelWorkerInstance,
		"wharf.iver.com/build-ref":     labelWorkerBuildRef,
		"wharf.iver.com/project-id":    labelWorkerProjectID,
	}
	for k, v := range extraLabels {
		labels[k] = v
	}

	volumeMounts := []v1.VolumeMount{
		{
			Name:      repoVolumeName,
			MountPath: repoVolumeMountPath,
		},
	}

	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "wharf-provisioner-",
			Labels:       labels,
		},
		Spec: v1.PodSpec{
			AutomountServiceAccountToken: new(bool),
			RestartPolicy:                v1.RestartPolicyNever,
			InitContainers: []v1.Container{
				{
					Name:            gitCloneContainerName,
					Image:           gitCloneImage,
					ImagePullPolicy: v1.PullIfNotPresent,
					Args:            append(podInitCloneArgs, gitURL, repoVolumeMountPath),
					VolumeMounts:    volumeMounts,
				},
			},
			Containers: []v1.Container{
				{
					Name:            workerContainerName,
					Image:           workerImage,
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
