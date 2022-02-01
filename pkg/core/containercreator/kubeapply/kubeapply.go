package kubeapply

import (
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/git"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/utils"
	kubecore "k8s.io/api/core/v1"
)

const (
	ImageName      = "wharfse/kubectl"
	DefaultVersion = "v1.18.2"
	VolumeName     = "kubectl-config-volume"
	VolumePath     = "/root/.kube"
	ContainerName  = "kubectl"
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
	repo          map[git.EnvVar]string
}

func GetVolume(configMapName string) kubecore.Volume {
	return kubecore.Volume{
		Name: VolumeName,
		VolumeSource: kubecore.VolumeSource{
			ConfigMap: &kubecore.ConfigMapVolumeSource{
				LocalObjectReference: kubecore.LocalObjectReference{
					Name: configMapName,
				},
			},
		},
	}
}

func NewContainerCreator(
	imageName string,
	variables Variables,
	stageName string,
	repo map[git.EnvVar]string,
) containercreator.ContainerCreator {

	return containerCreator{
		containerName: ContainerName,
		imageName:     imageName,
		envVars:       []kubecore.EnvVar{},
		volumes: []kubecore.VolumeMount{
			{
				Name:      VolumeName,
				MountPath: VolumePath,
			},
		},
		variables: variables,
		stageName: stageName,
		repo:      repo,
	}
}

func (c containerCreator) GetContainerName() string {
	return c.containerName
}

func (c containerCreator) GetImageName() string {
	return c.imageName
}

func (c containerCreator) GetShell() string {
	return "/bin/sh"
}

func (c containerCreator) GetCommands() []string {
	return []string{}
}

func (c containerCreator) GetLifeCycle() *kubecore.Lifecycle {
	return nil
}

func (c containerCreator) GetResources() kubecore.ResourceRequirements {
	return kubecore.ResourceRequirements{}
}

func (c containerCreator) GetEnvVars() []kubecore.EnvVar {
	return c.envVars
}

func (c containerCreator) GetVolumes() []kubecore.VolumeMount {
	return c.volumes
}
