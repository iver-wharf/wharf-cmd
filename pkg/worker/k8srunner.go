package worker

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/iver-wharf/wharf-cmd/internal/filecopy"
	"github.com/iver-wharf/wharf-cmd/internal/gitutil"
	"github.com/iver-wharf/wharf-cmd/internal/ignorer"
	"github.com/iver-wharf/wharf-cmd/pkg/config"
	"github.com/iver-wharf/wharf-cmd/pkg/resultstore"
	"github.com/iver-wharf/wharf-cmd/pkg/tarstore"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/worker/workermodel"
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

var (
	podInitWaitArgs     = []string{"/bin/sh", "-c", "sleep infinite || true"}
	podInitContinueArgs = []string{"killall", "-s", "SIGINT", "sleep"}

	errParentDirAccessNotAllowed = errors.New("accessing parent directory of repo root is not allowed")
)

// K8sRunnerOptions is a struct of options for a Kubernetes step runner.
type K8sRunnerOptions struct {
	BuildOptions
	Config        *config.Config
	RestConfig    *rest.Config
	ResultStore   resultstore.Store
	TarStore      tarstore.Store
	VarSource     varsub.Source
	SkipGitIgnore bool
	CurrentDir    string
}

// NewK8s is a helper function that creates a new builder using the
// NewK8sStepRunnerFactory.
func NewK8s(ctx context.Context, def wharfyml.Definition, opts K8sRunnerOptions) (Builder, error) {
	stageFactory, err := NewK8sStageRunnerFactory(opts)
	if err != nil {
		return nil, err
	}
	return New(ctx, stageFactory, def, opts.BuildOptions)
}

// NewK8sStageRunnerFactory is a helper function that creates a new stage runner
// factory using the NewK8sStepRunnerFactory.
func NewK8sStageRunnerFactory(opts K8sRunnerOptions) (StageRunnerFactory, error) {
	stepFactory, err := NewK8sStepRunnerFactory(opts)
	if err != nil {
		return nil, err
	}
	return NewStageRunnerFactory(stepFactory)
}

// NewK8sStepRunnerFactory returns a new step runner factory that creates
// step runners with implementation that targets Kubernetes using a specific
// Kubernetes namespace and REST config.
func NewK8sStepRunnerFactory(opts K8sRunnerOptions) (StepRunnerFactory, error) {
	clientset, err := kubernetes.NewForConfig(opts.RestConfig)
	if err != nil {
		return nil, err
	}
	factory := k8sStepRunnerFactory{
		K8sRunnerOptions: opts,
		clientset:        clientset,
	}
	return factory, nil
}

type k8sStepRunnerFactory struct {
	K8sRunnerOptions
	clientset *kubernetes.Clientset
}

func (f k8sStepRunnerFactory) NewStepRunner(
	ctx context.Context, step wharfyml.Step, stepID uint64) (StepRunner, error) {
	ctx = contextWithStepName(ctx, step.Name)
	pod, err := f.getStepPodSpec(ctx, step)
	if err != nil {
		return nil, err
	}

	tarball, err := f.prepareStepRepo(step, stepID)
	if err != nil {
		return nil, err
	}

	r := k8sStepRunner{
		K8sRunnerOptions: f.K8sRunnerOptions,
		log:              logger.NewScoped(contextStageStepName(ctx)),
		step:             step,
		pod:              &pod,
		clientset:        f.clientset,
		pods:             f.clientset.CoreV1().Pods(f.Config.K8s.Namespace),
		stepID:           stepID,
		repoTar:          tarball,
		target: &target{
			namespace: f.Config.K8s.Namespace,
			name:      "",
			container: "init",
		},
	}
	if err := r.dryRunStepError(ctx); err != nil {
		return nil, fmt.Errorf("dry-run: %w", err)
	}
	return r, nil
}

func (f k8sStepRunnerFactory) prepareStepRepo(step wharfyml.Step, stepID uint64) (tarstore.Tarball, error) {
	onlyFiles, hasFileFilter := getOnlyFilesToTransfer(step)
	copier := f.getStepRepoCopier(hasFileFilter)
	ignorer, err := f.getStepRepoIgnorer(onlyFiles, hasFileFilter)
	if err != nil {
		return "", err
	}
	tarID := f.getStepTarID(stepID, hasFileFilter)

	tarball, err := f.TarStore.GetPreparedTarball(copier, ignorer, tarID)
	if err != nil {
		return "", err
	}
	return tarball, nil
}

