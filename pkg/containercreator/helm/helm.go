package helm

import (
	"github.com/iver-wharf/wharf-cmd/pkg/containercreator/utils"
)

const (
	ImageName      = "wharfse/helm"
	DefaultVersion = "v2.14.1"
)

func GetImageLatest() string {
	return GetImage("latest")
}

func GetImage(version string) string {
	return utils.GetImage(ImageName, ':', version)
}
