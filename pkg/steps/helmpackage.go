package steps

import (
	"fmt"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
)

// HelmPackage represents a step type for building and uploading a Helm
// chart to a chart registry.
type HelmPackage struct {
	// Optional fields
	Version     string
	ChartPath   string
	Destination string
}

// StepTypeName returns the name of this step type.
func (HelmPackage) StepTypeName() string { return "helm-package" }

func (s *HelmPackage) init(stepName string, v visit.MapVisitor) errutil.Slice {
	var errSlice errutil.Slice

	if !v.HasNode("destination") {
		var chartRepo string
		var repoGroup string
		var errs errutil.Slice
		errs.Add(
			v.VisitStringFromVarSub("CHART_REPO", &chartRepo),
			v.VisitStringFromVarSub("REPO_GROUP", &repoGroup),
		)
		for _, err := range errs {
			errSlice.Add(fmt.Errorf(`eval "destination" default: %w`, err))
		}
		s.Destination = fmt.Sprintf("%s/%s", chartRepo, repoGroup)
	}

	// Visitling
	errSlice.Add(
		v.VisitString("version", &s.Version),
		v.VisitString("chart-path", &s.ChartPath),
		v.VisitString("destination", &s.Destination),
	)

	return errSlice
}
