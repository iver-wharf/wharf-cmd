package steps

import (
	"fmt"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	v1 "k8s.io/api/core/v1"
)

// HelmPackage represents a step type for building and uploading a Helm
// chart to a chart registry.
type HelmPackage struct {
	// Optional fields
	Version     string
	ChartPath   string
	Destination string

	podSpec *v1.PodSpec
}

// StepTypeName returns the name of this step type.
func (HelmPackage) StepTypeName() string { return "helm-package" }

// PodSpec returns this step's Kubernetes Pod specification. Meant to be used
// by the wharf-cmd-worker when creating the actual pods.
func (s HelmPackage) PodSpec() *v1.PodSpec { return s.podSpec }

func (s HelmPackage) init(_ string, v visit.MapVisitor) (StepType, errutil.Slice) {
	var errSlice errutil.Slice

	if !v.HasNode("destination") {
		var chartRepo string
		var repoGroup string
		var errs errutil.Slice
		errs.Add(
			v.RequireStringFromVarSub("CHART_REPO", &chartRepo),
			v.RequireStringFromVarSub("REPO_GROUP", &repoGroup),
		)
		for _, err := range errs {
			errSlice.Add(fmt.Errorf(`eval "destination" default: %w`, err))
		}
		s.Destination = fmt.Sprintf("%s/%s", chartRepo, repoGroup)
	}

	// Visiting
	errSlice.Add(
		v.VisitString("version", &s.Version),
		v.VisitString("chart-path", &s.ChartPath),
		v.VisitString("destination", &s.Destination),
	)

	return s, errSlice
}
