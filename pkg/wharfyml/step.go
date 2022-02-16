package wharfyml

import (
	"errors"

	"gopkg.in/yaml.v3"
)

// Errors related to parsing steps.
var (
	ErrStepEmpty             = errors.New("missing a step type")
	ErrStepMultipleStepTypes = errors.New("contains multiple step types")
)

// Step holds the step type and name of this Wharf build step.
type Step struct {
	Pos  Pos
	Name string
	Type StepType
}

func visitStepNode(name strNode, node *yaml.Node) (step Step, errSlice Errors) {
	step.Pos = newPosNode(node)
	step.Name = name.value
	nodes, errs := visitMapSlice(node)
	errSlice.add(errs...)
	if len(nodes) == 0 {
		errSlice.add(wrapPosErrorNode(ErrStepEmpty, node))
		return
	}
	if len(nodes) > 1 {
		errSlice.add(wrapPosErrorNode(ErrStepMultipleStepTypes, node))
		// Continue, its not a fatal issue
	}
	for _, stepTypeNode := range nodes {
		stepType, errs := visitStepTypeNode(stepTypeNode.key, stepTypeNode.value)
		step.Type = stepType
		if stepType != nil {
			errSlice.add(wrapPathErrorSlice(errs, stepType.StepTypeName())...)
		} else {
			errSlice.add(errs...)
		}
	}
	return
}
