package steps

import (
	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	v1 "k8s.io/api/core/v1"
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
	return s, errSlice
}
