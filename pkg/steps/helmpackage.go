package steps

import (
	_ "embed"
	"fmt"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	v1 "k8s.io/api/core/v1"
)

var (
	//go:embed k8sscript-helm-package.sh
	helmPackageScript string
)

// HelmPackage represents a step type for building and uploading a Helm
// chart to a chart registry.
type HelmPackage struct {
	// Optional fields
	Version     string
	ChartPath   string
	Destination string
	Secret      string

	podSpec v1.PodSpec
}

// StepTypeName returns the name of this step type.
func (HelmPackage) StepTypeName() string { return "helm-package" }

// PodSpec returns this step's Kubernetes Pod specification. Meant to be used
// by the wharf-cmd-worker when creating the actual pods.
func (s HelmPackage) PodSpec() v1.PodSpec { return s.podSpec }

func (s HelmPackage) init(_ string, v visit.MapVisitor) (StepType, errutil.Slice) {
	var errSlice errutil.Slice

	if !v.HasNode("destination") {
		var chartRepo string
		var repoGroup string
		var errs errutil.Slice
		errs.Add(
			v.RequireStringFromVarSub("CHART_REPO", &chartRepo),
			v.RequireStringFromVarSub("REPO_GROUP", &repoGroup),
		)
		for _, err := range errs {
			errSlice.Add(fmt.Errorf(`eval "destination" default: %w`, err))
		}
		s.Destination = fmt.Sprintf("%s/%s", chartRepo, repoGroup)
	}

	if v.HasNode("secret") {
		errSlice.Add(v.VisitString("secret", &s.Secret))
	} else {
		errSlice.Add(v.LookupStringFromVarSub("HELM_REG_SECRET", &s.Secret))
	}

	// Visiting
	errSlice.Add(
		v.VisitString("version", &s.Version),
		v.VisitString("chart-path", &s.ChartPath),
		v.VisitString("destination", &s.Destination),
	)

	s.podSpec = s.applyStep()

	return s, errSlice
}

func (s HelmPackage) applyStep() v1.PodSpec {
	podSpec := newBasePodSpec()

	cont := v1.Container{
		Name:       commonContainerName,
		Image:      "wharfse/helm:v3.8.1",
		WorkingDir: commonRepoVolumeMount.MountPath,
		VolumeMounts: []v1.VolumeMount{
			commonRepoVolumeMount,
		},
		Env: []v1.EnvVar{
			{Name: "CHART_PATH", Value: s.ChartPath},
			{Name: "CHART_REPO", Value: s.Destination},
			{Name: "CHART_VERSION", Value: s.Version},
		},
		Command: []string{"/bin/bash", "-c"},
		Args:    []string{helmPackageScript},
	}

	if s.Secret != "" {
		addHelmSecretVolume(s.Secret, &podSpec, &cont)
	}

	podSpec.Containers = append(podSpec.Containers, cont)
	return podSpec
}
