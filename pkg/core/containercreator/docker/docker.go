package docker

import (
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/git"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/utils"
	kubecore "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	ImageName      = "docker"
	DefaultVersion = "18.09"
)

func GetImageLatest() string {
	return GetImage("latest")
}

func GetImage(version string) string {
	return utils.GetImage(ImageName, ':', version)
}

type containerCreator struct {
	containerName string
	imageName     string
	envVars       []kubecore.EnvVar
	volumes       []kubecore.VolumeMount
	iverCertPath  string
	variables     Variables
	stageName     string
	secret        string
	repo          map[git.EnvVar]string
	builtinVars   map[containercreator.BuiltinVar]string
}

func NewContainerCreator(
	containerName string,
	imageName string,
	volumeMount []kubecore.VolumeMount,
	certPath string,
	variables Variables,
	stageName string,
	secret string,
	repo map[git.EnvVar]string,
	builtinVars map[containercreator.BuiltinVar]string) containercreator.ContainerCreator {

	return containerCreator{
		containerName: containerName,
		imageName:     imageName,
		envVars:       []kubecore.EnvVar{},
		volumes:       volumeMount,
		iverCertPath:  certPath,
		variables:     variables,
		stageName:     stageName,
		secret:        secret,
		repo:          repo,
		builtinVars:   builtinVars,
	}
}

func (c containerCreator) GetContainerName() string {
	return c.containerName
}

func (c containerCreator) GetImageName() string {
	return c.imageName
}

func (c containerCreator) GetShell() string {
	return "/busybox/sh"
}

func (c containerCreator) GetCommands() []string {
	return []string{c.variables.GetScript(c.stageName, c.secret, c.builtinVars)}
}

func (c containerCreator) GetLifeCycle() *kubecore.Lifecycle {
	return nil
}

func (c containerCreator) GetResources() kubecore.ResourceRequirements {
	return kubecore.ResourceRequirements{
		Limits: kubecore.ResourceList{
			kubecore.ResourceCPU:    resource.MustParse("4000m"),
			kubecore.ResourceMemory: resource.MustParse("12Gi"),
		},
		Requests: kubecore.ResourceList{
			kubecore.ResourceCPU:    resource.MustParse("128m"),
			kubecore.ResourceMemory: resource.MustParse("512Mi"),
		},
	}
}

func (c containerCreator) GetEnvVars() []kubecore.EnvVar {
	return c.envVars
}

func (c containerCreator) GetVolumes() []kubecore.VolumeMount {
	return c.volumes
}
