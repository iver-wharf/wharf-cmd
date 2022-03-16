package wharfyml

import (
	"fmt"

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

// StepTypeName returns the name of this step type.
func (StepHelm) StepTypeName() string { return "helm" }

func (s StepHelm) visitStepTypeNode(stepName string, p nodeMapParser, source varsub.Source) (StepType, Errors) {
	s.Meta = getStepTypeMeta(p)

	s.Cluster = "kubectl-config"
	s.HelmVersion = "v2.14.1"

	var errSlice Errors

	if !p.hasNode("repo") {
		var chartRepo string
		var repoGroup string
		errSlice.addNonNils(
			p.unmarshalStringFromVarSubForOther(
				"CHART_REPO", "repo", source, &chartRepo),
			p.unmarshalStringFromVarSubForOther(
				"REPO_GROUP", "repo", source, &repoGroup),
		)
		s.Repo = fmt.Sprintf("%s/%s", chartRepo, repoGroup)
	}

	// Unmarshalling
	errSlice.addNonNils(
		p.unmarshalString("chart", &s.Chart),
		p.unmarshalString("name", &s.Name),
		p.unmarshalString("namespace", &s.Namespace),
		p.unmarshalString("repo", &s.Repo),
		p.unmarshalString("chartVersion", &s.ChartVersion),
		p.unmarshalString("helmVersion", &s.HelmVersion),
		p.unmarshalString("cluster", &s.Cluster),
	)
	errSlice.add(p.unmarshalStringStringMap("set", &s.Set)...)
	errSlice.add(p.unmarshalStringSlice("files", &s.Files)...)
	if s.Repo == "stage" {
		s.Repo = "https://kubernetes-charts.storage.googleapis.com"
	}

	// Validation
	errSlice.addNonNils(
		p.validateRequiredString("chart"),
		p.validateRequiredString("name"),
		p.validateRequiredString("namespace"),
	)
	return s, errSlice
}
