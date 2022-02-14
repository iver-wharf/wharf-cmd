package wharfyml

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Errors related to parsing environments.
var (
	ErrStageEnvEmpty = errors.New("environment name cannot be empty")
)

// Env is an environments definition. Used in the root of the definition.
type Env struct {
	Source Pos
	Name   string
	Vars   map[string]interface{}
}

// EnvRef is a reference to an environments definition. Used in stages.
type EnvRef struct {
	Source Pos
	Name   string
}

func visitDocEnvironmentsNode(node *yaml.Node) (map[string]Env, Errors) {
	nodes, errs := visitMapSlice(node)
	var errSlice Errors
	errSlice.add(errs...)
	envs := make(map[string]Env, len(nodes))
	for _, n := range nodes {
		env, errs := visitEnvironmentNode(n.key, n.value)
		envs[n.key.value] = env
		errSlice.add(wrapPathErrorSlice(n.key.value, errs)...)
	}
	return envs, errSlice
}

func visitEnvironmentNode(nameNode strNode, node *yaml.Node) (env Env, errSlice Errors) {
	env = Env{
		Name:   nameNode.value,
		Vars:   make(map[string]interface{}),
		Source: newPosNode2(node),
	}
	nodes, errs := visitMapSlice(node)
	errSlice.add(errs...)
	for _, n := range nodes {
		val, err := visitEnvironmentVariableNode(n.value)
		if err != nil {
			errSlice.add(wrapPathError(n.key.value, err))
		}
		env.Vars[n.key.value] = val
	}
	return
}

func visitEnvironmentVariableNode(node *yaml.Node) (interface{}, error) {
	if err := verifyKind(node, "string, boolean, or number", yaml.ScalarNode); err != nil {
		return nil, err
	}
	switch node.ShortTag() {
	case shortTagBool:
		return visitBool(node)
	case shortTagInt:
		return visitInt(node)
	case shortTagFloat:
		return visitFloat64(node)
	case shortTagString:
		return visitString(node)
	default:
		return nil, wrapPosErrorNode2(fmt.Errorf(
			"%w: expected string, boolean, or number, but found %s",
			ErrInvalidFieldType, prettyNodeTypeName2(node)), node)
	}
}

func visitStageEnvironmentsNode(node *yaml.Node) (envs []EnvRef, errSlice Errors) {
	nodes, err := visitSequence(node)
	if err != nil {
		return nil, Errors{err}
	}
	envs = make([]EnvRef, 0, len(nodes))
	for _, envNode := range nodes {
		env, err := visitString(envNode)
		if err != nil {
			errSlice.add(err)
			continue
		}
		if env == "" {
			errSlice.add(wrapPosErrorNode2(ErrStageEnvEmpty, envNode))
			continue
		}
		envs = append(envs, EnvRef{
			Source: newPosNode2(envNode),
			Name:   env,
		})
	}
	return
}
