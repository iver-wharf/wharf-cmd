package steps

import (
	"fmt"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
)

// StepHelm represents a step type for installing a Helm chart into a Kubernetes
// cluster.
type StepHelm struct {
	// Step type metadata
	Meta StepTypeMeta

	// Required fields
	Chart     string
	Name      string
	Namespace string

	// Optional fields
	Repo         string
	Set          map[string]string
	Files        []string
	ChartVersion string
	HelmVersion  string
	Cluster      string
}

// Name returns the name of this step type.
func (StepHelm) Name() string { return "helm" }

func (s StepHelm) visitStepTypeNode(stepName string, p nodeMapParser, source varsub.Source) (StepType, errutil.Slice) {
	s.Meta = getStepTypeMeta(p, stepName)

	s.Cluster = "kubectl-config"
	s.HelmVersion = "v2.14.1"

	var errSlice errutil.Slice

	if !p.hasNode("repo") {
		var chartRepo string
		var repoGroup string
		var errs errutil.Slice
		errs.Add(
			p.unmarshalStringFromVarSub("CHART_REPO", source, &chartRepo),
			p.unmarshalStringFromVarSub("REPO_GROUP", source, &repoGroup),
		)
		for _, err := range errs {
			errSlice.Add(fmt.Errorf(`eval "repo" default: %w`, err))
		}
		s.Repo = fmt.Sprintf("%s/%s", chartRepo, repoGroup)
	}

	// Unmarshalling
	errSlice.Add(
		p.unmarshalString("chart", &s.Chart),
		p.unmarshalString("name", &s.Name),
		p.unmarshalString("namespace", &s.Namespace),
		p.unmarshalString("repo", &s.Repo),
		p.unmarshalString("chartVersion", &s.ChartVersion),
		p.unmarshalString("helmVersion", &s.HelmVersion),
		p.unmarshalString("cluster", &s.Cluster),
	)
	errSlice.Add(p.unmarshalStringStringMap("set", &s.Set)...)
	errSlice.Add(p.unmarshalStringSlice("files", &s.Files)...)
	if s.Repo == "stage" {
		s.Repo = "https://kubernetes-charts.storage.googleapis.com"
	}

	// Validation
	errSlice.Add(
		p.validateRequiredString("chart"),
		p.validateRequiredString("name"),
		p.validateRequiredString("namespace"),
	)
	return s, errSlice
}
