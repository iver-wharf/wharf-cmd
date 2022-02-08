package wharfyml

type StepHelm struct {
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

func (StepHelm) StepTypeName() string { return "helm" }

func (s StepHelm) unmarshalNodes(nodes nodeMapUnmarshaller) (StepType, errorSlice) {
	s.Repo = "" // TODO: default to "${CHART_REPO}/${REPO_GROUP}"
	s.Cluster = "kubectl-config"
	s.HelmVersion = "v2.14.1"

	var errSlice errorSlice

	// Unmarshalling
	errSlice.addNonNils(
		nodes.unmarshalString("chart", &s.Chart),
		nodes.unmarshalString("name", &s.Name),
		nodes.unmarshalString("namespace", &s.Namespace),
		nodes.unmarshalString("repo", &s.Repo),
		nodes.unmarshalString("chartVersion", &s.ChartVersion),
		nodes.unmarshalString("helmVersion", &s.HelmVersion),
		nodes.unmarshalString("cluster", &s.Cluster),
	)
	errSlice.add(nodes.unmarshalStringStringMap("set", &s.Set)...)
	errSlice.add(nodes.unmarshalStringSlice("files", &s.Files)...)
	if s.Repo == "stage" {
		s.Repo = "https://kubernetes-charts.storage.googleapis.com"
	}

	// Validation
	errSlice.addNonNils(
		nodes.validateRequiredString("chart"),
		nodes.validateRequiredString("name"),
		nodes.validateRequiredString("namespace"),
	)
	return s, errSlice
}
