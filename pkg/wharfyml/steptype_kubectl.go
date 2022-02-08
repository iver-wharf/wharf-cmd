package wharfyml

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

func (StepKubectl) StepTypeName() string { return "kubectl" }

func (s StepKubectl) unmarshalNodes(nodes nodeMapUnmarshaller) (StepType, errorSlice) {
	s.Cluster = "kubectl-config"
	s.Action = "apply"

	var errSlice errorSlice

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
