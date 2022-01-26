package wharfyml

import (
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/docker"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/git"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/kaniko"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/kubeapply"
	kubecore "k8s.io/api/core/v1"
)

const (
	ConfigVolumeName         = "config-volume"
	ConfigVolumeMountPath    = "/iverCerts"
	GitlabRegistryVolumeName = "gitlab-registry"
	GitlabRegistrySecretName = "gitlab-registry"
	GitlabRegistryMountPath  = "/root/.docker"
)

func BuildPodSpec(
	flags containercreator.PodContainersFlags,
	gitParams map[git.EnvVar]string,
	builtinVars map[containercreator.BuiltinVar]string,
	stepName string,
	variables map[string]interface{}) kubecore.PodSpec {

	podVolumes := []kubecore.Volume{GetConfigVolume(), GetGitLabRegistryVolume()}

	initContainers, initVolumes := getInitContainers(flags, gitParams)
	for _, v := range initVolumes {
		podVolumes = append(podVolumes, v)
	}

	containers, volumes := getContainers(flags, gitParams, builtinVars, stepName, variables)
	for _, v := range volumes {
		podVolumes = append(podVolumes, v)
	}

	return kubecore.PodSpec{
		RestartPolicy: kubecore.RestartPolicyNever,
		NodeSelector: map[string]string{
			"kubernetes.io/os": "linux",
		},
		InitContainers: initContainers,
		Containers:     containers,
		Volumes:        podVolumes,
	}
}

func getInitContainers(flags containercreator.PodContainersFlags, gitParams map[git.EnvVar]string) ([]kubecore.Container, []kubecore.Volume) {
	var initContainers []kubecore.Container
	var volumes []kubecore.Volume
	if flags.GetInitContainersCount() == 0 {
		return initContainers, volumes
	}

	if flags.HasInitContainer(containercreator.Git) {
		gitCloner := GetDefaultContainerSpec(git.NewContainerCreator(git.GetImage(git.DefaultVersion), gitParams, ConfigVolumeMountPath))
		initContainers = append(initContainers, gitCloner)
		volumes = append(volumes, git.GetVolume())
	}

	return initContainers, volumes
}

func getContainers(flags containercreator.PodContainersFlags,
	gitParams map[git.EnvVar]string,
	builtinVars map[containercreator.BuiltinVar]string,
	stepName string,
	variables map[string]interface{}) ([]kubecore.Container, []kubecore.Volume) {

	var containers []kubecore.Container
	var volumes []kubecore.Volume
	if flags.GetContainersCount() == 0 {
		return containers, volumes
	}

	gitVolumeMount := flags.HasInitContainer(containercreator.Git)

	gitVolume := kubecore.VolumeMount{
		Name:      git.VolumeName,
		MountPath: git.RepoDest,
	}

	gitLabVolume := kubecore.VolumeMount{
		Name:      GitlabRegistryVolumeName,
		MountPath: GitlabRegistryMountPath,
	}

	if flags.HasContainer(containercreator.Container) {
	}

	if flags.HasContainer(containercreator.Docker) {
	}

	if flags.HasContainer(containercreator.Kaniko) {
		containerVolumes := []kubecore.VolumeMount{gitLabVolume}
		if gitVolumeMount {
			containerVolumes = append(containerVolumes, gitVolume)
		}

		dockerKaniko := GetDefaultContainerSpec(
			docker.NewContainerCreator(
				kaniko.ContainerName,
				kaniko.GetImage(kaniko.DefaultVersion),
				containerVolumes,
				GitlabRegistryMountPath,
				variables,
				stepName,
				GitlabRegistrySecretName,
				gitParams,
				builtinVars))
		containers = append(containers, dockerKaniko)
	}

	if flags.HasContainer(containercreator.Helm) {
	}

	if flags.HasContainer(containercreator.KubeApply) {
		kubedeploy := GetDefaultContainerSpec(
			kubeapply.NewContainerCreator(
				kubeapply.GetImage(kubeapply.DefaultVersion),
				variables,
				stepName,
				gitParams))
		containers = append(containers, kubedeploy)
		volumes = append(volumes, kubeapply.GetVolume(kubeapply.Variables(variables).GetVariableOrDefault(kubeapply.Cluster)))
	}

	return containers, volumes
}

func GetConfigVolume() kubecore.Volume {
	return kubecore.Volume{
		Name: ConfigVolumeName,
		VolumeSource: kubecore.VolumeSource{
			ConfigMap: &kubecore.ConfigMapVolumeSource{
				LocalObjectReference: kubecore.LocalObjectReference{Name: "ca-certificates-config"},
				Items:                nil,
				DefaultMode:          nil,
				Optional:             nil,
			},
		},
	}
}

func GetGitLabRegistryVolume() kubecore.Volume {
	return kubecore.Volume{
		Name: GitlabRegistryVolumeName,
		VolumeSource: kubecore.VolumeSource{
			Secret: &kubecore.SecretVolumeSource{
				SecretName: GitlabRegistrySecretName,
				Items: []kubecore.KeyToPath{{
					Key:  ".dockerconfigjson",
					Path: "config.json",
				}},
			},
		},
	}
}
