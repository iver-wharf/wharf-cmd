package steps

import (
	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
)

// StepNuGetPackage represents a step type used for building .NET NuGet
// packages.
type StepNuGetPackage struct {
	// Step type metadata
	Meta StepTypeMeta

	// Required fields
	Version     string
	ProjectPath string
	Repo        string

	// Optional fields
	SkipDuplicate         bool
	CertificatesMountPath string
}

// Name returns the name of this step type.
func (StepNuGetPackage) Name() string { return "nuget-package" }

func (s StepNuGetPackage) visitStepTypeNode(stepName string, p nodeMapParser, _ varsub.Source) (StepType, errutil.Slice) {
	s.Meta = getStepTypeMeta(p, stepName)

	var errSlice errutil.Slice

	// Unmarshalling
	errSlice.Add(
		p.unmarshalString("version", &s.Version),
		p.unmarshalString("project-path", &s.ProjectPath),
		p.unmarshalString("repo", &s.Repo),
		p.unmarshalBool("skip-duplicate", &s.SkipDuplicate),
		p.unmarshalString("certificatesMountPath", &s.CertificatesMountPath),
	)

	// Validation
	errSlice.Add(
		p.validateRequiredString("version"),
		p.validateRequiredString("project-path"),
		p.validateRequiredString("repo"),
	)
	return s, errSlice
}
