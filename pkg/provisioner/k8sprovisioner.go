package provisioner

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"

	"gopkg.in/typ.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type k8sProvisioner struct {
	Namespace              string
	Clientset              *kubernetes.Clientset
	Pods                   corev1.PodInterface
	restConfig             *rest.Config
	instanceID             string
	listOptionsMatchLabels metav1.ListOptions
}

// NewK8sProvisioner returns a new Provisioner implementation that targets
// Kubernetes using a specific Kubernetes namespace and REST config.
func NewK8sProvisioner(instanceID, namespace string, restConfig *rest.Config) (Provisioner, error) {
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return k8sProvisioner{
		Namespace:  namespace,
		Clientset:  clientset,
		Pods:       clientset.CoreV1().Pods(namespace),
		restConfig: restConfig,
		instanceID: instanceID,
		listOptionsMatchLabels: metav1.ListOptions{
			LabelSelector: "app.kubernetes.io/name=wharf-cmd-worker," +
				"app.kubernetes.io/managed-by=wharf-cmd-provisioner," +
				"wharf.iver.com/instance=" + instanceID,
		},
	}, nil
}

func (p k8sProvisioner) ListWorkers(ctx context.Context) ([]Worker, error) {
	podList, err := p.listPods(ctx, p.listOptionsMatchLabels)
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

func (p k8sProvisioner) CreateWorker(ctx context.Context, args WorkerArgs) (Worker, error) {
	if args.GitCloneURL == "" {
		return Worker{}, errors.New("missing required Git clone URL")
	}
	podMeta := p.newWorkerPod(args)
	newPod, err := p.Pods.Create(ctx, &podMeta, metav1.CreateOptions{})
	return convertPodToWorker(newPod), err
}

func (p k8sProvisioner) getPod(ctx context.Context, workerID string) (*v1.Pod, error) {
	podList, err := p.listPods(ctx, p.listOptionsMatchLabels)
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

func (p k8sProvisioner) newWorkerPod(args WorkerArgs) v1.Pod {
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
		"wharf.iver.com/instance":      typ.Coal(args.WharfInstanceID, p.instanceID),
		"wharf.iver.com/build-ref":     uitoa(args.BuildID),
		"wharf.iver.com/project-id":    uitoa(args.ProjectID),
	}
	volumeMounts := []v1.VolumeMount{
		{
			Name:      repoVolumeName,
			MountPath: repoVolumeMountPath,
		},
	}
	gitVolumeMounts := append(volumeMounts, v1.VolumeMount{
		Name:      "ssh",
		ReadOnly:  true,
		MountPath: "/root/.ssh",
	})
	volumes := []v1.Volume{
		{
			Name: repoVolumeName,
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "ssh",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName:  "wharf-cmd-worker-git-ssh",
					DefaultMode: typ.Ref[int32](0600),
					Optional:    typ.Ref(true),
				},
			},
		},
	}

	gitArgs := []string{"git", "clone", args.GitCloneURL, "--single-branch"}
	if args.GitCloneBranch != "" {
		gitArgs = append(gitArgs, "--branch", args.GitCloneBranch)
	}
	gitArgs = append(gitArgs, repoVolumeMountPath)

	wharfArgs := []string{"run", "--loglevel", "debug", "--serve"}
	if args.Environment != "" {
		wharfArgs = append(wharfArgs, "--environment", args.Environment)
	}
	if args.Stage != "" {
		wharfArgs = append(wharfArgs, "--stage", args.Stage)
	}
	for k, v := range args.Inputs {
		wharfArgs = append(wharfArgs, "--input", fmt.Sprintf("%s=%s", k, v))
	}
	if args.SubDir != "" {
		relSubDir := filepath.ToSlash(args.SubDir)
		if filepath.IsAbs(relSubDir) {
			relSubDir = "." + relSubDir
		}
		// Need "--" so the path isn't malliciously treated as another flag
		wharfArgs = append(wharfArgs, "--", relSubDir)
	}

	wharfEnvs := []v1.EnvVar{
		{
			Name:  "WHARF_KUBERNETES_OWNER_ENABLE",
			Value: "true",
		},
		{
			Name: "WHARF_KUBERNETES_OWNER_NAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
		{
			Name: "WHARF_KUBERNETES_OWNER_UID",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.uid",
				},
			},
		},
	}

	for k, v := range args.AdditionalVars {
		wharfEnvs = append(wharfEnvs, v1.EnvVar{
			Name:  "WHARF_VAR_" + k,
			Value: stringify(v),
		})
	}

	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "wharf-cmd-worker-",
			Labels:       labels,
		},
		Spec: v1.PodSpec{
			// TODO: Use serviceaccount name from configs:
			ServiceAccountName: "wharf-cmd",
			RestartPolicy:      v1.RestartPolicyNever,
			InitContainers: []v1.Container{
				{
					Name: "init",
					// TODO: Use image, tag, and pull policy from configs:
					Image:           "bitnami/git:2-debian-10",
					ImagePullPolicy: v1.PullIfNotPresent,
					Command:         gitArgs,
					VolumeMounts:    gitVolumeMounts,
				},
			},
			Containers: []v1.Container{
				{
					Name: "app",
					// TODO: Use image, tag, and pull policy from configs:
					Image:           "quay.io/iver-wharf/wharf-cmd:latest",
					ImagePullPolicy: v1.PullAlways,
					Args:            wharfArgs,
					WorkingDir:      repoVolumeMountPath,
					VolumeMounts:    volumeMounts,
					Env:             wharfEnvs,
				},
			},
			Volumes: volumes,
		},
	}
}

func uitoa(i uint) string {
	return strconv.FormatUint(uint64(i), 10)
}

func stringify(value any) string {
	switch value := value.(type) {
	case string:
		return value
	case nil:
		return "" // fmt.Sprint returns "<nil>", which we don't want
	default:
		return fmt.Sprint(value)
	}
}
