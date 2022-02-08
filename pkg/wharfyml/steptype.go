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

type StepType int

const (
	Container = StepType(iota + 1)
	Kaniko
	Docker
	HelmDeploy
	HelmPackage
	KubeApply
)

var strToStepType = map[string]StepType{
	"container":    Container,
	"kaniko":       Kaniko,
	"helm":         HelmDeploy,
	"helm-package": HelmPackage,
	"kubectl":      KubeApply,
	"docker":       Docker,
}

var stepTypeToString = map[StepType]string{}

func init() {
	for str, st := range strToStepType {
		stepTypeToString[st] = str
	}
}

func (t StepType) String() string {
	return stepTypeToString[t]
}

func ParseStepType(name string) StepType {
	return strToStepType[name]
}

// ------------------

type StepType2 interface {
	StepTypeName() string
	Validate() errorSlice
}

type stepTypePrep interface {
	StepType2

	resetDefaults() errorSlice
	unmarshalNodes(nodes nodeMapUnmarshaller) errorSlice
}

func visitStepTypeNode(key *ast.StringNode, node ast.Node) (StepType2, errorSlice) {
	nodes, err := stepTypeBodyAsNodes(node)
	if err != nil {
		return nil, errorSlice{err}
	}
	stepType, errs := unmarshalStepTypeNode(key, nodes)
	if len(errs) > 0 {
		return stepType, errs
	}
	return stepType, stepType.Validate()
}

func unmarshalStepTypeNode(key *ast.StringNode, nodes []*ast.MappingValueNode) (StepType2, errorSlice) {
	stepType, err := getStepTypeUnmarshaller(key)
	if err != nil {
		return nil, errorSlice{err}
	}
	var errSlice errorSlice
	errSlice.add(stepType.resetDefaults()...)

	m, errs := mappingValueNodeSliceToMap(nodes)
	errSlice.add(errs...)
	if m != nil {
		errSlice.add(stepType.unmarshalNodes(nodeMapUnmarshaller(m))...)
	}
	return stepType, errSlice
}

func getStepTypeUnmarshaller(key *ast.StringNode) (stepTypePrep, error) {
	switch key.Value {
	case "container":
		return &StepContainer{}, nil
	case "docker":
		return &StepDocker{}, nil
	case "helm":
		return &StepHelm{}, nil
	case "helm-package":
		return &StepHelmPackage{}, nil
	case "kubectl":
		return &StepKubectl{}, nil
	case "nuget-package":
		return &StepNuGetPackage{}, nil
	default:
		return nil, newParseErrorNode(ErrStepTypeUnknown, key)
	}
}

func stepTypeBodyAsNodes(body ast.Node) ([]*ast.MappingValueNode, error) {
	n, ok := getMappingValueNodes(body)
	if !ok {
		return nil, newParseErrorNode(fmt.Errorf("step type type: %s: %w",
			body.Type(), ErrStepTypeNotMap), body)
	}
	return n, nil
}

func yamlUnmarshalNodeWithValidator(node ast.Node, valuePtr interface{}) error {
	var buf bytes.Buffer
	return yaml.NewDecoder(&buf).DecodeFromNode(node, valuePtr)
}
