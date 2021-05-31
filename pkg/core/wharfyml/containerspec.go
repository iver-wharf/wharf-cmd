package wharfyml

import (
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator"
	kubecore "k8s.io/api/core/v1"
)

func GetDefaultContainerSpec(creator containercreator.ContainerCreator) kubecore.Container {
	return kubecore.Container{
		Name:            creator.GetContainerName(),
		Image:           creator.GetImageName(),
		ImagePullPolicy: kubecore.PullIfNotPresent,
		Command:         []string{creator.GetShell()},
		TTY:             true,
		Stdin:           true,
		Args:            getArgs(creator.GetCommands()),
		Lifecycle:       creator.GetLifeCycle(),
		Resources:       creator.GetResources(),
		Env:             creator.GetEnvVars(),
		VolumeMounts:    creator.GetVolumes(),
	}
}

func getArgs(args []string) []string {
	result := []string{"-c"}
	for _, a := range args {
		result = append(result, a)
	}
	return result
}
