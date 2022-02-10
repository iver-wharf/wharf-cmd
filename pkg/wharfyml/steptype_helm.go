package wharfyml

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

func (s StepHelm) visitStepTypeNode(p nodeMapParser) (StepType, Errors) {
	s.Meta = getStepTypeMeta(p)

	s.Repo = "" // TODO: default to "${CHART_REPO}/${REPO_GROUP}"
	s.Cluster = "kubectl-config"
	s.HelmVersion = "v2.14.1"

	var errSlice Errors

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
