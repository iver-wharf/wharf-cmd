package steps

import (
	_ "embed"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	"gopkg.in/typ.v4"
	v1 "k8s.io/api/core/v1"
)

var (
	//go:embed k8sscript-nuget-package.sh
	nugetPackageScript string
)

// NuGetPackage represents a step type used for building .NET NuGet
// packages.
type NuGetPackage struct {
	// Required fields
	Version     string
	ProjectPath string
	Repo        string

	// Optional fields
	SkipDuplicate         bool
	CertificatesMountPath string

	podSpec *v1.PodSpec
}

// StepTypeName returns the name of this step type.
func (NuGetPackage) StepTypeName() string { return "nuget-package" }

// PodSpec returns this step's Kubernetes Pod specification. Meant to be used
// by the wharf-cmd-worker when creating the actual pods.
func (s NuGetPackage) PodSpec() *v1.PodSpec { return s.podSpec }

func (s NuGetPackage) init(_ string, v visit.MapVisitor) (StepType, errutil.Slice) {
	var errSlice errutil.Slice

	// Visiting
	errSlice.Add(
		v.VisitString("version", &s.Version),
		v.VisitString("project-path", &s.ProjectPath),
		v.VisitString("repo", &s.Repo),
		v.VisitBool("skip-duplicate", &s.SkipDuplicate),
		v.VisitString("certificatesMountPath", &s.CertificatesMountPath),
	)

	// Validation
	errSlice.Add(
		v.ValidateRequiredString("version"),
		v.ValidateRequiredString("project-path"),
		v.ValidateRequiredString("repo"),
	)

	podSpec, errs := s.applyStep(v)
	s.podSpec = podSpec
	errSlice.Add(errs...)

	return s, errSlice
}

func (step NuGetPackage) applyStep(v visit.MapVisitor) (*v1.PodSpec, errutil.Slice) {
	var errSlice errutil.Slice
	podSpec := newBasePodSpec()

	cont := v1.Container{
		Name:       commonContainerName,
		Image:      "mcr.microsoft.com/dotnet/sdk:3.1-alpine",
		WorkingDir: commonRepoVolumeMount.MountPath,
		VolumeMounts: []v1.VolumeMount{
			commonRepoVolumeMount,
		},
		Env: []v1.EnvVar{
			{
				Name: "NUGET_TOKEN",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "wharf-nuget-api-token",
						},
						Key: "token",
					},
				},
			},
			{Name: "NUGET_REPO", Value: step.Repo},
			{Name: "NUGET_PROJECT_PATH", Value: step.ProjectPath},
			{Name: "NUGET_VERSION", Value: step.Version},
			{Name: "NUGET_SKIP_DUP", Value: typ.Tern(step.SkipDuplicate, "true", "false")},
		},
		Command: []string{"/bin/bash", "-c"},
		Args:    []string{nugetPackageScript},
	}

	podSpec.Containers = append(podSpec.Containers, cont)
	return &podSpec, errSlice
}
