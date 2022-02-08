package wharfyml

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
)

var (
	ErrStepTypeUnknown         = errors.New("unknown step type")
	ErrStepTypeInvalidField    = errors.New("invalid field type")
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

func visitStepTypeNode(key *ast.StringNode, node ast.Node) (StepType2, errorSlice) {
	stepType, err := unmarshalStepTypeNode(key, node)
	if err != nil {
		return nil, errorSlice{err}
	}
	// TODO: unmarshal explicitly to keep node references in validation errors
	return stepType, stepType.Validate()
}

func unmarshalStepTypeNode(key *ast.StringNode, node ast.Node) (StepType2, error) {
	var stepType StepType2
	var err error
	switch key.Value {
	case "container":
		container := DefaultStepContainer
		err = yamlUnmarshalNodeWithValidator(node, &container)
		stepType = container
	case "docker":
		docker := DefaultStepDocker
		err = yamlUnmarshalNodeWithValidator(node, &docker)
		stepType = docker
	case "helm-package":
		helmPackage := DefaultStepHelmPackage
		err = yamlUnmarshalNodeWithValidator(node, &helmPackage)
		stepType = helmPackage
	case "helm", "kubectl", "nuget-package":
		return nil, errors.New("not yet implemented")
	default:
		return nil, newParseErrorNode(ErrStepTypeUnknown, key)
	}
	if err != nil {
		return nil, newParseErrorNode(fmt.Errorf("%w: %v", ErrStepTypeInvalidField, err), node)
	}
	return stepType, nil
}

func yamlUnmarshalNodeWithValidator(node ast.Node, valuePtr interface{}) error {
	var buf bytes.Buffer
	return yaml.NewDecoder(&buf).DecodeFromNode(node, valuePtr)
}

type StepContainer struct {
	Image string   `yaml:"image"`
	Cmds  []string `yaml:"cmds"`

	OS                    string `yaml:"os"`
	Shell                 string `yaml:"shell"`
	SecretName            string `yaml:"secretName"`
	ServiceAccount        string `yaml:"serviceAccount"`
	CertificatesMountPath string `yaml:"certificatesMountPath"`
}

var DefaultStepContainer = StepContainer{
	OS:             "linux",
	Shell:          "/bin/sh",
	ServiceAccount: "default",
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

type StepDocker struct {
	File string `yaml:"file"`
	Tag  string `yaml:"tag"`

	Destination string   `yaml:"destination"`
	Name        string   `yaml:"name"`
	Group       string   `yaml:"group"`
	Context     string   `yaml:"context"`
	Secret      string   `yaml:"secret"`
	Registry    string   `yaml:"registry"`
	AppendCert  bool     `yaml:"append-cert"`
	Push        bool     `yaml:"push"`
	Args        []string `yaml:"args"`
}

var DefaultStepDocker = StepDocker{
	Destination: "",   // TODO: default to "${registry}/${group}/${REPO_NAME}/${step_name}"
	Name:        "",   // TODO: default to "${step_name}"
	Group:       "",   // TODO: default to "${REPO_GROUP}"
	Registry:    "",   // TODO: default to "${REG_URL}"
	AppendCert:  true, // TODO: default to true if REPO_GROUP starts with "default", case insensitive
	Push:        true,
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

type StepHelm struct{}

func (StepHelm) StepTypeName() string { return "helm" }

type StepHelmPackage struct {
	Version     string `yaml:"version"`
	ChartPath   string `yaml:"chart-path"`
	Destination string `yaml:"destination"`
}

var DefaultStepHelmPackage = StepHelmPackage{
	Destination: "", // TODO: default to "${CHART_REPO}/${REPO_GROUP}"
}

func (StepHelmPackage) StepTypeName() string { return "helm-package" }

func (s StepHelmPackage) Validate() (errSlice errorSlice) {
	return
}

type StepKubectl struct{}

func (StepKubectl) StepTypeName() string { return "kubectl" }

type StepNuGetPackage struct{}

func (StepNuGetPackage) StepTypeName() string { return "nuget-package" }
