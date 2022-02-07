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
}

func parseStepType(key *ast.StringNode, node ast.Node) (StepType2, []error) {
	stepType, err := parseStepTypeNode(key, node)
	if err != nil {
		return nil, []error{err}
	}
	// TODO: validate step type
	return stepType, nil
}

func parseStepTypeNode(key *ast.StringNode, node ast.Node) (StepType2, error) {
	var stepType StepType2
	var err error
	switch key.Value {
	case "container":
		container := DefaultStepContainer
		err = yamlUnmarshalNode(node, &container)
		stepType = container
	case "docker":
		docker := DefaultStepDocker
		err = yamlUnmarshalNode(node, &docker)
		stepType = docker
	case "helm", "helm-package", "kubectl", "nuget-package":
		return nil, errors.New("not yet implemented")
	default:
		return nil, wrapParseErrNode(ErrStepTypeUnknown, key)
	}
	if err != nil {
		return nil, wrapParseErrNode(fmt.Errorf("%w: %v", ErrStepTypeInvalidField, err), node)
	}
	return stepType, nil
}

func yamlUnmarshalNode(node ast.Node, valuePtr interface{}) error {
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

type StepHelm struct{}

func (StepHelm) StepTypeName() string { return "helm" }

type StepHelmPackage struct{}

func (StepHelmPackage) StepTypeName() string { return "helm-package" }

type StepKubectl struct{}

func (StepKubectl) StepTypeName() string { return "kubectl" }

type StepNuGetPackage struct{}

func (StepNuGetPackage) StepTypeName() string { return "nuget-package" }
