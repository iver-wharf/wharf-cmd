package wharfyml

import (
	"fmt"

	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
)

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

func (s StepHelmPackage) visitStepTypeNode(stepName string, p nodeMapParser, source varsub.Source) (StepType, Errors) {
	s.Meta = getStepTypeMeta(p, stepName)

	var errSlice Errors

	if !p.hasNode("destination") {
		var chartRepo string
		var repoGroup string
		errSlice.addNonNils(
			p.unmarshalStringFromVarSubForOther(
				"CHART_REPO", "destination", source, &chartRepo),
			p.unmarshalStringFromVarSubForOther(
				"REPO_GROUP", "destination", source, &repoGroup),
		)
		s.Destination = fmt.Sprintf("%s/%s", chartRepo, repoGroup)
	}

	// Unmarshalling
	errSlice.addNonNils(
		p.unmarshalString("version", &s.Version),
		p.unmarshalString("chart-path", &s.ChartPath),
		p.unmarshalString("destination", &s.Destination),
	)

	return s, errSlice
}
