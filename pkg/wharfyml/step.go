package wharfyml

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml/ast"
)

var (
	ErrStepNotMap              = errors.New("step should be a YAML map")
	ErrStepEmpty               = errors.New("missing a step type")
	ErrStepEmptyName           = errors.New("step name cannot be empty")
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

func parseStep2(key *ast.StringNode, node ast.Node) (step Step, errSlice []error) {
	step.Name = key.Value
	if key.Value == "" {
		errSlice = append(errSlice, wrapParseErrNode(ErrStepEmptyName, key))
		// Continue, its not a fatal issue
	}
	nodes, err := stepBodyAsNodes(node)
	if err != nil {
		errSlice = append(errSlice, err)
		return
	}
	if len(nodes) == 0 {
		errSlice = append(errSlice, wrapParseErrNode(ErrStepEmpty, node))
		return
	}
	if len(nodes) > 1 {
		errSlice = append(errSlice, wrapParseErrNode(ErrStepMultipleStepTypes, node))
		// Continue, its not a fatal issue
	}
	for _, stepTypeNode := range nodes {
		key, err := parseMapKey(stepTypeNode.Key)
		if err != nil {
			errSlice = append(errSlice, err)
			continue
		}
		errs := parseStepTypeNodeIntoStep(&step, key, stepTypeNode)
		errSlice = append(errSlice, errs...)
	}
	return
}

func parseStepTypeNodeIntoStep(step *Step, ket *ast.StringNode, node *ast.MappingValueNode) []error {
	return nil
}

func stepBodyAsNodes(body ast.Node) ([]*ast.MappingValueNode, error) {
	n, ok := getMappingValueNodes(body)
	if !ok {
		return nil, wrapParseErrNode(fmt.Errorf("step type: %s: %w", body.Type(), ErrStepNotMap), body)
	}
	return n, nil
}
