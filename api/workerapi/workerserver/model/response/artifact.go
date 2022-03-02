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
