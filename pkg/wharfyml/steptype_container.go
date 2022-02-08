package wharfyml

type StepContainer struct {
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

func (StepContainer) StepTypeName() string { return "container" }

func (s StepContainer) unmarshalNodes(nodes nodeMapUnmarshaller) (StepType, errorSlice) {
	s.OS = "linux"
	s.Shell = "/bin/sh"
	s.ServiceAccount = "default"

	var errSlice errorSlice

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
