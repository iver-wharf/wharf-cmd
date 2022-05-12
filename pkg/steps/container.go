package steps

import (
	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
)

// Container represents a step type for running commands inside a Docker
// container.
type Container struct {
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
func (Container) StepTypeName() string { return "container" }

func (s *Container) init(stepName string, v visit.MapVisitor) errutil.Slice {
	s.OS = "linux"
	s.Shell = "/bin/sh"
	s.ServiceAccount = "default"

	var errSlice errutil.Slice

	// Visitling
	errSlice.Add(
		v.VisitString("image", &s.Image),
		v.VisitString("os", &s.OS),
		v.VisitString("shell", &s.Shell),
		v.VisitString("secretName", &s.SecretName),
		v.VisitString("serviceAccount", &s.ServiceAccount),
		v.VisitString("certificatesMountPath", &s.CertificatesMountPath),
	)
	errSlice.Add(v.VisitStringSlice("cmds", &s.Cmds)...)

	// Validation
	errSlice.Add(
		v.ValidateRequiredString("image"),
		v.ValidateRequiredSlice("cmds"),
	)
	return errSlice
}
