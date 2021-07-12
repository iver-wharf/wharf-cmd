package run

import (
	"fmt"
	"sync"

	"github.com/iver-wharf/wharf-api-client-go/pkg/wharfapi"
	"github.com/iver-wharf/wharf-cmd/pkg/core/buildclient"
	"github.com/iver-wharf/wharf-cmd/pkg/core/kubernetes"
	"github.com/iver-wharf/wharf-cmd/pkg/core/wharfyml"
	"github.com/iver-wharf/wharf-core/pkg/logger"
	"k8s.io/client-go/rest"
)

var log = logger.New()

type Runner struct {
	Kubeconfig  *rest.Config
	buildClient buildclient.Client
	AuthHeader  string
}

func NewRunner(kubeconfig *rest.Config, authHeader string) Runner {
	return Runner{
		Kubeconfig:  kubeconfig,
		buildClient: buildclient.New(authHeader),
		AuthHeader:  authHeader,
	}
}

func (r Runner) Run(path, environment, namespace, stageName string, buildID int, builtinVars map[wharfyml.BuiltinVar]string) error {
	var withRunMeta = func(ev logger.Event) logger.Event {
		return ev.
			WithString("path", path).
			WithString("namespace", namespace).
			WithString("stage", stageName).
			WithInt("buildID", buildID)
	}
	withRunMeta(log.Debug()).
		Message("Run called.")

	def, err := wharfyml.Parse(path, builtinVars)
	if err != nil {
		withRunMeta(log.Error()).
			WithError(err).
			Message("Failed to parse wharf-ci file.")
		return fmt.Errorf("run: parse definition: %w", err)
	}

	return r.RunDefinition(def, environment, namespace, stageName, buildID, builtinVars)
}

func (r Runner) RunDefinition(
	definition wharfyml.BuildDefinition,
	environment, namespace, stageName string,
	buildID int,
	builtinVars map[wharfyml.BuiltinVar]string) error {
	var withRunMeta = func(ev logger.Event) logger.Event {
		return ev.
			WithString("namespace", namespace).
			WithString("stage", stageName).
			WithInt("buildID", buildID)
	}
	stage, err := definition.GetStageWithReplacement(stageName, environment)
	if err != nil {
		withRunMeta(log.Error()).
			WithError(err).
			Message("Stage not found.")
		return fmt.Errorf("stage %q not found in definition", stageName)
	}

	r.postLogWithStatus(buildID, "Parsed run definition.", wharfapi.BuildScheduling)

	podClient, err := kubernetes.NewPodClient(namespace, r.Kubeconfig)
	if err != nil {
		withRunMeta(log.Error()).
			WithError(err).
			Message("Error getting new pod client")
		return fmt.Errorf("run definition: get new pod client: %w", err)
	}

	wg := sync.WaitGroup{}
	for _, step := range stage.Steps {
		pod, err := podClient.CreatePod(step, stage, buildID)
		if err != nil {
			withRunMeta(log.Error()).
				WithError(err).
				Message("Failed to create pod.")

			r.postLogWithStatus(buildID, "Unable to create pod", wharfapi.BuildFailed)
			return fmt.Errorf("run definition: create pod: %w", err)
		}

		r.postLogWithStatus(buildID, fmt.Sprintf("Pod %s created.", pod.Name), wharfapi.BuildRunning)

		wg.Add(1)
		go func() {
			defer wg.Done()

			logChannel := make(chan string)
			go podClient.ReadLogsFromPod(logChannel, pod.Name)

			for message := range logChannel {
				r.postLog(buildID, message)
			}

			done, err := podClient.WaitUntilPodFinished(pod.Name)
			if err != nil || !done {
				return
			}

			err = podClient.DeletePod(pod.Name)
			if err != nil {
				withRunMeta(log.Error()).
					WithError(err).
					WithString("pod", pod.Name).
					Message("Failed to delete pod.")
			}

			r.postLog(buildID, fmt.Sprintf("Pod %s is deleted.", pod.Name))
		}()
	}

	wg.Wait()
	r.postLogWithStatus(buildID, "Build completed.", wharfapi.BuildCompleted)
	return nil
}

func (r Runner) postLog(buildID int, message string) {
	err := r.buildClient.PostLog(uint(buildID), message)
	if err != nil {
		log.Error().
			WithError(err).
			WithInt("build", buildID).
			Message("Failed to post log.")
	}
}

func (r Runner) postLogWithStatus(buildID int, message string, status wharfapi.BuildStatus) {
	err := r.buildClient.PostLogWithStatus(uint(buildID), message, status)
	if err != nil {
		log.Error().
			WithError(err).
			WithInt("build", buildID).
			WithString("newStatus", status.String()).
			Message("Failed to update build status.")
	}
}