func (f k8sStepRunnerFactory) getStepTarID(stepID uint64, hasFileFilter bool) string {
	if hasFileFilter {
		return fmt.Sprintf("step-%d", stepID)
	}
	return "full"
}

func (f k8sStepRunnerFactory) getStepRepoCopier(hasFileFilter bool) filecopy.Copier {
	if hasFileFilter {
		return varsub.NewCopier(f.VarSource)
	}
	return filecopy.IOCopier
}

func (f k8sStepRunnerFactory) getStepRepoIgnorer(onlyFiles []string, hasFileFilter bool) (ignorer.Ignorer, error) {
	var igns []ignorer.Ignorer
	if hasFileFilter {
		igns = append(igns, ignorer.NewFileIncluder(onlyFiles))
	}

	if !f.SkipGitIgnore {
		repoRoot, err := gitutil.GitRepoRoot(f.CurrentDir)
		if err != nil {
			return nil, err
		}
		gitIgn, err := gitutil.NewIgnorer(f.CurrentDir, repoRoot)
		if err != nil {
			return nil, err
		}
		igns = append(igns, gitIgn)
	}

	if len(igns) == 0 {
		return nil, nil
	}
	return ignorer.Merge(igns...), nil
}

type k8sStepRunner struct {
	K8sRunnerOptions
	log       logger.Logger
	step      wharfyml.Step
	pod       *v1.Pod
	clientset *kubernetes.Clientset
	pods      corev1.PodInterface
	stepID    uint64
	repoTar   tarstore.Tarball
	target    *target
}

type target struct {
	namespace string
	name      string
	container string
}

func (r k8sStepRunner) Step() wharfyml.Step {
	return r.step
}

func (r k8sStepRunner) RunStep(ctx context.Context) StepResult {
	ctx = contextWithStepName(ctx, r.step.Name)
	start := time.Now()
	status := workermodel.StatusSuccess
	err := r.runStepError(ctx)
	if errors.Is(ctx.Err(), context.Canceled) {
		status = workermodel.StatusCancelled
	} else if err != nil {
		status = workermodel.StatusFailed
	}
	r.addStatusUpdate(status)
	return StepResult{
		Name:     r.step.Name,
		Status:   status,
		Type:     r.step.Type.StepTypeName(),
		Error:    err,
		Duration: time.Since(start),
	}
}

func (r k8sStepRunner) dryRunStepError(ctx context.Context) error {
	log.Debug().
		WithString("step", r.step.Name).
		WithString("pod", r.pod.GenerateName).
		Message("DRY RUN: Creating pod.")
	newPod, err := r.pods.Create(ctx, r.pod, metav1.CreateOptions{
		DryRun: []string{"All"},
	})
	if err != nil {
		return fmt.Errorf("create pod: %w", err)
	}
	log.Debug().
		WithString("step", r.step.Name).
		WithString("pod", newPod.Name).
		Message("DRY RUN: Created pod.")
	return nil
}

