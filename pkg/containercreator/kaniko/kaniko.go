package kaniko

import (
	"github.com/iver-wharf/wharf-cmd/pkg/containercreator/utils"
)

const (
	ImageName      = "boolman/kaniko:busybox"
	DefaultVersion = "2020-01-15"
	ContainerName  = "kaniko"
)

func GetImageLatest() string {
	return GetImage("latest")
}

func GetImage(version string) string {
	return utils.GetImage(ImageName, '-', version)
}

func GetShell() string {
	return "/busybox/sh"
}
