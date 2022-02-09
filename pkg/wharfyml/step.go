package wharfyml

import (
	"errors"

	"github.com/goccy/go-yaml/ast"
)

// Errors related to parsing steps.
var (
	ErrStepEmpty             = errors.New("missing a step type")
	ErrStepMultipleStepTypes = errors.New("contains multiple step types")
)

// Step holds the step type and name of this Wharf build step.
type Step struct {
	Name string
	Type StepType
}

func visitStepNode(name string, node ast.Node) (step Step, errSlice Errors) {
	step.Name = name
	nodes, err := parseMappingValueNodes(node)
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
		stepType, errs := visitStepTypeNode(stepTypeNode)
		step.Type = stepType
		if stepType != nil {
			errSlice.add(wrapPathErrorSlice(stepType.StepTypeName(), errs)...)
		} else {
			errSlice.add(errs...)
		}
	}
	return
}
