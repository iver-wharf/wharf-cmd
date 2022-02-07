package wharfyml

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml/ast"
)

var (
	ErrStepNotMap              = errors.New("step should be a YAML map")
	ErrStepEmpty               = errors.New("missing a step type")
	ErrStepMultipleStepTypes   = errors.New("contains multiple step types")
	ErrStepTypeInvalidField    = errors.New("invalid field type")
	ErrStepTypeMissingRequired = errors.New("missing required field")
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

// -------------------

func parseStep2(key *ast.StringNode, node *ast.MappingValueNode) (Step, []error) {
	return Step{}, nil
}
