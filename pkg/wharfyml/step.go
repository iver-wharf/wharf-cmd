package wharfyml

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v2"
)

var (
	ErrStepWithMultipleStepTypes = errors.New("step contains multiple step types")
	ErrStepMissingStepType       = errors.New("step is missing a step type")
	ErrStepTypeMissingRequired   = errors.New("step type is missing required field")
)

type Step struct {
	Name      string
	Type      StepType
	Variables map[string]interface{}
}

func parseStep(name string, content map[string]interface{}) (Step, error) {
	if len(content) != 1 {
		return Step{}, fmt.Errorf("expected single step-type, got %d", len(content))
	}

	var stepType StepType
	var variables map[string]interface{}
	for k, v := range content {
		stepType = ParseStepType(k)
		variables = v.(map[string]interface{})
	}

	return Step{Name: name, Type: stepType, Variables: variables}, nil
}

func parseStep2(mapItem yaml.MapItem) (Step, []error) {
	return Step{}, nil
}
