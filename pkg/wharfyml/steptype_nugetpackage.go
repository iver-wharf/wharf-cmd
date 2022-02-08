package wharfyml

type StepNuGetPackage struct {
	// Required fields
	Version     string
	ProjectPath string
	Repo        string

	// Optional fields
	SkipDuplicate         bool
	CertificatesMountPath string
}

func (StepNuGetPackage) StepTypeName() string { return "nuget-package" }

func (s StepNuGetPackage) Validate() (errSlice errorSlice) {
	validateRequiredString(&errSlice, "version", s.Version)
	validateRequiredString(&errSlice, "project-path", s.ProjectPath)
	validateRequiredString(&errSlice, "repo", s.Repo)
	return
}

func (s StepNuGetPackage) unmarshalNodes(nodes nodeMapUnmarshaller) (StepType, errorSlice) {
	var errSlice errorSlice
	errSlice.addNonNils(
		nodes.unmarshalString("version", &s.Version),
		nodes.unmarshalString("project-path", &s.ProjectPath),
		nodes.unmarshalString("repo", &s.Repo),
		nodes.unmarshalBool("skip-duplicate", &s.SkipDuplicate),
		nodes.unmarshalString("certificatesMountPath", &s.CertificatesMountPath),
	)
	return s, errSlice
}
