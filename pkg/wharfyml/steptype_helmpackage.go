package wharfyml

type StepHelmPackage struct {
	// Optional fields
	Version     string
	ChartPath   string
	Destination string
}

func (StepHelmPackage) StepTypeName() string { return "helm-package" }

func (s StepHelmPackage) Validate() (errSlice errorSlice) {
	return
}

func (s *StepHelmPackage) unmarshalNodes(nodes nodeMapUnmarshaller) (errSlice errorSlice) {
	errSlice.addNonNils(
		nodes.unmarshalString("version", &s.Version),
		nodes.unmarshalString("chart-path", &s.ChartPath),
		nodes.unmarshalString("destination", &s.Destination),
	)
	return
}

func (s *StepHelmPackage) resetDefaults() errorSlice {
	s.Destination = "" // TODO: default to "${CHART_REPO}/${REPO_GROUP}"
	return nil
}
