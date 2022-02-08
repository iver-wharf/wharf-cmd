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
	case "helm-package":
		return &StepHelmPackage{}, nil
	case "helm", "kubectl", "nuget-package":
		return nil, newParseErrorNode(errors.New("not yet implemented"), key)
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

type StepContainer struct {
	// Required fields
	Image string
	Cmds  []string

	// Optional fields
	OS                    string
	Shell                 string
	SecretName            string
	ServiceAccount        string
	CertificatesMountPath string
}

func (StepContainer) StepTypeName() string { return "container" }

func (s StepContainer) Validate() (errSlice errorSlice) {
	if s.Image == "" {
		errSlice.add(fmt.Errorf("%w: image", ErrStepTypeMissingRequired))
	}
	if len(s.Cmds) == 0 {
		errSlice.add(fmt.Errorf("%w: cmds", ErrStepTypeMissingRequired))
	}
	return
}

func (s *StepContainer) unmarshalNodes(nodes nodeMapUnmarshaller) (errSlice errorSlice) {
	errSlice.addNonNils(
		nodes.unmarshalString("image", &s.Image),
		nodes.unmarshalString("os", &s.OS),
		nodes.unmarshalString("shell", &s.Shell),
		nodes.unmarshalString("secretName", &s.SecretName),
		nodes.unmarshalString("serviceAccount", &s.ServiceAccount),
		nodes.unmarshalString("certificatesMountPath", &s.CertificatesMountPath),
	)
	errSlice.add(nodes.unmarshalStringSlice("cmds", &s.Cmds)...)
	return
}

func (s *StepContainer) resetDefaults() errorSlice {
	s.OS = "linux"
	s.Shell = "/bin/sh"
	s.ServiceAccount = "default"
	return nil
}

type StepDocker struct {
	// Required fields
	File string
	Tag  string

	// Optional fields
	Destination string
	Name        string
	Group       string
	Context     string
	Secret      string
	Registry    string
	AppendCert  bool
	Push        bool
	Args        []string
}

func (StepDocker) StepTypeName() string { return "docker" }

func (s StepDocker) Validate() (errSlice errorSlice) {
	if s.File == "" {
		errSlice.add(fmt.Errorf("%w: file", ErrStepTypeMissingRequired))
	}
	if s.Tag == "" {
		errSlice.add(fmt.Errorf("%w: tag", ErrStepTypeMissingRequired))
	}
	return
}

func (s *StepDocker) unmarshalNodes(nodes nodeMapUnmarshaller) (errSlice errorSlice) {
	errSlice.addNonNils(
		nodes.unmarshalString("file", &s.File),
		nodes.unmarshalString("tag", &s.Tag),
		nodes.unmarshalString("destination", &s.Destination),
		nodes.unmarshalString("name", &s.Name),
		nodes.unmarshalString("group", &s.Group),
		nodes.unmarshalString("context", &s.Context),
		nodes.unmarshalString("secret", &s.Secret),
		nodes.unmarshalString("registry", &s.Registry),
		nodes.unmarshalBool("append-cert", &s.AppendCert),
		nodes.unmarshalBool("push", &s.Push),
	)
	errSlice.add(nodes.unmarshalStringSlice("args", &s.Args)...)
	return
}

func (s *StepDocker) resetDefaults() errorSlice {
	s.Destination = ""  // TODO: default to "${registry}/${group}/${REPO_NAME}/${step_name}"
	s.Name = ""         // TODO: default to "${step_name}"
	s.Group = ""        // TODO: default to "${REPO_GROUP}"
	s.Registry = ""     // TODO: default to "${REG_URL}"
	s.AppendCert = true // TODO: default to true if REPO_GROUP starts with "default", case insensitive

	s.Push = true
	return nil
}

type StepHelm struct{}

func (StepHelm) StepTypeName() string { return "helm" }

type StepHelmPackage struct {
	// Optional fields
	Version     string
	ChartPath   string
	Destination string
}

func (StepHelmPackage) StepTypeName() string { return "helm-package" }

func (s StepHelmPackage) Validate() (errSlice errorSlice) {
	return
}

func (s *StepHelmPackage) unmarshalNodes(nodes nodeMapUnmarshaller) (errSlice errorSlice) {
	errSlice.addNonNils(
		nodes.unmarshalString("version", &s.Version),
		nodes.unmarshalString("chart-path", &s.ChartPath),
		nodes.unmarshalString("destination", &s.Destination),
	)
	return
}

func (s *StepHelmPackage) resetDefaults() errorSlice {
	s.Destination = "" // TODO: default to "${CHART_REPO}/${REPO_GROUP}"
	return nil
}

type StepKubectl struct{}

func (StepKubectl) StepTypeName() string { return "kubectl" }

type StepNuGetPackage struct{}

func (StepNuGetPackage) StepTypeName() string { return "nuget-package" }
