package run

import (
	"fmt"
	"sync"

	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/git"

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
	DryRun      bool
}

func NewRunner(kubeconfig *rest.Config, authHeader string) Runner {
	return Runner{
		Kubeconfig:  kubeconfig,
		buildClient: buildclient.New(authHeader),
		AuthHeader:  authHeader,
	}
}

func (r Runner) Run(
	path string,
	environment string,
	namespace string,
	stageName string,
	buildID int,
	gitParams map[git.EnvVar]string,
	builtinVars map[containercreator.BuiltinVar]string) error {
	var withRunMeta = func(ev logger.Event) logger.Event {
		return ev.
			WithString("path", path).
			WithString("namespace", namespace).
			WithString("stage", stageName).
			WithInt("buildID", buildID)
	}
	log.Debug().WithFunc(withRunMeta).
		Message("Run called.")

	def, err := wharfyml.Parse(path, builtinVars)
	if err != nil {
		return fmt.Errorf("run: parse definition: %w", err)
	}

	return r.RunDefinition(def, environment, namespace, stageName, buildID, gitParams, builtinVars)
}

func (r Runner) RunDefinition(
	definition wharfyml.BuildDefinition,
	environment string,
	namespace string,
	stageName string,
	buildID int,
	gitParams map[git.EnvVar]string,
	builtinVars map[containercreator.BuiltinVar]string) error {
	var withRunMeta = func(ev logger.Event) logger.Event {
		return ev.
			WithString("namespace", namespace).
			WithString("stage", stageName).
			WithInt("buildID", buildID)
	}
	stage, err := definition.GetStageWithReplacement(stageName, environment)
	if err != nil {
		log.Error().WithFunc(withRunMeta).
			WithError(err).
			Message("Stage not found.")
		return fmt.Errorf("stage %q not found in definition", stageName)
	}

	// TODO: Remove
	//err = r.buildClient.PostLogWithStatus(uint(buildID), "run definition", wharfapi.BuildScheduling)
	//if err != nil {
	//	log.Error().WithFunc(withRunMeta).
	//		WithError(err).
	//		Message("Unable to update build status.")
	//}

	podClient, err := kubernetes.NewPodClient(namespace, r.Kubeconfig, gitParams, builtinVars)
	if err != nil {
		log.Error().WithFunc(withRunMeta).
			WithError(err).
			Message("Error getting new pod client.")
		return fmt.Errorf("run definition: get new pod client: %w", err)
	}

	wg := sync.WaitGroup{}
	for _, step := range stage.Steps {
		var podName string
		if r.DryRun {
			podName = fmt.Sprintf("wharf-build-%s-%s", stage.Name, step.Name)
			log.Info().WithFunc(withRunMeta).
				WithString("pod", podName).
				Message("Created pod (dry run)")
		} else {
			pod, err := podClient.CreatePod(step, stage, buildID)
			if err != nil {
				log.Error().WithFunc(withRunMeta).
					WithError(err).
					Message("Failed to create pod.")

				// TODO: Remove
				//err = r.buildClient.PostLogWithStatus(uint(buildID), "Unable to create pod.", wharfapi.BuildFailed)
				//if err != nil {
				//	log.Error().WithFunc(withRunMeta).
				//		WithError(err).
				//		Message("Unable to update build status.")
				//}
				return fmt.Errorf("run definition: create pod: %w", err)
			}
			podName = pod.Name
		}

		// TODO: Remove
		//err = r.buildClient.PostLogWithStatus(uint(buildID), fmt.Sprintf("Pod %s created.", pod.Name), wharfapi.BuildRunning)
		//if err != nil {
		//	log.Error().WithFunc(withRunMeta).
		//		WithError(err).
		//		Message("Unable to update build status.")
		//}

		wg.Add(1)
		go func() {
			defer wg.Done()

			logChannel := make(chan string)
			if r.DryRun {
				log.Info().WithFunc(withRunMeta).
					WithString("pod", podName).
					Message("Read logs from pod (dry run)")
				close(logChannel)
			} else {
				go podClient.ReadLogsFromPod(logChannel, podName)
			}

			for message := range logChannel {
				err := r.buildClient.PostLog(uint(buildID), message)
				if err != nil {
					log.Error().WithFunc(withRunMeta).
						WithError(err).
						Message("Unable to post log.")
				}
			}

			if r.DryRun {
				log.Info().WithFunc(withRunMeta).
					WithString("pod", podName).
					Message("Wait until pod has finished (dry run)")
			} else {
				done, err := podClient.WaitUntilPodFinished(podName)
				if err != nil || !done {
					return
				}
			}

			if r.DryRun {
				log.Info().WithFunc(withRunMeta).
					WithString("pod", podName).
					Message("Delete pod (dry run)")
			} else {
				err = podClient.DeletePod(podName)
				if err != nil {
					log.Error().WithFunc(withRunMeta).
						WithError(err).
						WithString("pod", podName).
						Message("Failed to delete pod.")
				}
			}

			// TODO: Remove
			//err = r.buildClient.PostLog(uint(buildID), fmt.Sprintf("Pod %s is deleted.", pod.Name))
			//if err != nil {
			//	log.Error().WithFunc(withRunMeta).
			//		WithError(err).
			//		Message("Failed to post log.")
			//}
		}()
	}

	wg.Wait()
	// TODO: Remove
	//err = r.buildClient.PostLogWithStatus(uint(buildID), "Build completed", wharfapi.BuildCompleted)
	//if err != nil {
	//	log.Error().WithFunc(withRunMeta).
	//		WithError(err).
	//		WithStringer("newStatus", wharfapi.BuildCompleted).
	//		Message("Failed to update build status.")
	//}

	return nil
}
