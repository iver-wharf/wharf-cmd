package wharfyml

type StepHelmPackage struct {
	// Optional fields
	Version     string
	ChartPath   string
	Destination string
}

func (StepHelmPackage) StepTypeName() string { return "helm-package" }

func (s StepHelmPackage) unmarshalNodes(nodes nodeMapUnmarshaller) (StepType, errorSlice) {
	s.Destination = "" // TODO: default to "${CHART_REPO}/${REPO_GROUP}"

	var errSlice errorSlice

	// Unmarshalling
	errSlice.addNonNils(
		nodes.unmarshalString("version", &s.Version),
		nodes.unmarshalString("chart-path", &s.ChartPath),
		nodes.unmarshalString("destination", &s.Destination),
	)

	return s, errSlice
}
