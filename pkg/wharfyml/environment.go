package wharfyml

import (
	"errors"
	"fmt"

	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
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
	Vars   map[string]VarSubNode
}

// VarSource returns a varsub.Source compliant value of the environment
// variables.
func (e Env) VarSource() varsub.Source {
	source := make(varsub.SourceMap)
	name := fmt.Sprintf(".wharf-ci.yml, environment %q", e.Name)
	for k, v := range e.Vars {
		source[k] = varsub.Val{
			Value:  v,
			Source: name,
		}
	}
	return source
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
		errSlice.add(wrapPathErrorSlice(errs, n.key.value)...)
	}
	return envs, errSlice
}

func visitEnvironmentNode(nameNode strNode, node *yaml.Node) (env Env, errSlice Errors) {
	env = Env{
		Name:   nameNode.value,
		Vars:   make(map[string]VarSubNode),
		Source: newPosNode(node),
	}
	nodes, errs := visitMapSlice(node)
	errSlice.add(errs...)
	for _, n := range nodes {
		if err := verifyEnvironmentVariableNode(n.value); err != nil {
			errSlice.add(wrapPathError(err, n.key.value))
		}
		env.Vars[n.key.value] = VarSubNode{n.value}
	}
	return
}

func verifyEnvironmentVariableNode(node *yaml.Node) error {
	return verifyKind(node, "string, boolean, or number", yaml.ScalarNode)
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
			errSlice.add(wrapPosErrorNode(ErrStageEnvEmpty, envNode))
			continue
		}
		envs = append(envs, EnvRef{
			Source: newPosNode(envNode),
			Name:   env,
		})
	}
	return
}
