package builder

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/core/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/tarutil"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
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
	events     corev1.EventInterface
}

func NewK8sStepRunner(namespace string, restConfig *rest.Config) (StepRunner, error) {
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return k8sStepRunner{
		namespace:  namespace,
		restConfig: restConfig,
		clientset:  clientset,
		pods:       clientset.CoreV1().Pods(namespace),
		events:     clientset.CoreV1().Events(namespace),
	}, nil
}

func (r k8sStepRunner) RunStep(step wharfyml.Step) StepResult {
	start := time.Now()
	err := r.runStepError(step)
	return StepResult{
		Name:     step.Name,
		Type:     step.Type.String(),
		Success:  err == nil,
		Error:    err,
		Duration: time.Since(start),
	}
}

func (r k8sStepRunner) runStepError(step wharfyml.Step) error {
	pod, err := getPodSpec(step)
	if err != nil {
		return err
	}
	log.Debug().
		WithString("step", step.Name).
		WithString("pod", pod.GenerateName).
		Message("Creating pod.")
	newPod, err := r.pods.Create(context.TODO(), &pod, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create pod: %w", err)
	}
	var logFunc = func(ev logger.Event) logger.Event {
		return ev.
			WithString("step", step.Name).
			WithString("pod", newPod.Name)
	}
	log.Debug().WithFunc(logFunc).Message("Created pod.")
	defer r.stopPodNow(step.Name, newPod.Name)
	log.Debug().WithFunc(logFunc).Message("Waiting for init container to start.")
	if err := r.waitForInitContainerRunning(newPod.ObjectMeta); err != nil {
		return fmt.Errorf("wait for init container: %w", err)
	}
	log.Debug().WithFunc(logFunc).Message("Transferring repo to init container.")
	if err := r.copyDirToPod(".", "/mnt/repo", r.namespace, newPod.Name, "init"); err != nil {
		return fmt.Errorf("transfer repo: %w", err)
	}
	log.Debug().WithFunc(logFunc).Message("Transferred repo to init container.")
	if err := r.continueInitContainer(newPod.Name); err != nil {
		return fmt.Errorf("continue init container: %w", err)
	}
	log.Debug().WithFunc(logFunc).Message("Waiting for app container to start.")
	if err := r.waitForAppContainerRunningOrDone(newPod.ObjectMeta); err != nil {
		return fmt.Errorf("wait for app container: %w", err)
	}
	if err := r.streamLogsUntilCompleted(step.Name, newPod.Name); err != nil {
		return fmt.Errorf("stream logs: %w", err)
	}
	return nil
}

func (r k8sStepRunner) waitForInitContainerRunning(podMeta metav1.ObjectMeta) error {
	return r.waitForPodModifiedFunc(podMeta, func(pod *v1.Pod) (bool, error) {
		for _, c := range pod.Status.InitContainerStatuses {
			if c.State.Running != nil {
				return true, nil
			}
		}
		return false, nil
	})
}

func (r k8sStepRunner) waitForAppContainerRunningOrDone(podMeta metav1.ObjectMeta) error {
	return r.waitForPodModifiedFunc(podMeta, func(pod *v1.Pod) (bool, error) {
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
	})
}

func (r k8sStepRunner) waitForPodModifiedFunc(podMeta metav1.ObjectMeta, f func(pod *v1.Pod) (bool, error)) error {
	w, err := r.pods.Watch(context.TODO(), metav1.SingleObject(podMeta))
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

func (r k8sStepRunner) streamLogsUntilCompleted(logScope, podName string) error {
	req := r.pods.GetLogs(podName, &v1.PodLogOptions{
		Follow: true,
	})
	readCloser, err := req.Stream(context.TODO())
	if err != nil {
		return err
	}
	defer readCloser.Close()
	podLog := logger.NewScoped(logScope)
	scanner := bufio.NewScanner(readCloser)
	for scanner.Scan() {
		podLog.Info().Message(scanner.Text())
	}
	return scanner.Err()
}

func (r k8sStepRunner) stopPodNow(stepName, podName string) {
	var gracePeriod int64 = 0 // 0=immediately
	err := r.pods.Delete(context.TODO(), podName, metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	})
	if err != nil {
		log.Warn().
			WithString("step", stepName).
			WithString("pod", podName).
			Message("Failed to delete pod.")
	} else {
		log.Debug().
			WithString("step", stepName).
			WithString("pod", podName).
			Message("Deleted pod.")
	}
}

func (r k8sStepRunner) continueInitContainer(podName string) error {
	exec, err := execInPodPipeStdout(r.restConfig, r.namespace, podName, "init", podInitContinueArgs)
	if err != nil {
		return err
	}
	exec.Stream(remotecommand.StreamOptions{
		Stdout: nopWriter{},
	})
	return nil
}

func (r k8sStepRunner) copyDirToPod(srcPath, destPath, namespace, podName, containerName string) error {
	// Based on: https://stackoverflow.com/a/57952887
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	args := []string{"tar", "-xf", "-", "-C", destPath}
	exec, err := execInPodPipedStdin(r.restConfig, namespace, podName, containerName, args)
	if err != nil {
		return err
	}
	tarErrCh := make(chan error, 1)
	go func(writer io.WriteCloser, ch chan<- error) {
		ch <- tarutil.Dir(writer, srcPath)
		writer.Close()
	}(writer, tarErrCh)
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  reader,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		return err
	}
	return <-tarErrCh
}

func execInPodPipedStdin(c *rest.Config, namespace, podName, containerName string, args []string) (remotecommand.Executor, error) {
	return execInPod(c, namespace, podName, &v1.PodExecOptions{
		Container: containerName,
		Command:   args,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	})
}

func execInPodPipeStdout(c *rest.Config, namespace, podName, containerName string, args []string) (remotecommand.Executor, error) {
	return execInPod(c, namespace, podName, &v1.PodExecOptions{
		Container: containerName,
		Command:   args,
		Stdin:     false,
		Stdout:    true,
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
		VersionedParams(execOpts, scheme.ParameterCodec)
	return remotecommand.NewSPDYExecutor(c, "POST", req.URL())
}
