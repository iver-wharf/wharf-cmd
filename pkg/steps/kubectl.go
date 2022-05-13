package steps

import (
	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/config"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	v1 "k8s.io/api/core/v1"
)

// Kubectl represents a step type for running kubectl commands on some
// Kubernetes manifest files.
type Kubectl struct {
	// Required fields
	File  string
	Files []string

	// Optional fields
	Namespace string
	Action    string
	Force     bool
	Cluster   string

	config  *config.KubectlStepConfig
	podSpec *v1.PodSpec
}

// StepTypeName returns the name of this step type.
func (Kubectl) StepTypeName() string { return "kubectl" }

func (s Kubectl) PodSpec() *v1.PodSpec { return s.podSpec }

func (s Kubectl) init(stepName string, v visit.MapVisitor) (StepType, errutil.Slice) {
	s.Cluster = "kubectl-config"
	s.Action = "apply"

	var errSlice errutil.Slice

	// Visitling
	errSlice.Add(
		v.VisitString("file", &s.File),
		v.VisitString("namespace", &s.Namespace),
		v.VisitString("action", &s.Action),
		v.VisitBool("force", &s.Force),
		v.VisitString("cluster", &s.Cluster),
	)
	errSlice.Add(v.VisitStringSlice("files", &s.Files)...)

	// Validation
	if len(s.Files) == 0 {
		// Only either file or files is required
		errSlice.Add(v.ValidateRequiredString("file"))
	}
	return s, errSlice
}