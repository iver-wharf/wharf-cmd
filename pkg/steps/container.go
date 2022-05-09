package steps

import (
	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
)

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

// Name returns the name of this step type.
func (StepContainer) Name() string { return "container" }

func (s StepContainer) visitStepTypeNode(stepName string, p nodeMapParser, _ varsub.Source) (StepType, errutil.Slice) {
	s.Meta = getStepTypeMeta(p, stepName)

	s.OS = "linux"
	s.Shell = "/bin/sh"
	s.ServiceAccount = "default"

	var errSlice errutil.Slice

	// Unmarshalling
	errSlice.Add(
		p.unmarshalString("image", &s.Image),
		p.unmarshalString("os", &s.OS),
		p.unmarshalString("shell", &s.Shell),
		p.unmarshalString("secretName", &s.SecretName),
		p.unmarshalString("serviceAccount", &s.ServiceAccount),
		p.unmarshalString("certificatesMountPath", &s.CertificatesMountPath),
	)
	errSlice.Add(p.unmarshalStringSlice("cmds", &s.Cmds)...)

	// Validation
	errSlice.Add(
		p.validateRequiredString("image"),
		p.validateRequiredSlice("cmds"),
	)
	return s, errSlice
}
