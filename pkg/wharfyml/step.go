package wharfyml

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml/ast"
)

// Errors related to parsing steps.
var (
	ErrStepNotMap            = errors.New("step should be a YAML map")
	ErrStepEmpty             = errors.New("missing a step type")
	ErrStepEmptyName         = errors.New("step name cannot be empty")
	ErrStepMultipleStepTypes = errors.New("contains multiple step types")
)

// Step holds the step type and name of this Wharf build step.
type Step struct {
	Name string
	Type StepType
}

func visitStepNode(key *ast.StringNode, node ast.Node) (step Step, errSlice Errors) {
	step.Name = key.Value
	if key.Value == "" {
		errSlice.add(newPositionedErrorNode(ErrStepEmptyName, key))
		// Continue, its not a fatal issue
	}
	nodes, err := stepBodyAsNodes(node)
	if err != nil {
		errSlice.add(err)
		return
	}
	if len(nodes) == 0 {
		errSlice.add(newPositionedErrorNode(ErrStepEmpty, node))
		return
	}
	if len(nodes) > 1 {
		errSlice.add(newPositionedErrorNode(ErrStepMultipleStepTypes, node))
		// Continue, its not a fatal issue
	}
	for _, stepTypeNode := range nodes {
		key, err := parseMapKey(stepTypeNode.Key)
		if err != nil {
			errSlice.add(err)
			continue
		}
		stepType, errs := visitStepStepTypeNode(key, stepTypeNode)
		step.Type = stepType
		errSlice.add(errs...)
	}
	return
}

func visitStepStepTypeNode(key *ast.StringNode, node *ast.MappingValueNode) (StepType, Errors) {
	stepType, errs := visitStepTypeNode(key, node.Value)
	errs = wrapPathErrorSlice(key.Value, errs)
	return stepType, errs
}

func stepBodyAsNodes(body ast.Node) ([]*ast.MappingValueNode, error) {
	n, ok := getMappingValueNodes(body)
	if !ok {
		return nil, newPositionedErrorNode(fmt.Errorf("step type: %s: %w", body.Type(), ErrStepNotMap), body)
	}
	return n, nil
}
