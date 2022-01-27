package builder

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/iver-wharf/wharf-cmd/pkg/core/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/tarutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

var podInitWaitArgs = []string{"/bin/sh", "-c", "sleep infinite || true"}
var podInitContinueArgs = []string{"killall", "-s", "SIGINT", "sleep"}

type k8sStepRunner struct {
	namespace  string
	restConfig *rest.Config
	clientset  *kubernetes.Clientset
	pods       corev1.PodInterface
}

func NewK8sStepRunner(namespace string, restConfig *rest.Config, clientset *kubernetes.Clientset) StepRunner {
	return k8sStepRunner{
		namespace:  namespace,
		restConfig: restConfig,
		clientset:  clientset,
		pods:       clientset.CoreV1().Pods(namespace),
	}
}

func (r k8sStepRunner) RunStep(step wharfyml.Step) StepResult {
	pod, err := getPodSpec(step)
	if err != nil {
		return StepResult{Error: err}
	}
	newPod, err := r.pods.Create(context.TODO(), &pod, metav1.CreateOptions{})
	if err != nil {
		return StepResult{Error: fmt.Errorf("create pod: %w", err)}
	}
	log.Debug().
		WithString("step", step.Name).
		WithString("pod", newPod.Name).
		Message("Created pod.")
	defer func() {
		err := r.pods.Delete(context.TODO(), newPod.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Warn().
				WithString("step", step.Name).
				WithString("pod", newPod.Name).
				Message("Failed to delete pod.")
		} else {
			log.Debug().
				WithString("step", step.Name).
				WithString("pod", newPod.Name).
				Message("Deleted pod.")
		}
	}()
	if err := r.copyDirToPod(".", "/mnt/repo", r.namespace, newPod.Name, "init"); err != nil {
		return StepResult{Error: fmt.Errorf("transfer repo: %w", err)}
	}
	log.Debug().
		WithString("step", step.Name).
		WithString("pod", newPod.Name).
		Message("Transferred repo to init container.")
	if err := r.continueInitContainer(newPod.Name); err != nil {
		return StepResult{Error: fmt.Errorf("continue init container: %w", err)}
	}
	log.Debug().
		WithString("step", step.Name).
		WithString("pod", newPod.Name).
		Message("Stopped sleep in init container.")
	// TODO: Wait for pod to complete run
	return StepResult{Error: errors.New("not implemented")}
}

func (r k8sStepRunner) continueInitContainer(podName string) error {
	exec, err := execInPodNoPipe(r.restConfig, r.namespace, podName, "init", podInitContinueArgs)
	if err != nil {
		return err
	}
	exec.Stream(remotecommand.StreamOptions{})
	return nil
}

func getPodSpec(step wharfyml.Step) (v1.Pod, error) {
	image, err := getPodImage(step)
	if err != nil {
		return v1.Pod{}, err
	}
	cmds, args, err := getPodCommandArgs(step)
	if err != nil {
		return v1.Pod{}, err
	}
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("wharf-build-%s-%s-",
				strings.ToLower(step.Type.String()),
				strings.ToLower(step.Name)),
			Annotations: map[string]string{
				"wharf.iver.com/step": step.Name,
			},
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			InitContainers: []v1.Container{
				{
					Name:            "init",
					Image:           "alpine:3",
					ImagePullPolicy: v1.PullAlways,
					Command:         podInitWaitArgs,
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
					Name:            "step",
					Image:           image,
					ImagePullPolicy: v1.PullAlways,
					Command:         cmds,
					Args:            args,
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

func getPodImage(step wharfyml.Step) (string, error) {
	switch step.Type {
	case wharfyml.Container:
		image, ok := step.Variables["image"]
		if !ok {
			return "", errors.New("missing required field: image")
		}
		imageStr, ok := image.(string)
		if !ok {
			return "", fmt.Errorf("invalid field type: image: want string, got: %T", image)
		}
		return imageStr, nil
	default:
		return "", fmt.Errorf("unsupported step type: %q", step.Type)
	}
}

func getPodCommandArgs(step wharfyml.Step) (cmds, args []string, err error) {
	switch step.Type {
	case wharfyml.Container:
		cmdsAny, ok := step.Variables["cmds"]
		if !ok {
			return nil, nil, errors.New("missing required field: cmds")
		}
		cmds, err := convStepFieldToStrings("cmds", cmdsAny)
		if err != nil {
			return nil, nil, err
		}
		shell := "/bin/sh"
		if shellAny, ok := step.Variables["shell"]; ok {
			shell, ok = shellAny.(string)
			if !ok {
				return nil, nil, fmt.Errorf("invalid field type: shell: want string, got %T", shellAny)
			}
		}
		return []string{shell, "-c"}, cmds, nil
	default:
		return nil, nil, fmt.Errorf("unsupported step type: %q", step.Type)
	}
}

func convStepFieldToStrings(fieldName string, value interface{}) ([]string, error) {
	anyArr, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid field type: %s: want string array, got: %T", fieldName, value)
	}
	strs := make([]string, 0, len(anyArr))
	for i, v := range anyArr {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("invalid field type: %s: index %d: want string, got: %T", fieldName, i, value)
		}
		strs = append(strs, str)
	}
	return strs, nil
}

func (r k8sStepRunner) copyDirToPod(srcPath, destPath, namespace, podName, containerName string) error {
	// Based on: https://stackoverflow.com/a/57952887
	reader, writer := io.Pipe()
	args := []string{"tar", "-xf", "-", "-C", destPath}
	exec, err := execInPodPipedStdin(r.restConfig, namespace, podName, containerName, args)
	if err != nil {
		return err
	}
	var tarErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func(writer io.WriteCloser, wg *sync.WaitGroup) {
		defer wg.Done()
		defer writer.Close()
		defer func() {
			if p := recover(); p != nil {
				tarErr = fmt.Errorf("panic: %v", wg)
			}
		}()
		tarErr = tarutil.Dir(writer, srcPath)
	}(writer, &wg)
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin: reader,
		Tty:   false,
	})
	if err != nil {
		return err
	}
	wg.Wait()
	return tarErr
}

func execInPodPipedStdin(c *rest.Config, namespace, podName, containerName string, args []string) (remotecommand.Executor, error) {
	return execInPod(c, namespace, podName, &v1.PodExecOptions{
		Container: containerName,
		Command:   args,
		Stdin:     true,
		Stdout:    false,
		Stderr:    false,
		TTY:       false,
	})
}

func execInPodNoPipe(c *rest.Config, namespace, podName, containerName string, args []string) (remotecommand.Executor, error) {
	return execInPod(c, namespace, podName, &v1.PodExecOptions{
		Container: containerName,
		Command:   args,
		Stdin:     false,
		Stdout:    false,
		Stderr:    false,
		TTY:       false,
	})
}

func execInPod(c *rest.Config, namespace, podName string, execOpts *v1.PodExecOptions) (remotecommand.Executor, error) {
	coreclient, err := corev1client.NewForConfig(c)
	if err != nil {
		return nil, err
	}
	req := coreclient.RESTClient().
		Post().
		Namespace(namespace).
		Resource("pods").
		Name(podName).
		SubResource("exec").
		VersionedParams(execOpts, metav1.ParameterCodec)
	return remotecommand.NewSPDYExecutor(c, "POST", req.URL())
}
