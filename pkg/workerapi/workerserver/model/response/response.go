package response

// Artifact contains some data about an artifact, like its name and what step it
// came from.
type Artifact struct {
	// ArtifactID is the worker's own ID of the artifact. It's unique per build
	// step for a given build, but may have collisons across multiple steps or
	// builds.
	ArtifactID uint `json:"artifactId"`
	// StepID is the worker's own ID of the build step that produced the
	// artifact.
	StepID uint `json:"stepId"`
	// Name is the name of the artifact.
	Name string `json:"name"`
}

// Step holds the step type and name of a Wharf build step.
type Step struct {
	Name     string   `json:"name"`
	StepType StepType `json:"stepType"`
}

// StepType contains the name of a Wharf build step.
type StepType struct {
	Name string `json:"name"`
}
