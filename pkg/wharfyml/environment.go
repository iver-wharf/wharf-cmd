package wharfyml

import (
	"errors"
	"fmt"
	"math"

	"github.com/goccy/go-yaml/ast"
)

// Errors related to parsing environments.
var (
	ErrEnvInvalidVarType = errors.New("invalid environment variable type")
	ErrStageEnvsNotArray = errors.New("stage environments should be a YAML array")
	ErrStageEnvNotString = errors.New("stage environment element should be a YAML string")
	ErrStageEnvEmpty     = errors.New("environment name cannot be empty")
)

// Env is an environments definition.
type Env struct {
	Name string
	Vars map[string]interface{}
}

func visitDocEnvironmentsNode(node ast.Node) (map[string]Env, Errors) {
	nodes, err := parseMappingValueNodes(node)
	if err != nil {
		return nil, Errors{err}
	}
	envs := make(map[string]Env, len(nodes))
	var errSlice Errors
	for _, n := range nodes {
		key, err := parseMapKeyNonEmpty(n.Key)
		if err != nil {
			errSlice.add(err)
			continue
		}
		env, errs := visitEnvironmentNode(key, n.Value)
		envs[key] = env
		errSlice.add(wrapPathErrorSlice(key, errs)...)
	}
	return envs, errSlice
}

func visitEnvironmentNode(name string, node ast.Node) (env Env, errSlice Errors) {
	env = Env{
		Name: name,
		Vars: make(map[string]interface{}),
	}
	nodes, err := parseMappingValueNodes(node)
	if err != nil {
		errSlice.add(err)
		return
	}
	for _, n := range nodes {
		key, err := parseMapKeyNonEmpty(n.Key)
		if err != nil {
			errSlice.add(err)
			continue
		}
		val, errs := visitEnvironmentVariableNode(n.Value)
		errSlice.add(wrapPathErrorSlice(key, errs)...)
		env.Vars[key] = val
	}
	return
}

func visitEnvironmentVariableNode(node ast.Node) (interface{}, Errors) {
	var errSlice Errors
	switch n := node.(type) {
	case *ast.BoolNode:
		return n.Value, errSlice
	case *ast.IntegerNode:
		return n.Value, errSlice // int64 or uint64
	case *ast.InfinityNode:
		return n.Value, errSlice
	case *ast.NanNode:
		return math.NaN(), errSlice
	case *ast.FloatNode:
		return n.Value, errSlice
	case *ast.StringNode:
		return n.Value, errSlice
	default:
		errSlice.add(newPositionedErrorNode(fmt.Errorf(
			"%w: expected string, boolean, or number, but found %s",
			ErrEnvInvalidVarType, prettyNodeTypeName(node)), node))
		return nil, errSlice
	}
}

func visitStageEnvironmentsNode(node ast.Node) (envs []string, errSlice Errors) {
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
