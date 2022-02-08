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
	// TODO: validate
	return
}

func (s *StepNuGetPackage) unmarshalNodes(nodes nodeMapUnmarshaller) (errSlice errorSlice) {
	errSlice.addNonNils(
		nodes.unmarshalString("version", &s.Version),
		nodes.unmarshalString("project-path", &s.ProjectPath),
		nodes.unmarshalString("repo", &s.Repo),
		nodes.unmarshalBool("skip-duplicate", &s.SkipDuplicate),
		nodes.unmarshalString("certificatesMountPath", &s.CertificatesMountPath),
	)
	return
}

func (s *StepNuGetPackage) resetDefaults() errorSlice {
	return nil
}
