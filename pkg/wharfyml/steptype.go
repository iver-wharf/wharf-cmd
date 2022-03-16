package wharfyml

import (
	"errors"
	"fmt"

	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"gopkg.in/yaml.v3"
)

// Errors related to parsing step types.
var (
	ErrStepTypeUnknown   = errors.New("unknown step type")
	ErrMissingBuiltinVar = errors.New("missing built-in var")
)

// StepType is an interface that is implemented by all step types.
type StepType interface {
	StepTypeName() string
}

// StepTypeMeta contains metadata about a step type.
type StepTypeMeta struct {
	StepName string
	Source   Pos
	FieldPos map[string]Pos
}

func visitStepTypeNode(stepName string, key strNode, node *yaml.Node, source varsub.Source) (StepType, Errors) {
	visitor, err := visitStepTypeKeyNode(key)
	if err != nil {
		return nil, Errors{err}
	}
	return visitor.visitStepTypeValueNode(stepName, node, source)
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
	visitNode func(stepName string, p nodeMapParser, source varsub.Source) (StepType, Errors)
}

func (v stepTypeVisitor) visitStepTypeValueNode(stepName string, node *yaml.Node, source varsub.Source) (StepType, Errors) {
	var errSlice Errors
	m, errs := visitMap(node)
	errSlice.add(errs...)

	parser := newNodeMapParser(v.keyNode, m)
	stepType, errs := v.visitNode(stepName, parser, source)
	errSlice.add(errs...)

	return stepType, errSlice
}

func getStepTypeMeta(p nodeMapParser) StepTypeMeta {
	return StepTypeMeta{
		Source:   p.parentPos(),
		FieldPos: p.positions,
	}
}
