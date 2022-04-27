package provisioner

import (
	"context"
	"fmt"

	"gopkg.in/typ.v3"
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
			AutomountServiceAccountToken: typ.Ref(false),
			ServiceAccountName: "wharf-cmd",
			RestartPolicy:      v1.RestartPolicyNever,
			InitContainers: []v1.Container{
				{
					Name:            "init",
					Image:           "bitnami/git:2-debian-10",
					ImagePullPolicy: v1.PullIfNotPresent,
					// TODO: Should use repo URL and branch from build params.
					Args: []string{
						"git",
						"clone",
						"--single-branch",
						"--branch", "feature/set-k8s-metadata-on-build-pods",
						"https://github.com/iver-wharf/wharf-cmd",
						repoVolumeMountPath,
					},
					VolumeMounts: volumeMounts,
				},
			},
			Containers: []v1.Container{
				{
					Name: "app",
					// TODO: Do some research on which image would be best to use.
					//
					// Note: golang:latest was faster than ubuntu:20.04 up until
					// running `go install`, by around 20 seconds.
					//
					// Note2: The testing environment's internet speed was very
					// fast, so the larger size:
					//  Go: 353MB compressed
					//  ubuntu: 73 MB uncompressed
					// was barely a factor.
					Image:           "golang:latest",
					ImagePullPolicy: v1.PullAlways,
					// TODO: Needs better implementation.
					// Currently works for wharf-cmd, but is not thought through. Takes a long
					// time to get running.
					Command: []string{
						"/bin/sh", "-c",
						`
make deps-go swag install && \
cd test && \
export WHARF_VAR_CHART_REPO="chart_repo" && \
export WHARF_VAR_REG_URL="reg_url" && \
wharf run --serve --stage test --loglevel debug`,
					},
					WorkingDir:   repoVolumeMountPath,
					VolumeMounts: volumeMounts,
					Env: []v1.EnvVar{
						{
							Name: "WHARF_KUBERNETES_OWNER_ENABLE",
							Value: "true",
						},
						{
							Name: "WHARF_KUBERNETES_OWNER_NAME",
							ValueFrom: &v1.EnvVarSource{
								FieldRef: &v1.ObjectFieldSelector{
									APIVersion: "v1",
									FieldPath: "metadata.name",
								},
							},
						},
						{
							Name: "WHARF_KUBERNETES_OWNER_UID",
							ValueFrom: &v1.EnvVarSource{
								FieldRef: &v1.ObjectFieldSelector{
									APIVersion: "v1",
									FieldPath: "metadata.uid",
								},
							},
						},
					},
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
