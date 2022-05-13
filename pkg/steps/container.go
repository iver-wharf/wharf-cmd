package steps

import (
	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	v1 "k8s.io/api/core/v1"
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

	podSpec *v1.PodSpec
}

// StepTypeName returns the name of this step type.
func (Container) StepTypeName() string { return "container" }

func (s Container) PodSpec() *v1.PodSpec { return s.podSpec }

func (s Container) init(stepName string, v visit.MapVisitor) (StepType, errutil.Slice) {
	s.OS = "linux"
	s.Shell = "/bin/sh"
	s.ServiceAccount = "default"

	var errSlice errutil.Slice

	// Visiting
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
	return s, errSlice
}
