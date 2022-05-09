package steps

import (
	"fmt"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
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

// Name returns the name of this step type.
func (StepHelmPackage) Name() string { return "helm-package" }

func (s StepHelmPackage) visitStepTypeNode(stepName string, p nodeMapParser, source varsub.Source) (StepType, errutil.Slice) {
	s.Meta = getStepTypeMeta(p, stepName)

	var errSlice errutil.Slice

	if !p.hasNode("destination") {
		var chartRepo string
		var repoGroup string
		var errs errutil.Slice
		errs.Add(
			p.unmarshalStringFromVarSub("CHART_REPO", source, &chartRepo),
			p.unmarshalStringFromVarSub("REPO_GROUP", source, &repoGroup),
		)
		for _, err := range errs {
			errSlice.Add(fmt.Errorf(`eval "destination" default: %w`, err))
		}
		s.Destination = fmt.Sprintf("%s/%s", chartRepo, repoGroup)
	}

	// Unmarshalling
	errSlice.Add(
		p.unmarshalString("version", &s.Version),
		p.unmarshalString("chart-path", &s.ChartPath),
		p.unmarshalString("destination", &s.Destination),
	)

	return s, errSlice
}