func (r k8sStepRunner) runStepError(ctx context.Context) error {
	log.Debug().
		WithString("step", r.step.Name).
		WithString("pod", r.pod.GenerateName).
		Message("Creating pod.")
	r.addStatusUpdate(workermodel.StatusInitializing)
	newPod, err := r.pods.Create(ctx, r.pod, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create pod: %w", err)
	}
	r.target.name = newPod.Name
	var logFunc = func(ev logger.Event) logger.Event {
		return ev.
			WithString("step", r.step.Name).
			WithString("pod", newPod.Name)
	}

	log.Debug().WithFunc(logFunc).Message("Created pod.")
	defer r.stopPodNow(context.Background())
	log.Debug().WithFunc(logFunc).Message("Waiting for init container to start.")
	if err := r.waitForInitContainerRunning(ctx, newPod.ObjectMeta); err != nil {
		return fmt.Errorf("wait for init container: %w", err)
	}

	log.Debug().WithFunc(logFunc).Message("Transferring data to pod.")
	if err := r.transferDataToPod(ctx); err != nil {
		return err
	}
	log.Debug().WithFunc(logFunc).Message("Transferred data to pod.")

	if err := r.continueInitContainer(); err != nil {
		return fmt.Errorf("continue init container: %w", err)
	}
	r.addStatusUpdate(workermodel.StatusRunning)

	log.Debug().WithFunc(logFunc).Message("Waiting for app container to start.")
	if err := r.waitForAppContainerRunningOrDone(ctx, newPod.ObjectMeta); err != nil {
		if err := r.readLogs(ctx, &v1.PodLogOptions{}); err != nil {
			log.Debug().WithError(err).
				Message("Failed to read logs from failed container.")
		}
		return fmt.Errorf("wait for app container: %w", err)
	}
	log.Debug().WithFunc(logFunc).Message("App container running. Streaming logs.")
	if err := r.readLogs(ctx, &v1.PodLogOptions{Follow: true, Timestamps: true}); err != nil {
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
		switch obj := ev.Object.(type) {
		case *v1.Pod:
			switch ev.Type {
			case watch.Modified:
				ok, err := f(obj)
				if err != nil {
					return err
				} else if ok {
					return nil
				}
			case watch.Deleted:
				return fmt.Errorf("pod was removed: %v", obj.Name)
			}
		case *metav1.Status:
			if errors.Is(ctx.Err(), context.Canceled) {
				return fmt.Errorf("watching pod: %s: %w", podMeta.Name, ctx.Err())
			}

			return fmt.Errorf("error for pod: %v: %v", podMeta.Name, errors.New(obj.Message))
		}
	}
	return fmt.Errorf("got no more events when watching pod: %v", podMeta.Name)
}

func (r k8sStepRunner) readLogs(ctx context.Context, opts *v1.PodLogOptions) error {
	req := r.pods.GetLogs(r.target.name, opts)
	readCloser, err := req.Stream(ctx)
	if err != nil {
		return err
	}
	defer readCloser.Close()
	scanner := bufio.NewScanner(readCloser)
	writer, err := r.ResultStore.OpenLogWriter(uint64(r.stepID))
	if err != nil {
		r.log.Error().WithError(err).Message("Failed to open log writer. No logs will be written.")
	} else {
		defer func() {
			if err := writer.Close(); err != nil {
				r.log.Error().WithError(err).Message("Failed to close log writer.")
			}
		}()
	}
	for scanner.Scan() {
		txt := scanner.Text()
		idx := strings.LastIndexByte(txt, '\r')
		if idx != -1 {
			txt = txt[idx+1:]
		}
		r.log.Info().Message(txt)
		if writer != nil {
			if err := writer.WriteLogLine(txt); err != nil {
				r.log.Error().WithError(err).Message("Failed to write log line.")
			}
		}
	}
	return scanner.Err()
}

func (r k8sStepRunner) stopPodNow(ctx context.Context) {
	gracePeriod := int64(0) // 0=immediately
	err := r.pods.Delete(ctx, r.target.name, metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	})
	if err != nil {
		log.Warn().
			WithError(err).
			WithString("step", r.step.Name).
			WithString("pod", r.target.name).
			Message("Failed to delete pod.")
	} else {
		log.Debug().
			WithString("step", r.step.Name).
			WithString("pod", r.target.name).
			Message("Deleted pod.")
	}
}

func (r k8sStepRunner) transferDataToPod(ctx context.Context) error {
	var logFunc = func(ev logger.Event) logger.Event {
		return ev.
			WithString("step", r.step.Name).
			WithString("pod", r.target.name)
	}

	log.Debug().WithFunc(logFunc).Message("Transferring repo to init container.")
	if err := r.copyDirToPod(ctx, commonRepoVolumeMount.MountPath); err != nil {
		return fmt.Errorf("transfer repo: %w", err)
	}
	log.Debug().WithFunc(logFunc).Message("Transferred repo to init container.")

	if step, ok := r.step.Type.(wharfyml.StepDocker); ok {
		{
			log.Debug().WithFunc(logFunc).Message("Transferring modified Dockerfile to init container.")
			dockerfilePath := filepath.Join(r.CurrentDir, step.File)
			if !validateNoParentDirAccess(dockerfilePath) {
				return fmt.Errorf("transfer modified dockerfile: %w: %q", errParentDirAccessNotAllowed, dockerfilePath)
			}
			if _, err := os.Stat(dockerfilePath); err != nil {
				return fmt.Errorf("transfer modified dockerfile: %w", err)
			}
			if err := r.copyDockerfileToPod(ctx, step); err != nil {
				return fmt.Errorf("transfer modified dockerfile: %w", err)
			}
			log.Debug().WithFunc(logFunc).Message("Transferred modified Dockerfile to init container.")
		}

		{
			log.Debug().WithFunc(logFunc).Message("Transferring cert file to init container.")
			certFile, err := os.Open("/mnt/cert/root.crt")
			if err != nil {
				return fmt.Errorf("transfer cert: %w", err)
			}
			destPath := filepath.Join(commonRepoVolumeMount.MountPath, step.Context, "root.crt")
			if !validateNoParentDirAccess(destPath) {
				return fmt.Errorf("transfer cert: %w: %q", errParentDirAccessNotAllowed, destPath)
			}
			args := []string{"tee", destPath}
			if err := r.copyToPodStdin(ctx, certFile, args); err != nil {
				return fmt.Errorf("transfer cert: %w", err)
			}
			log.Debug().WithFunc(logFunc).Message("Transferred cert file to init container.")
		}
	}
	return nil
}

