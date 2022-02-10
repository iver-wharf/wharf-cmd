package wharfyml

// StepContainer represents a step type for running commands inside a Docker
// container.
type StepContainer struct {
	// Step type metadata
	Pos Pos

	// Required fields
	Image string
	Cmds  []string

	// Optional fields
	OS                    string
	Shell                 string
	SecretName            string
	ServiceAccount        string
	CertificatesMountPath string
}

// StepTypeName returns the name of this step type.
func (StepContainer) StepTypeName() string { return "container" }

func (s StepContainer) visitStepTypeNode(nodes nodeMapParser) (StepType, Errors) {
	s.Pos = nodes.parentPos()
	s.OS = "linux"
	s.Shell = "/bin/sh"
	s.ServiceAccount = "default"

	var errSlice Errors

	// Unmarshalling
	errSlice.addNonNils(
		nodes.unmarshalString("image", &s.Image),
		nodes.unmarshalString("os", &s.OS),
		nodes.unmarshalString("shell", &s.Shell),
		nodes.unmarshalString("secretName", &s.SecretName),
		nodes.unmarshalString("serviceAccount", &s.ServiceAccount),
		nodes.unmarshalString("certificatesMountPath", &s.CertificatesMountPath),
	)
	errSlice.add(nodes.unmarshalStringSlice("cmds", &s.Cmds)...)

	// Validation
	errSlice.addNonNils(
		nodes.validateRequiredString("image"),
		nodes.validateRequiredSlice("cmds"),
	)
	return s, errSlice
}
