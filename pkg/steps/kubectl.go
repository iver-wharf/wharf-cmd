package steps

import (
	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
)

// StepKubectl represents a step type for running kubectl commands on some
// Kubernetes manifest files.
type StepKubectl struct {
	// Step type metadata
	Meta StepTypeMeta

	// Required fields
	File  string
	Files []string

	// Optional fields
	Namespace string
	Action    string
	Force     bool
	Cluster   string
}

// Name returns the name of this step type.
func (StepKubectl) Name() string { return "kubectl" }

func (s StepKubectl) visitStepTypeNode(stepName string, p nodeMapParser, _ varsub.Source) (StepType, errutil.Slice) {
	s.Meta = getStepTypeMeta(p, stepName)

	s.Cluster = "kubectl-config"
	s.Action = "apply"

	var errSlice errutil.Slice

	// Unmarshalling
	errSlice.Add(
		p.unmarshalString("file", &s.File),
		p.unmarshalString("namespace", &s.Namespace),
		p.unmarshalString("action", &s.Action),
		p.unmarshalBool("force", &s.Force),
		p.unmarshalString("cluster", &s.Cluster),
	)
	errSlice.Add(p.unmarshalStringSlice("files", &s.Files)...)

	// Validation
	if len(s.Files) == 0 {
		// Only either file or files is required
		errSlice.Add(p.validateRequiredString("file"))
	}
	return s, errSlice
}
