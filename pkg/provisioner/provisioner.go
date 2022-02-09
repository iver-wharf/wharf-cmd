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

var log = logger.NewScoped("PROVISIONER")

var podInitCloneArgs = []string{"git", "clone"}
var podContainerListArgs = []string{"/bin/sh", "-c", "ls -alh"}

// Provisioner is an interface declaring what methods are required
// for a provisioner.
type Provisioner interface {
	Serve(ctx context.Context) error
	ListPods(ctx context.Context) error
	DeletePod(ctx context.Context, UID string) error
}

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

func (p k8sProvisioner) ListPods(ctx context.Context) error {
	podList, err := p.Pods.List(ctx, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=wharf-cmd-worker,app.kubernetes.io/managed-by=wharf-cmd-provisioner,wharf.iver.com/instance=prod",
	})
	if err != nil {
		return err
	}
	pods := podList.Items

	printPod := func(pod metav1.ObjectMeta, i int) func(logger.Event) logger.Event {
		return func(ev logger.Event) logger.Event {
			ev = ev.WithInt("index", i).
				WithString("UID", string(pod.UID)).
				WithString("name", pod.Name).
				WithString("namespace", pod.Namespace)

			for k, v := range pod.Labels {
				ev = ev.WithString(k, v)
			}

			return ev
		}
	}

	log.Info().WithInt("count", len(pods)).Message("Fetched pods with matching labels.")
	for i, pod := range pods {
		log.Info().WithFunc(printPod(pod.ObjectMeta, i)).Message("Pod")
	}

	return nil
}

func (p k8sProvisioner) DeletePod(ctx context.Context, UID string) error {
	podList, err := p.Pods.List(ctx, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=wharf-cmd-worker,app.kubernetes.io/managed-by=wharf-cmd-provisioner,wharf.iver.com/instance=prod",
	})
	if err != nil {
		return err
	}

	var matchingPod *v1.Pod
	for _, pod := range podList.Items {
		if string(pod.ObjectMeta.UID) == UID {
			matchingPod = &pod
			break
		}
	}

	if matchingPod == nil {
		return fmt.Errorf("found no pod with appropriate labels matching UID: %s", UID)
	}

	return p.Pods.Delete(ctx, matchingPod.Name, metav1.DeleteOptions{})
}

func (p k8sProvisioner) Serve(ctx context.Context) error {
	podMeta := createPodMeta()

	newPod, err := p.Pods.Create(ctx, &podMeta, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	err = p.waitForInitContainerDone(ctx, newPod.ObjectMeta)
	if err != nil {
		return err
	}

	log.Debug().Message("Init Container done.")

	err = p.waitForAppContainerRunning(ctx, newPod.ObjectMeta)
	if err != nil {
		return err
	}

	log.Debug().Message("App Container running.")

	err = p.streamLogsUntilCompleted(ctx, newPod.Name)
	return err
}

func createPodMeta() v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "wharf-provisioner",
			// GenerateName: "wharf-provisioner",
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
