package worker

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/tarutil"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
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

// NewK8sStepRunner returns a new step runner implementation that targets
// Kubernetes using a specific Kubernetes namespace and REST config.
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

func (r k8sStepRunner) RunStep(ctx context.Context, step wharfyml.Step) StepResult {
	ctx = contextWithStepName(ctx, step.Name)
	start := time.Now()
	err := r.runStepError(ctx, step)
	status := StatusSuccess
	if errors.Is(err, context.Canceled) {
		status = StatusCancelled
	} else if err != nil {
		status = StatusFailed
	}
	return StepResult{
		Name:     step.Name,
		Status:   status,
		Type:     step.Type.StepTypeName(),
		Error:    err,
		Duration: time.Since(start),
	}
}

func (r k8sStepRunner) runStepError(ctx context.Context, step wharfyml.Step) error {
	pod, err := getPodSpec(ctx, step)
	if err != nil {
		return err
	}
	log.Debug().
		WithString("step", step.Name).
		WithString("pod", pod.GenerateName).
		Message("Creating pod.")
	newPod, err := r.pods.Create(ctx, &pod, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create pod: %w", err)
	}
	var logFunc = func(ev logger.Event) logger.Event {
		return ev.
			WithString("step", step.Name).
			WithString("pod", newPod.Name)
	}
	log.Debug().WithFunc(logFunc).Message("Created pod.")
	defer r.stopPodNow(context.Background(), step.Name, newPod.Name)
	log.Debug().WithFunc(logFunc).Message("Waiting for init container to start.")
	if err := r.waitForInitContainerRunning(ctx, newPod.ObjectMeta); err != nil {
		return fmt.Errorf("wait for init container: %w", err)
	}
	log.Debug().WithFunc(logFunc).Message("Transferring repo to init container.")
	if err := r.copyDirToPod(ctx, ".", "/mnt/repo", r.namespace, newPod.Name, "init"); err != nil {
		return fmt.Errorf("transfer repo: %w", err)
	}
	log.Debug().WithFunc(logFunc).Message("Transferred repo to init container.")
	if err := r.continueInitContainer(newPod.Name); err != nil {
		return fmt.Errorf("continue init container: %w", err)
	}
	log.Debug().WithFunc(logFunc).Message("Waiting for app container to start.")
	if err := r.waitForAppContainerRunningOrDone(ctx, newPod.ObjectMeta); err != nil {
		if err := r.readLogs(ctx, newPod.Name, &v1.PodLogOptions{}); err != nil {
			log.Debug().WithError(err).
				Message("Failed to read logs from failed container.")
		}
		return fmt.Errorf("wait for app container: %w", err)
	}
	log.Debug().WithFunc(logFunc).Message("App container running. Streaming logs.")
	if err := r.readLogs(ctx, newPod.Name, &v1.PodLogOptions{Follow: true}); err != nil {
		return fmt.Errorf("stream logs: %w", err)
	}
	log.Debug().WithFunc(logFunc).Message("Logs ended. Waiting for termination.")
	return r.waitForAppContainerDone(ctx, newPod.ObjectMeta)
}

func (r k8sStepRunner) waitForInitContainerRunning(ctx context.Context, podMeta metav1.ObjectMeta) error {
	return r.waitForPodModifiedFunc(ctx, podMeta, func(pod *v1.Pod) (bool, error) {
		for _, c := range pod.Status.InitContainerStatuses {
			if c.State.Running != nil {
				return true, nil
			}
		}
		return false, nil
	})
}

func (r k8sStepRunner) waitForAppContainerRunningOrDone(ctx context.Context, podMeta metav1.ObjectMeta) error {
	return r.waitForPodModifiedFunc(ctx, podMeta, func(pod *v1.Pod) (bool, error) {
		for _, c := range pod.Status.ContainerStatuses {
			if c.State.Terminated != nil {
				if c.State.Terminated.ExitCode != 0 {
					return false, fmt.Errorf("non-zero exit code: %d", c.State.Terminated.ExitCode)
				}
				return true, nil
			}
			if c.State.Waiting != nil &&
				c.State.Waiting.Reason == "CreateContainerConfigError" {
				return false, fmt.Errorf("config error: %s", c.State.Waiting.Message)
			}
			if c.State.Running != nil {
				return true, nil
			}
		}
		return false, nil
	})
}

func (r k8sStepRunner) waitForAppContainerDone(ctx context.Context, podMeta metav1.ObjectMeta) error {
	return r.waitForPodModifiedFunc(ctx, podMeta, func(pod *v1.Pod) (bool, error) {
		for _, c := range pod.Status.ContainerStatuses {
			if c.State.Terminated != nil {
				if c.State.Terminated.ExitCode != 0 {
					return false, fmt.Errorf("non-zero exit code: %d", c.State.Terminated.ExitCode)
				}
				return true, nil
			}
		}
		return false, nil
	})
}

func (r k8sStepRunner) waitForPodModifiedFunc(ctx context.Context, podMeta metav1.ObjectMeta, f func(pod *v1.Pod) (bool, error)) error {
	w, err := r.pods.Watch(ctx, metav1.SingleObject(podMeta))
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

func (r k8sStepRunner) readLogs(ctx context.Context, podName string, opts *v1.PodLogOptions) error {
	req := r.pods.GetLogs(podName, opts)
	readCloser, err := req.Stream(ctx)
	if err != nil {
		return err
	}
	defer readCloser.Close()
	podLog := logger.NewScoped(contextStageStepName(ctx))
	scanner := bufio.NewScanner(readCloser)
	for scanner.Scan() {
		txt := scanner.Text()
		idx := strings.LastIndexByte(txt, '\r')
		if idx != -1 {
			txt = txt[idx+1:]
		}
		podLog.Info().Message(txt)
	}
	return scanner.Err()
}

func (r k8sStepRunner) stopPodNow(ctx context.Context, stepName, podName string) {
	gracePeriod := int64(0) // 0=immediately
	err := r.pods.Delete(ctx, podName, metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	})
	if err != nil {
		log.Warn().
			WithError(err).
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

func (r k8sStepRunner) copyDirToPod(ctx context.Context, srcPath, destPath, namespace, podName, containerName string) error {
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
		Stdin: reader,
	})
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return errors.New("aborted")
	case err := <-tarErrCh:
		return err
	}
}

func execInPodPipedStdin(c *rest.Config, namespace, podName, containerName string, args []string) (remotecommand.Executor, error) {
	return execInPod(c, namespace, podName, &v1.PodExecOptions{
		Container: containerName,
		Command:   args,
		Stdin:     true,
	})
}

func execInPodPipeStdout(c *rest.Config, namespace, podName, containerName string, args []string) (remotecommand.Executor, error) {
	return execInPod(c, namespace, podName, &v1.PodExecOptions{
		Container: containerName,
		Command:   args,
		Stdout:    true,
	})
}

func execInPod(c *rest.Config, namespace, podName string, execOpts *v1.PodExecOptions) (remotecommand.Executor, error) {
	coreclient, err := corev1.NewForConfig(c)
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
