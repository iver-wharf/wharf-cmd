package steps

import (
	"fmt"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/config"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	v1 "k8s.io/api/core/v1"
)

// Helm represents a step type for installing a Helm chart into a Kubernetes
// cluster.
type Helm struct {
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

	config  *config.HelmStepConfig
	podSpec *v1.PodSpec
}

// StepTypeName returns the name of this step type.
func (Helm) StepTypeName() string { return "helm" }

func (s Helm) PodSpec() *v1.PodSpec { return s.podSpec }

func (s Helm) init(stepName string, v visit.MapVisitor) (StepType, errutil.Slice) {
	s.Cluster = "kubectl-config"
	s.HelmVersion = "v2.14.1"

	var errSlice errutil.Slice

	if !v.HasNode("repo") {
		var chartRepo string
		var repoGroup string
		var errs errutil.Slice
		errs.Add(
			v.VisitStringFromVarSub("CHART_REPO", &chartRepo),
			v.VisitStringFromVarSub("REPO_GROUP", &repoGroup),
		)
		for _, err := range errs {
			errSlice.Add(fmt.Errorf(`eval "repo" default: %w`, err))
		}
		s.Repo = fmt.Sprintf("%s/%s", chartRepo, repoGroup)
	}

	// Visitling
	errSlice.Add(
		v.VisitString("chart", &s.Chart),
		v.VisitString("name", &s.Name),
		v.VisitString("namespace", &s.Namespace),
		v.VisitString("repo", &s.Repo),
		v.VisitString("chartVersion", &s.ChartVersion),
		v.VisitString("helmVersion", &s.HelmVersion),
		v.VisitString("cluster", &s.Cluster),
	)
	errSlice.Add(v.VisitStringStringMap("set", &s.Set)...)
	errSlice.Add(v.VisitStringSlice("files", &s.Files)...)
	if s.Repo == "stage" {
		s.Repo = "https://kubernetes-charts.storage.googleapis.com"
	}

	// Validation
	errSlice.Add(
		v.ValidateRequiredString("chart"),
		v.ValidateRequiredString("name"),
		v.ValidateRequiredString("namespace"),
	)
	return s, errSlice
}
