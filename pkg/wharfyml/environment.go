package wharfyml

import (
	"errors"

	"github.com/goccy/go-yaml/ast"
)

var (
	ErrStageEnvsNotArray = errors.New("stage environments should be a YAML array")
	ErrStageEnvNotString = errors.New("stage environment element should be a YAML string")
	ErrStageEnvEmpty     = errors.New("environment name cannot be empty")
)

type Environment struct {
	Variables map[string]interface{}
}

type Env struct {
	Name string
	Vars map[string]interface{}
}

func parseDefEnvironments(node ast.Node) (map[string]Env, errorSlice) {
	return nil, nil
}

func parseStageEnvironments2(node ast.Node) (envs []string, errSlice errorSlice) {
	if node.Type() != ast.SequenceType {
		return nil, errorSlice{newParseErrorNode(ErrStageEnvsNotArray, node)}
	}
	seq := node.(*ast.SequenceNode)
	envs = make([]string, 0, len(seq.Values))
	for _, envNode := range seq.Values {
		envStrNode, ok := envNode.(*ast.StringNode)
		if !ok {
			errSlice.add(newParseErrorNode(ErrStageEnvNotString, envNode))
			continue
		}
		if envStrNode.Value == "" {
			errSlice.add(newParseErrorNode(ErrStageEnvEmpty, envNode))
			continue
		}
		envs = append(envs, envStrNode.Value)
	}
	return
}
