package wharfyml

import (
	"errors"

	"github.com/goccy/go-yaml/ast"
)

// Errors related to parsing environments.
var (
	ErrStageEnvsNotArray = errors.New("stage environments should be a YAML array")
	ErrStageEnvNotString = errors.New("stage environment element should be a YAML string")
	ErrStageEnvEmpty     = errors.New("environment name cannot be empty")
)

// Env is an environments definition.
type Env struct {
	Name string
	Vars map[string]interface{}
}

func visitEnvironmentMapsNode(node ast.Node) (map[string]Env, Errors) {
	return nil, Errors{errors.New("not yet implemented")}
}

func visitEnvironmentStringsNode(node ast.Node) (envs []string, errSlice Errors) {
	if node.Type() != ast.SequenceType {
		return nil, Errors{newPositionedErrorNode(ErrStageEnvsNotArray, node)}
	}
	seq := node.(*ast.SequenceNode)
	envs = make([]string, 0, len(seq.Values))
	for _, envNode := range seq.Values {
		envStrNode, ok := envNode.(*ast.StringNode)
		if !ok {
			errSlice.add(newPositionedErrorNode(ErrStageEnvNotString, envNode))
			continue
		}
		if envStrNode.Value == "" {
			errSlice.add(newPositionedErrorNode(ErrStageEnvEmpty, envNode))
			continue
		}
		envs = append(envs, envStrNode.Value)
	}
	return
}
