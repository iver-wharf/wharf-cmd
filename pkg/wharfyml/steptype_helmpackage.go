package wharfyml

import "github.com/iver-wharf/wharf-cmd/pkg/varsub"

// StepHelmPackage represents a step type for building and uploading a Helm
// chart to a chart registry.
type StepHelmPackage struct {
	// Step type metadata
	Meta StepTypeMeta

	// Optional fields
	Version     string
	ChartPath   string
	Destination string
}

// StepTypeName returns the name of this step type.
func (StepHelmPackage) StepTypeName() string { return "helm-package" }

func (s StepHelmPackage) visitStepTypeNode(p nodeMapParser, source varsub.Source) (StepType, Errors) {
	s.Meta = getStepTypeMeta(p)

	s.Destination = "" // TODO: default to "${CHART_REPO}/${REPO_GROUP}"

	var errSlice Errors

	// Unmarshalling
	errSlice.addNonNils(
		p.unmarshalString("version", &s.Version),
		p.unmarshalString("chart-path", &s.ChartPath),
		p.unmarshalString("destination", &s.Destination),
	)

	return s, errSlice
}
