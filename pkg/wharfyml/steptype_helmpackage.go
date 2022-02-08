package wharfyml

// StepHelmPackage represents a step type for building and uploading a Helm
// chart to a chart registry.
type StepHelmPackage struct {
	// Optional fields
	Version     string
	ChartPath   string
	Destination string
}

// StepTypeName returns the name of this step type.
func (StepHelmPackage) StepTypeName() string { return "helm-package" }

func (s StepHelmPackage) unmarshalNodes(nodes stepTypeParser) (StepType, Errors) {
	s.Destination = "" // TODO: default to "${CHART_REPO}/${REPO_GROUP}"

	var errSlice Errors

	// Unmarshalling
	errSlice.addNonNils(
		nodes.unmarshalString("version", &s.Version),
		nodes.unmarshalString("chart-path", &s.ChartPath),
		nodes.unmarshalString("destination", &s.Destination),
	)

	return s, errSlice
}