func (r k8sStepRunner) continueInitContainer() error {
	exec, err := execInPodPipeStdout(r.RestConfig, r.target, podInitContinueArgs)
	if err != nil {
		return err
	}
	exec.Stream(remotecommand.StreamOptions{
		Stdout: nopWriter{},
	})
	return nil
}

func (r k8sStepRunner) copyDirToPod(ctx context.Context, destPath string) error {
	tarReader, err := r.repoTar.Open()
	if err != nil {
		return err
	}
	defer tarReader.Close()

	args := []string{"tar", "-xf", "-", "-C", destPath}
	return r.copyToPodStdin(ctx, tarReader, args)
}

func (r k8sStepRunner) copyDockerfileToPod(ctx context.Context, step wharfyml.StepDocker) error {
	path := filepath.Join(r.CurrentDir, step.File)
	if !validateNoParentDirAccess(path) {
		return fmt.Errorf("%w: %q", errParentDirAccessNotAllowed, path)
	}

	destPath := filepath.Join(commonRepoVolumeMount.MountPath, step.File)
	if !validateNoParentDirAccess(destPath) {
		return fmt.Errorf("%w: %q", errParentDirAccessNotAllowed, destPath)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	b = append(b, []byte(`
COPY ./root.crt /usr/local/share/ca-certificates/root.crt
RUN mkdir -p /etc/ssl/certs/ \
	&& touch /etc/ssl/certs/ca-certificates.crt \
	&& cat /usr/local/share/ca-certificates/root.crt >> /etc/ssl/certs/ca-certificates.crt`)...)
	reader := bytes.NewReader(b)
	args := []string{"tee"}
	return r.copyToPodStdin(ctx, reader, args)
}

func (r k8sStepRunner) copyToPodStdin(ctx context.Context, reader io.Reader, args []string) error {
	// Based on: https://stackoverflow.com/a/57952887
	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()
	defer pipeWriter.Close()
	exec, err := execInPodPipedStdin(r.RestConfig, r.target, args)
	if err != nil {
		return err
	}
	writeErrCh := make(chan error, 1)
	go func() {
		defer pipeWriter.Close()
		_, err := io.Copy(pipeWriter, reader)
		writeErrCh <- err
	}()
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin: pipeReader,
	})
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return errors.New("aborted")
	case err := <-writeErrCh:
		return err
	}
}

func execInPodPipedStdin(c *rest.Config, t *target, args []string) (remotecommand.Executor, error) {
	return execInPod(c, t.namespace, t.name, &v1.PodExecOptions{
		Container: t.container,
		Command:   args,
		Stdin:     true,
	})
}

func execInPodPipeStdout(c *rest.Config, t *target, args []string) (remotecommand.Executor, error) {
	return execInPod(c, t.namespace, t.name, &v1.PodExecOptions{
		Container: t.container,
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

func (r *k8sStepRunner) addStatusUpdate(status workermodel.Status) {
	r.ResultStore.AddStatusUpdate(r.stepID, time.Now(), status)
}

func validateNoParentDirAccess(p string) bool {
	parts := strings.Split(p, string(filepath.Separator))
	level := 0
	for _, v := range parts {
		if v == ".." {
			level--
		} else {
			level++
		}
		if level < 0 {
			return false
		}
	}
	return true
}
