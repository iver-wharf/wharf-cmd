package kubernetes

import (
	"fmt"
	"strings"
	"sync"

	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator"

	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/git"
	"github.com/iver-wharf/wharf-cmd/pkg/core/utils"
	"github.com/iver-wharf/wharf-cmd/pkg/core/wharfyml"
	kubecore "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type PodClient interface {
	CreatePod(step wharfyml.Step, stage wharfyml.Stage, buildID int) (*kubecore.Pod, error)
	ReadLogsFromPod(podLogsChannel chan<- string, podName string) error
	WaitUntilPodFinished(podName string) (bool, error)
	DeletePod(podName string) error
}

type podClient struct {
	namespace    string
	podInterface corev1.PodInterface
	logsReader   ContainerLogsReader
	gitParams    map[git.EnvVar]string
	builtinVars  map[containercreator.BuiltinVar]string
}

func NewPodClient(namespace string, kubeconfig *rest.Config, gitParams map[git.EnvVar]string, builtinVars map[containercreator.BuiltinVar]string) (PodClient, error) {
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		log.Panic().WithError(err).Message("Failed loading kube-config.")
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}

	podInterface := clientset.CoreV1().Pods(namespace)

	return &podClient{
		namespace:    namespace,
		podInterface: podInterface,
		logsReader:   NewContainerLogsReader(podInterface),
		gitParams:    gitParams,
		builtinVars:  builtinVars,
	}, nil
}

func (c *podClient) CreatePod(step wharfyml.Step, stage wharfyml.Stage, buildID int) (*kubecore.Pod, error) {
	pod, err := c.getPodDefinition(step, stage, buildID)
	if err != nil {
		log.Error().WithError(err).Message("Failed to get pod definition.")
		return nil, fmt.Errorf("run definition: create pod definition: %w", err)
	}

	created, err := c.podInterface.Create(&pod)
	if err != nil {
		log.Panic().WithError(err).Message("Failed to create pod.")
		return nil, fmt.Errorf("run definition: create pod: %w", err)
	}

	return created, nil
}

func (c *podClient) DeletePod(podName string) error {
	err := c.podInterface.Delete(podName, nil)
	if err != nil {
		log.Error().WithError(err).WithString("pod", podName).Message("Failed to delete pod.")
		return err
	}

	return nil
}

func (c *podClient) getPodDefinition(step wharfyml.Step, stage wharfyml.Stage, buildID int) (kubecore.Pod, error) {
	log.Debug().
		WithString("namespace", c.namespace).
		WithString("step", step.Name).
		WithString("stage", stage.Name).
		WithInt("build", buildID).
		Message("")

	spec, err := step.GetPodSpec(c.gitParams, c.builtinVars)
	if err != nil {
		log.Error().WithError(err).Message("Failed to get pod spec.")
		return kubecore.Pod{}, err
	}

	podName := fmt.Sprintf("%s-%s-%d", stage.Name, step.Name, buildID)
	return kubecore.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: c.namespace,
			Labels: map[string]string{
				"app":   "wharf",
				"step":  step.Name,
				"stage": stage.Name,
			},
		},
		Spec: spec,
	}, nil
}

func (c *podClient) ReadLogsFromPod(podLogsChannel chan<- string, podName string) error {
	if podLogsChannel == nil {
		log.Error().Message("Uninitialized channel.")
		return fmt.Errorf("uninitialized channel")
	}

	defer close(podLogsChannel)

	crw, err := NewContainerReadyWaiter(c.podInterface, podName)
	if err != nil {
		log.Error().WithError(err).Message("Failed to create container ready waiter.")
		return err
	}

	var errStrings []string
	var mutex sync.Mutex
	wg := sync.WaitGroup{}
	for crw.AnyRemaining() {
		container, err := crw.WaitNext()
		if err != nil {
			log.Error().WithError(err).Message("Failed to get next container.")
			return err
		}

		podLogsChannel <- fmt.Sprintf("container %q in pod %q has reached ready state", container.Name, container.PodName)
		wg.Add(1)
		go func() {
			defer wg.Done()

			stream, err := c.logsReader.StreamContainerLogs(podName, container.Name)
			if err != nil {
				mutex.Lock()
				defer mutex.Unlock()
				errStrings = append(errStrings, err.Error())
				return
			}

			scanner := utils.NewStreamScanner(stream, utils.AllSanitizationMethods)
			for scanner.Scan() {
				podLogsChannel <- scanner.Text()
			}

			stream.Close()

			err = scanner.Err()
			if err != nil {
				mutex.Lock()
				defer mutex.Unlock()
				errStrings = append(errStrings, err.Error())
			}
		}()
	}

	wg.Wait()
	if len(errStrings) > 0 {
		return fmt.Errorf("got %d error(s) when reading logs: { %s }",
			len(errStrings), strings.Join(errStrings, "; "))
	}

	return nil
}

func (c *podClient) WaitUntilPodFinished(podName string) (bool, error) {
	cdw, err := NewContainerDoneWaiter(c.podInterface, podName)
	if err != nil {
		log.Error().WithError(err).Message("Failed to create container ready waiter.")
		return false, err
	}

	for cdw.AnyRemaining() {
		container, err := cdw.WaitNext()
		if err != nil {
			log.Error().WithError(err).Message("Failed to get next done container.")
			return false, err
		}

		log.Debug().
			WithString("namespace", container.Namespace).
			WithString("pod", container.PodName).
			WithString("container", container.Name).
			Message("Container has reached done state.")
	}

	return true, nil
}
