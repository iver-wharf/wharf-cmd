package wharfyml

// StepKubectl represents a step type for running kubectl commands on some
// Kubernetes manifest files.
type StepKubectl struct {
	// Required fields
	File  string
	Files []string

	// Optional fields
	Namespace string
	Action    string
	Force     bool
	Cluster   string
}

// StepTypeName returns the name of this step type.
func (StepKubectl) StepTypeName() string { return "kubectl" }

func (s StepKubectl) unmarshalNodes(nodes stepTypeParser) (StepType, Errors) {
	s.Cluster = "kubectl-config"
	s.Action = "apply"

	var errSlice Errors

	// Unmarshalling
	errSlice.addNonNils(
		nodes.unmarshalString("file", &s.File),
		nodes.unmarshalString("namespace", &s.Namespace),
		nodes.unmarshalString("action", &s.Action),
		nodes.unmarshalBool("force", &s.Force),
		nodes.unmarshalString("cluster", &s.Cluster),
	)
	errSlice.add(nodes.unmarshalStringSlice("files", &s.Files)...)

	// Validation
	if len(s.Files) == 0 {
		// Only either file or files is required
		errSlice.addNonNils(nodes.validateRequiredString("file"))
	}
	return s, errSlice
}
