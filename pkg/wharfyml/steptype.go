package wharfyml

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Errors related to parsing step types.
var (
	ErrStepTypeUnknown = errors.New("unknown step type")
)

// StepType is an interface that is implemented by all step types.
type StepType interface {
	StepTypeName() string
}

// StepTypeMeta contains metadata about a step type.
type StepTypeMeta struct {
	Source   Pos
	FieldPos map[string]Pos
}

func visitStepTypeNode(key strNode, node *yaml.Node) (StepType, Errors) {
	visitor, err := visitStepTypeKeyNode(key)
	if err != nil {
		return nil, Errors{err}
	}
	return visitor.visitStepTypeValueNode(node)
}

func visitStepTypeKeyNode(key strNode) (stepTypeVisitor, error) {
	visitor := stepTypeVisitor{
		keyNode: key.node,
	}
	switch key.value {
	case "container":
		visitor.visitNode = StepContainer{}.visitStepTypeNode
	case "docker":
		visitor.visitNode = StepDocker{}.visitStepTypeNode
	case "helm":
		visitor.visitNode = StepHelm{}.visitStepTypeNode
	case "helm-package":
		visitor.visitNode = StepHelmPackage{}.visitStepTypeNode
	case "kubectl":
		visitor.visitNode = StepKubectl{}.visitStepTypeNode
	case "nuget-package":
		visitor.visitNode = StepNuGetPackage{}.visitStepTypeNode
	default:
		err := fmt.Errorf("%w: %q", ErrStepTypeUnknown, key.value)
		return stepTypeVisitor{}, wrapPosErrorNode(err, key.node)
	}
	return visitor, nil
}

type stepTypeVisitor struct {
	keyNode   *yaml.Node
	visitNode func(nodeMapParser) (StepType, Errors)
}

func (v stepTypeVisitor) visitStepTypeValueNode(node *yaml.Node) (StepType, Errors) {
	var errSlice Errors
	m, errs := visitMap(node)
	errSlice.add(errs...)

	parser := newNodeMapParser(v.keyNode, m)
	stepType, errs := v.visitNode(parser)
	errSlice.add(errs...)

	return stepType, errSlice
}

func getStepTypeMeta(p nodeMapParser) StepTypeMeta {
	return StepTypeMeta{
		Source:   p.parentPos(),
		FieldPos: p.positions,
	}
}
