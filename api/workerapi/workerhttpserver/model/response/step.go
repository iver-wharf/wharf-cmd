package response

// Step holds the step type and name of a Wharf build step.
type Step struct {
	Pos      Pos      `json:"pos"`
	Name     string   `json:"name"`
	StepType StepType `json:"stepType"`
}
