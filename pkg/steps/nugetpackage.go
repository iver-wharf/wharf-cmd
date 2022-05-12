package steps

import (
	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
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
}

// StepTypeName returns the name of this step type.
func (NuGetPackage) StepTypeName() string { return "nuget-package" }

func (s NuGetPackage) init(stepName string, v visit.MapVisitor) (wharfyml.StepType, errutil.Slice) {
	var errSlice errutil.Slice

	// Visitling
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
