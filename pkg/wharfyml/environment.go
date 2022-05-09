package wharfyml

import (
	"errors"
	"fmt"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	"gopkg.in/yaml.v3"
)

// errutil.Slice related to parsing environments.
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

func visitDocEnvironmentsNode(node *yaml.Node) (map[string]Env, errutil.Slice) {
	nodes, errs := visit.MapSlice(node)
	var errSlice errutil.Slice
	errSlice.Add(errs...)
	envs := make(map[string]Env, len(nodes))
	for _, n := range nodes {
		env, errs := visitEnvironmentNode(n.Key, n.Value)
		envs[n.Key.Value] = env
		errSlice.Add(errutil.ScopeSlice(errs, n.Key.Value)...)
	}
	return envs, errSlice
}

func visitEnvironmentNode(nameNode visit.StringNode, node *yaml.Node) (env Env, errSlice errutil.Slice) {
	env = Env{
		Name:   nameNode.Value,
		Vars:   make(map[string]VarSubNode),
		Source: newPosNode(node),
	}
	nodes, errs := visit.MapSlice(node)
	errSlice.Add(errs...)
	for _, n := range nodes {
		if err := verifyEnvironmentVariableNode(n.Value); err != nil {
			errSlice.Add(errutil.Scope(err, n.Key.Value))
		}
		env.Vars[n.Key.Value] = VarSubNode{n.Value}
	}
	return
}

func verifyEnvironmentVariableNode(node *yaml.Node) error {
	return visit.VerifyKind(node, "string, boolean, or number", yaml.ScalarNode)
}

func visitStageEnvironmentsNode(node *yaml.Node) (envs []EnvRef, errSlice errutil.Slice) {
	nodes, err := visit.Sequence(node)
	if err != nil {
		return nil, errutil.Slice{err}
	}
	envs = make([]EnvRef, 0, len(nodes))
	for _, envNode := range nodes {
		env, err := visit.String(envNode)
		if err != nil {
			errSlice.Add(err)
			continue
		}
		if env == "" {
			errSlice.Add(wrapPosErrorNode(ErrStageEnvEmpty, envNode))
			continue
		}
		envs = append(envs, EnvRef{
			Source: newPosNode(envNode),
			Name:   env,
		})
	}
	return
}
