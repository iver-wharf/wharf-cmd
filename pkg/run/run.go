package run

import (
	"fmt"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/git"
	"sync"

	"github.com/iver-wharf/wharf-api-client-go/pkg/wharfapi"
	log "github.com/sirupsen/logrus"
	"github.com/iver-wharf/wharf-cmd/pkg/core/buildclient"
	"github.com/iver-wharf/wharf-cmd/pkg/core/kubernetes"
	"github.com/iver-wharf/wharf-cmd/pkg/core/wharfyml"
	"k8s.io/client-go/rest"
)

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

func (r Runner) Run(
	path string,
	environment string,
	namespace string,
	stageName string,
	buildID int,
	gitParams map[git.EnvVar]string,
	builtinVars map[containercreator.BuiltinVar]string) error {
	log.WithFields(log.Fields{
		"path":      path,
		"namespace": namespace,
		"stage":     stageName,
		"buildID":   buildID,
	}).Traceln("Run called")

	def, err := wharfyml.Parse(path, builtinVars)
	if err != nil {
		log.WithError(err).Errorln("Failed to parse wharf-ci file!")
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
	stage, err := definition.GetStageWithReplacement(stageName, environment)
	if err != nil {
		log.WithError(err).
			WithFields(log.Fields{"stage": stageName, "environment": environment}).
			Errorln("Stage not found")
		return fmt.Errorf("stage %q not found in definition", stageName)
	}

	err = r.buildClient.PostLogWithStatus(uint(buildID), "run definition", wharfapi.BuildScheduling)
	if err != nil {
		log.WithError(err).Errorln("Unable to update build status")
	}

	podClient, err := kubernetes.NewPodClient(namespace, r.Kubeconfig, gitParams, builtinVars)
	if err != nil {
		log.WithError(err).Fatalln("Error getting new pod client")
		return fmt.Errorf("run definition: get new pod client: %w", err)
	}

	wg := sync.WaitGroup{}
	for _, step := range stage.Steps {
		pod, err := podClient.CreatePod(step, stage, buildID)
		if err != nil {
			log.WithError(err).Fatalln("Failed to create pod")

			err = r.buildClient.PostLogWithStatus(uint(buildID), "Unable to create pod.", wharfapi.BuildFailed)
			if err != nil {
				log.WithError(err).Errorln("Unable to update build status")
			}
			return fmt.Errorf("run definition: create pod: %w", err)
		}

		err = r.buildClient.PostLogWithStatus(uint(buildID), fmt.Sprintf("Pod %s created.", pod.Name), wharfapi.BuildRunning)
		if err != nil {
			log.WithError(err).Errorln("Unable to update build status")
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			logChannel := make(chan string)
			go podClient.ReadLogsFromPod(logChannel, pod.Name)

			for message := range logChannel {
				err := r.buildClient.PostLog(uint(buildID), message)
				if err != nil {
					log.WithError(err).Errorln("Unable to post log")
				}
			}

			done, err := podClient.WaitUntilPodFinished(pod.Name)
			if err != nil || !done {
				return
			}

			err = podClient.DeletePod(pod.Name)
			if err != nil {
				log.WithError(err).WithField("pod name", pod.Name).Errorln("Unable to delete the pod.")
			}

			err = r.buildClient.PostLog(uint(buildID), fmt.Sprintf("Pod %s is deleted.", pod.Name))
			if err != nil {
				log.WithError(err).Errorln("Unable to post build log")
			}
		}()
	}

	wg.Wait()
	err = r.buildClient.PostLogWithStatus(uint(buildID), "Build completed", wharfapi.BuildCompleted)
	if err != nil {
		log.WithError(err).Errorln("Unable to update build status")
	}
	return nil
}
