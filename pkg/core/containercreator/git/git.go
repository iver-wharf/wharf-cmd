package git

import (
	"fmt"

	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator"
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator/utils"
	kubecore "k8s.io/api/core/v1"
)

const (
	ImageName      = "alpine/git"
	DefaultVersion = "v2.30.1"
	ContainerName  = "git-cloner"
	VolumeName     = "git-repo"
	certDest       = "/etc/ssl/certs/"
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
}

func NewContainerCreator(imageName string, envs map[EnvVar]string, certPath string) containercreator.ContainerCreator {
	return containerCreator{
		containerName: ContainerName,
		imageName:     imageName,
		envVars:       buildEnvsSlice(envs),
		volumes: []kubecore.VolumeMount{
			{
				Name:      VolumeName,
				MountPath: RepoDest,
			},
			{
				Name:      "config-volume",
				MountPath: certPath,
			},
		},
		iverCertPath: certPath,
	}
}

func buildEnvsSlice(values map[EnvVar]string) []kubecore.EnvVar {
	var envs []kubecore.EnvVar
	for k, v := range values {
		envs = append(envs, getEnvVar(k, v))
	}
	return envs
}

func getEnvVar(name EnvVar, value string) kubecore.EnvVar {
	return kubecore.EnvVar{
		Name:      name.String(),
		Value:     value,
		ValueFrom: nil,
	}
}

func GetVolume() kubecore.Volume {
	return kubecore.Volume{
		Name: VolumeName,
		VolumeSource: kubecore.VolumeSource{
			EmptyDir: &kubecore.EmptyDirVolumeSource{
				Medium:    "",
				SizeLimit: nil,
			},
		},
	}
}

func (c containerCreator) GetContainerName() string {
	return c.containerName
}

func (c containerCreator) GetImageName() string {
	return c.imageName
}

func (c containerCreator) GetShell() string {
	return "sh"
}

func (c containerCreator) GetCommands() []string {
	copyCertsCommand := fmt.Sprintf("cp %s/* %s", c.iverCertPath, certDest)
	gitConfig := "git config --global http.sslVerify false"
	gitClone := fmt.Sprintf("git clone -b $%s -- $%s %s", SyncBranch, RepoURL, RepoDest)
	return []string{fmt.Sprintf("%s; %s; %s", copyCertsCommand, gitConfig, gitClone)}
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
