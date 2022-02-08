package wharfyml

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

var (
	ErrStepTypeNotMap          = errors.New("step type should be a YAML map")
	ErrStepTypeUnknown         = errors.New("unknown step type")
	ErrStepTypeMissingRequired = errors.New("missing required field")
)

type StepType interface {
	StepTypeName() string
}

func visitStepTypeNode(key *ast.StringNode, node ast.Node) (StepType, errorSlice) {
	nodes, err := stepTypeBodyAsNodes(node)
	if err != nil {
		return nil, errorSlice{err}
	}
	return unmarshalStepTypeNode(key, nodes)
}

func unmarshalStepTypeNode(key *ast.StringNode, nodes []*ast.MappingValueNode) (StepType, errorSlice) {
	var errSlice errorSlice
	m, errs := mappingValueNodeSliceToMap(nodes)
	errSlice.add(errs...)
	stepType, errs := getStepTypeUnmarshalled(key, nodeMapUnmarshaller{
		parent: key,
		nodes:  m,
	})
	errSlice.add(errs...)
	return stepType, errSlice
}

func getStepTypeUnmarshalled(key *ast.StringNode, nodes nodeMapUnmarshaller) (StepType, errorSlice) {
	switch key.Value {
	case "container":
		return StepContainer{}.unmarshalNodes(nodes)
	case "docker":
		return StepDocker{}.unmarshalNodes(nodes)
	case "helm":
		return StepHelm{}.unmarshalNodes(nodes)
	case "helm-package":
		return StepHelmPackage{}.unmarshalNodes(nodes)
	case "kubectl":
		return StepKubectl{}.unmarshalNodes(nodes)
	case "nuget-package":
		return StepNuGetPackage{}.unmarshalNodes(nodes)
	default:
		return nil, errorSlice{newPositionedErrorNode(ErrStepTypeUnknown, key)}
	}
}

func stepTypeBodyAsNodes(body ast.Node) ([]*ast.MappingValueNode, error) {
	n, ok := getMappingValueNodes(body)
	if !ok {
		return nil, newPositionedErrorNode(fmt.Errorf("step type type: %s: %w",
			body.Type(), ErrStepTypeNotMap), body)
	}
	return n, nil
}

func yamlUnmarshalNodeWithValidator(node ast.Node, valuePtr interface{}) error {
	var buf bytes.Buffer
	return yaml.NewDecoder(&buf).DecodeFromNode(node, valuePtr)
}
