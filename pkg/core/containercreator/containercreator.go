package containercreator

import kubecore "k8s.io/api/core/v1"

type ContainerCreator interface {
	GetContainerName() string
	GetImageName() string
	GetShell() string
	GetCommands() []string
	GetEnvVars() []kubecore.EnvVar
	GetLifeCycle() *kubecore.Lifecycle
	GetResources() kubecore.ResourceRequirements
	GetVolumes() []kubecore.VolumeMount
}
