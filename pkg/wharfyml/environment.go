package wharfyml

import (
	"errors"
	"fmt"
	"math"

	"github.com/goccy/go-yaml/ast"
)

// Errors related to parsing environments.
var (
	ErrEnvsNotMap        = errors.New("environments should be a YAML map")
	ErrEnvNotMap         = errors.New("environment should be a YAML map")
	ErrEnvEmptyName      = errors.New("name cannot be empty")
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

func visitEnvironmentMapsNode(node ast.Node) (map[string]Env, Errors) {
	nodes, err := envsBodyAsNodes(node)
	if err != nil {
		return nil, Errors{err}
	}
	envs := make(map[string]Env, len(nodes))
	var errSlice Errors
	for _, n := range nodes {
		key, err := parseMapKey(n.Key)
		if err != nil {
			errSlice.add(err)
			continue
		}
		env, errs := visitEnvironmentNode(key, n.Value)
		envs[key.Value] = env
		errSlice.add(wrapPathErrorSlice(key.Value, errs)...)
	}
	return envs, errSlice
}

func visitEnvironmentNode(key *ast.StringNode, node ast.Node) (env Env, errSlice Errors) {
	if key.Value == "" {
		// TODO: reuse same error for all empty key strings
		// Could be wise to inject that into parseMapKey instead
		errSlice.add(newPositionedErrorNode(ErrEnvEmptyName, key))
		// Continue, it's not a fatal issue
	}
	env = Env{
		Name: key.Value,
		Vars: make(map[string]interface{}),
	}
	nodes, err := envBodyAsNodes(node)
	if err != nil {
		errSlice.add(err)
		return
	}
	for _, n := range nodes {
		key, err := parseMapKey(n.Key)
		if err != nil {
			errSlice.add(err)
			continue
		}
		val, errs := visitEnvironmentVariableNode(key, n.Value)
		errSlice.add(wrapPathErrorSlice(key.Value, errs)...)
		env.Vars[key.Value] = val
	}
	return
}

func visitEnvironmentVariableNode(key *ast.StringNode, node ast.Node) (interface{}, Errors) {
	var errSlice Errors
	if key.Value == "" {
		errSlice.add(newPositionedErrorNode(ErrEnvEmptyName, key))
		// Continue, it's not a fatal issue
	}
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

func envsBodyAsNodes(body ast.Node) ([]*ast.MappingValueNode, error) {
	n, ok := getMappingValueNodes(body)
	if !ok {
		return nil, newPositionedErrorNode(fmt.Errorf("%s: %w", body.Type(), ErrEnvsNotMap), body)
	}
	return n, nil
}

func envBodyAsNodes(body ast.Node) ([]*ast.MappingValueNode, error) {
	n, ok := getMappingValueNodes(body)
	if !ok {
		return nil, newPositionedErrorNode(fmt.Errorf("%s: %w", body.Type(), ErrEnvNotMap), body)
	}
	return n, nil
}
