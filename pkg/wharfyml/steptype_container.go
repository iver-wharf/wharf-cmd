package wharfyml

import "github.com/iver-wharf/wharf-cmd/pkg/varsub"

// StepContainer represents a step type for running commands inside a Docker
// container.
type StepContainer struct {
	// Step type metadata
	Meta StepTypeMeta

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

func (s StepContainer) visitStepTypeNode(p nodeMapParser, source varsub.Source) (StepType, Errors) {
	s.Meta = getStepTypeMeta(p)

	s.OS = "linux"
	s.Shell = "/bin/sh"
	s.ServiceAccount = "default"

	var errSlice Errors

	// Unmarshalling
	errSlice.addNonNils(
		p.unmarshalString("image", &s.Image),
		p.unmarshalString("os", &s.OS),
		p.unmarshalString("shell", &s.Shell),
		p.unmarshalString("secretName", &s.SecretName),
		p.unmarshalString("serviceAccount", &s.ServiceAccount),
		p.unmarshalString("certificatesMountPath", &s.CertificatesMountPath),
	)
	errSlice.add(p.unmarshalStringSlice("cmds", &s.Cmds)...)

	// Validation
	errSlice.addNonNils(
		p.validateRequiredString("image"),
		p.validateRequiredSlice("cmds"),
	)
	return s, errSlice
}
