package kubernetes

import (
	"fmt"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
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
}

func NewPodClient(namespace string, kubeconfig *rest.Config) (PodClient, error) {
	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		log.WithError(err).Fatalln("Error loading config")
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}

	podInterface := clientset.CoreV1().Pods(namespace)

	return &podClient{
		namespace:    namespace,
		podInterface: podInterface,
		logsReader:   NewContainerLogsReader(podInterface),
	}, nil
}

func (c *podClient) CreatePod(step wharfyml.Step, stage wharfyml.Stage, buildID int) (*kubecore.Pod, error) {
	pod, err := c.getPodDefinition(step, stage, buildID)
	if err != nil {
		log.WithError(err).Errorln("Failed to get pod definition")
		return nil, fmt.Errorf("run definition: create pod definition: %w", err)
	}

	created, err := c.podInterface.Create(&pod)
	if err != nil {
		log.WithError(err).Fatalln("failed to create pod")
		return nil, fmt.Errorf("run definition: create pod: %w", err)
	}

	return created, nil
}

func (c *podClient) DeletePod(podName string) error {
	err := c.podInterface.Delete(podName, nil)
	if err != nil {
		log.WithError(err).WithField("name", podName).Errorln("Unable to delete pod")
		return err
	}

	return nil
}

func (c *podClient) getPodDefinition(step wharfyml.Step, stage wharfyml.Stage, buildID int) (kubecore.Pod, error) {
	log.WithFields(log.Fields{
		"namespace": c.namespace,
		"step":      step,
		"stage":     stage,
		"buildID":   buildID,
	}).Traceln()

	spec, err := getPodSpec(step)
	if err != nil {
		log.WithError(err).Errorln("Could not get pod spec")
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

func getPodSpec(step wharfyml.Step) (kubecore.PodSpec, error) {
	image, err := step.GetImage()
	if err != nil {
		log.WithError(err).Errorln("Could not get image")
		return kubecore.PodSpec{}, err
	}

	command, err := step.GetCommand()
	if err != nil {
		log.WithError(err).Errorln("Could not get command")
		return kubecore.PodSpec{}, err
	}

	return kubecore.PodSpec{
		Containers: []kubecore.Container{
			{
				Name:            "wharf",
				Image:           image,
				ImagePullPolicy: kubecore.PullIfNotPresent,
				Command:         command,
			},
		},
	}, nil
}

func (c *podClient) ReadLogsFromPod(podLogsChannel chan<- string, podName string) error {
	if podLogsChannel == nil {
		log.Errorln("Uninitialized channel")
		return fmt.Errorf("uninitialized channel")
	}

	defer close(podLogsChannel)

	crw, err := NewContainerReadyWaiter(c.podInterface, podName)
	if err != nil {
		log.WithError(err).Errorln("Unable to create container ready waiter.")
		return err
	}

	var errStrings []string
	var mutex sync.Mutex
	wg := sync.WaitGroup{}
	for crw.AnyRemaining() {
		container, err := crw.WaitNext()
		if err != nil {
			log.WithError(err).Errorln("Unable to get next container.")
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
		log.WithError(err).Errorln("Unable to create container ready waiter.")
		return false, err
	}

	for cdw.AnyRemaining() {
		container, err := cdw.WaitNext()
		if err != nil {
			log.WithError(err).Errorln("Unable to get next done container.")
			return false, err
		}

		log.WithField("container", container).
			Traceln(fmt.Sprintf("Container %q in pod %q has reached done state.", container.Name, container.PodName))
	}

	return true, nil
}