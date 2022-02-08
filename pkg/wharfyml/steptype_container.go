package wharfyml

import "fmt"

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

func (s StepContainer) Validate() (errSlice errorSlice) {
	if s.Image == "" {
		errSlice.add(fmt.Errorf("%w: image", ErrStepTypeMissingRequired))
	}
	if len(s.Cmds) == 0 {
		errSlice.add(fmt.Errorf("%w: cmds", ErrStepTypeMissingRequired))
	}
	return
}

func (s StepContainer) unmarshalNodes(nodes nodeMapUnmarshaller) (StepType, errorSlice) {
	s.OS = "linux"
	s.Shell = "/bin/sh"
	s.ServiceAccount = "default"

	var errSlice errorSlice
	errSlice.addNonNils(
		nodes.unmarshalString("image", &s.Image),
		nodes.unmarshalString("os", &s.OS),
		nodes.unmarshalString("shell", &s.Shell),
		nodes.unmarshalString("secretName", &s.SecretName),
		nodes.unmarshalString("serviceAccount", &s.ServiceAccount),
		nodes.unmarshalString("certificatesMountPath", &s.CertificatesMountPath),
	)
	errSlice.add(nodes.unmarshalStringSlice("cmds", &s.Cmds)...)
	return s, errSlice
}
