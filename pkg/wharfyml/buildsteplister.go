package wharfyml

// BuildStepLister is an interface which contains a single method
// that returns a slice of steps.
type BuildStepLister interface {
	// ListBuildSteps returns a slice of steps.
	ListBuildSteps() []Step
}
