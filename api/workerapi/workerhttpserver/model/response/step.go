package response

type Step struct {
	Pos      Pos      `json:"pos"`
	Name     string   `json:"name"`
	StepType StepType `json:"stepType"`
}
