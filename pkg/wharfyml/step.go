package wharfyml

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml/ast"
)

var (
	ErrStepNotMap            = errors.New("step should be a YAML map")
	ErrStepEmpty             = errors.New("missing a step type")
	ErrStepEmptyName         = errors.New("step name cannot be empty")
	ErrStepMultipleStepTypes = errors.New("contains multiple step types")
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

type Step2 struct {
	Name string
	Type StepType2
}

func parseStep2(key *ast.StringNode, node ast.Node) (step Step2, errSlice errorSlice) {
	step.Name = key.Value
	if key.Value == "" {
		errSlice.add(newParseErrorNode(ErrStepEmptyName, key))
		// Continue, its not a fatal issue
	}
	nodes, err := stepBodyAsNodes(node)
	if err != nil {
		errSlice.add(err)
		return
	}
	if len(nodes) == 0 {
		errSlice.add(newParseErrorNode(ErrStepEmpty, node))
		return
	}
	if len(nodes) > 1 {
		errSlice.add(newParseErrorNode(ErrStepMultipleStepTypes, node))
		// Continue, its not a fatal issue
	}
	for _, stepTypeNode := range nodes {
		key, err := parseMapKey(stepTypeNode.Key)
		if err != nil {
			errSlice.add(err)
			continue
		}
		stepType, errs := parseStepStepTypeNode(key, stepTypeNode)
		step.Type = stepType
		errSlice.add(errs...)
	}
	return
}

func parseStepStepTypeNode(key *ast.StringNode, node *ast.MappingValueNode) (StepType2, errorSlice) {
	stepType, errs := parseStepType(key, node.Value)
	errs.fmtErrorfAll("type %q: %w", key.Value, fmtErrorfPlaceholder)
	return stepType, errs
}

func stepBodyAsNodes(body ast.Node) ([]*ast.MappingValueNode, error) {
	n, ok := getMappingValueNodes(body)
	if !ok {
		return nil, newParseErrorNode(fmt.Errorf("step type: %s: %w", body.Type(), ErrStepNotMap), body)
	}
	return n, nil
}
