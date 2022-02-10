package wharfyml

// StepNuGetPackage represents a step type used for building .NET NuGet
// packages.
type StepNuGetPackage struct {
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

func (s StepNuGetPackage) visitStepTypeNode(nodes nodeMapParser) (StepType, Errors) {
	var errSlice Errors

	// Unmarshalling
	errSlice.addNonNils(
		nodes.unmarshalString("version", &s.Version),
		nodes.unmarshalString("project-path", &s.ProjectPath),
		nodes.unmarshalString("repo", &s.Repo),
		nodes.unmarshalBool("skip-duplicate", &s.SkipDuplicate),
		nodes.unmarshalString("certificatesMountPath", &s.CertificatesMountPath),
	)

	// Validation
	errSlice.addNonNils(
		nodes.validateRequiredString("version"),
		nodes.validateRequiredString("project-path"),
		nodes.validateRequiredString("repo"),
	)
	return s, errSlice
}
