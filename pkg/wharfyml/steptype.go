package wharfyml

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml/ast"
)

// Errors related to parsing step types.
var (
	ErrStepTypeUnknown         = errors.New("unknown step type")
	ErrStepTypeMissingRequired = errors.New("missing required field")
)

// StepType is an interface that is implemented by all step types.
type StepType interface {
	StepTypeName() string
}

func visitStepTypeNode(node *ast.MappingValueNode) (StepType, Errors) {
	visitor, err := visitStepTypeKeyNode(node.Key)
	if err != nil {
		return nil, Errors{err}
	}
	return visitor.visitStepTypeValueNode(node.Value)
}

func visitStepTypeKeyNode(node ast.Node) (stepTypeVisitor, error) {
	keyNode, err := parseMapKey(node)
	if err != nil {
		return stepTypeVisitor{}, err
	}
	visitor := stepTypeVisitor{
		keyNode: keyNode,
	}
	switch keyNode.Value {
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
		err := fmt.Errorf("%w: %q", ErrStepTypeUnknown, keyNode.Value)
		return stepTypeVisitor{}, wrapPositionedErrorNode(err, keyNode)
	}
	return visitor, nil
}

type stepTypeVisitor struct {
	keyNode   *ast.StringNode
	visitNode func(stepTypeParser) (StepType, Errors)
}

func (v stepTypeVisitor) visitStepTypeValueNode(node ast.Node) (StepType, Errors) {
	nodes, err := parseMappingValueNodes(node)
	if err != nil {
		return nil, Errors{err}
	}
	var errSlice Errors
	m, errs := mappingValueNodeSliceToMap(nodes)
	errSlice.add(errs...)
	parser := stepTypeParser{
		parent: v.keyNode,
		nodes:  m,
	}
	stepType, errs := v.visitNode(parser)
	errSlice.add(errs...)
	return stepType, errSlice
}
