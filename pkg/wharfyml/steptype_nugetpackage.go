package wharfyml

import "github.com/iver-wharf/wharf-cmd/pkg/varsub"

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

// StepTypeName returns the name of this step type.
func (StepNuGetPackage) StepTypeName() string { return "nuget-package" }

func (s StepNuGetPackage) visitStepTypeNode(stepName string, p nodeMapParser, _ varsub.Source) (StepType, Errors) {
	s.Meta = getStepTypeMeta(p, stepName)

	var errSlice Errors

	// Unmarshalling
	errSlice.addNonNils(
		p.unmarshalString("version", &s.Version),
		p.unmarshalString("project-path", &s.ProjectPath),
		p.unmarshalString("repo", &s.Repo),
		p.unmarshalBool("skip-duplicate", &s.SkipDuplicate),
		p.unmarshalString("certificatesMountPath", &s.CertificatesMountPath),
	)

	// Validation
	errSlice.addNonNils(
		p.validateRequiredString("version"),
		p.validateRequiredString("project-path"),
		p.validateRequiredString("repo"),
	)
	return s, errSlice
}
