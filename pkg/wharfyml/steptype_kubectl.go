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

func (s StepKubectl) Validate() (errSlice errorSlice) {
	// TODO: validate
	return
}

func (s *StepKubectl) unmarshalNodes(nodes nodeMapUnmarshaller) (errSlice errorSlice) {
	errSlice.addNonNils(
		nodes.unmarshalString("file", &s.File),
		nodes.unmarshalString("namespace", &s.Namespace),
		nodes.unmarshalString("action", &s.Action),
		nodes.unmarshalBool("force", &s.Force),
		nodes.unmarshalString("cluster", &s.Cluster),
	)
	errSlice.add(nodes.unmarshalStringSlice("files", &s.Files)...)
	return
}

func (s *StepKubectl) resetDefaults() errorSlice {
	s.Cluster = "kubectl-config"
	s.Action = "apply"
	return nil
}
