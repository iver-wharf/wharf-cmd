package wharfyml

import (
	"fmt"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/docker"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/git"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/helm"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/kaniko"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/kubeapply"
	kubecore "k8s.io/api/core/v1"
	"strings"
)

type Step struct {
	Name      string
	Type      StepType
	Variables map[string]interface{}
}

func parseStep(name string, content map[string]interface{}) (Step, error) {
	if len(content) != 1 {
		return Step{}, fmt.Errorf("expected single step-type, got %d", len(content))
	}

	var stepType StepType
	var variables map[string]interface{}
	for k, v := range content {
		stepType = ParseStepType(k)
		variables = v.(map[string]interface{})
	}

	return Step{Name: name, Type: stepType, Variables: variables}, nil
}

func (step Step) GetPodSpec(gitParams map[git.EnvVar]string, builtinVars map[containercreator.BuiltinVar]string) (kubecore.PodSpec, error) {
	var containers containercreator.PodContainersFlags
	switch step.Type {
	case KubeApply:
		containers = containers.AddInitContainer(containercreator.Git)
		containers = containers.AddContainer(containercreator.KubeApply)
		return BuildPodSpec(containers, gitParams, builtinVars, step.Name, step.Variables), nil
	case Docker:
		backend := docker.Variables(step.Variables).GetVariableOrDefault(docker.Backend, step.Name, GitlabRegistrySecretName, builtinVars)
		if backend == kaniko.ContainerName {
			containers = containers.AddInitContainer(containercreator.Git)
			containers = containers.AddContainer(containercreator.Kaniko)
			return BuildPodSpec(containers, gitParams, builtinVars, step.Name, step.Variables), nil
		}
	}

	cmd, err := step.GetCommand()
	return kubecore.PodSpec{
		Containers: []kubecore.Container{
			{
				Name:            "wharf",
				Image:           image,
				ImagePullPolicy: kubecore.PullIfNotPresent,
				Command:         cmd,
			},
		},
	}, err

}

func (step Step) GetImage(builtinVars map[containercreator.BuiltinVar]string) (string, error) {
	switch step.Type {
	case Container:
		return step.Variables[image].(string), nil
	case Kaniko:
		return kaniko.GetImageLatest(), nil
	case HelmPackage:
		return helm.GetImageLatest(), nil
	case HelmDeploy:
		version, ok := step.Variables[helmVersion].(string)
		if !ok {
			version = helm.DefaultVersion
		}
		return helm.GetImage(version), nil
	case Docker:
		backend := docker.Variables(step.Variables).GetVariableOrDefault(docker.Backend, step.Name, GitlabRegistrySecretName, builtinVars)
		if backend == kaniko.ContainerName {
			return kaniko.GetImage(kaniko.DefaultVersion), nil
		}
		return helm.GetImage(docker.DefaultVersion), nil
	case KubeApply:
		return kubeapply.GetImage(docker.DefaultVersion), nil

	}

	return "", fmt.Errorf("cannot translate %s step type to image name", step.Type)
}

func (step Step) GetCommand() ([]string, error) {
	switch step.Type {
	case Container:
		vars := step.Variables[commands].([]interface{})
		var strVars []string
		for _, v := range vars {
			strVars = append(strVars, v.(string))
		}
		return []string{"/bin/sh", "-c", strings.Join(strVars, ";")}, nil
	}

	return []string{}, nil
}
