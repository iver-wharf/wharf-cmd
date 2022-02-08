package wharfyml

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml/ast"
)

var (
	ErrStageNotMap    = errors.New("stage should be a YAML map")
	ErrStageEmpty     = errors.New("stage is missing steps")
	ErrStageEmptyName = errors.New("stage name cannot be empty")
)

type Stage struct {
	Name         string
	Environments []string
	Steps        []Step
}

func (s Stage) HasEnvironments() bool {
	return len(s.Environments) > 0
}

func (s Stage) ContainsEnvironment(name string) bool {
	for _, e := range s.Environments {
		if e == name {
			return true
		}
	}
	return false
}

func parseStage(name string, content map[string]interface{}) (Stage, error) {
	stage := Stage{Name: name, Environments: []string{}, Steps: []Step{}}

	for k, v := range content {
		if k == propEnvironments {
			envs, err := parseStageEnvironments(v.([]interface{}))
			if err != nil {
				return Stage{}, err
			}

			stage.Environments = envs
			continue
		}

		step, err := parseStep(k, v.(map[string]interface{}))
		if err != nil {
			return Stage{}, err
		}

		stage.Steps = append(stage.Steps, step)
	}

	return stage, nil
}

func parseStageEnvironments(content []interface{}) ([]string, error) {
	var envs []string
	for _, v := range content {
		str, ok := v.(string)
		if !ok {
			return envs, fmt.Errorf("expected value type string, got %T", v)
		}
		envs = append(envs, str)
	}
	return envs, nil
}

// ----------------------------------------

type Stage2 struct {
	Name         string
	Environments []string
	Steps        []Step2
}

func visitStageNode(key *ast.StringNode, node ast.Node) (stage Stage2, errSlice errorSlice) {
	stage.Name = key.Value
	if key.Value == "" {
		errSlice.add(newParseErrorNode(ErrStageEmptyName, key))
		// Continue, its not a fatal issue
	}
	nodes, err := stageBodyAsNodes(node)
	if err != nil {
		errSlice.add(err)
		return
	}
	if len(nodes) == 0 {
		errSlice.add(newParseErrorNode(ErrStageEmpty, node))
		return
	}
	for _, stepNode := range nodes {
		key, err := parseMapKey(stepNode.Key)
		if err != nil {
			errSlice.add(err)
			continue
		}
		switch key.Value {
		case propEnvironments:
			envs, errs := visitStageEnvironmentsNode(stepNode)
			stage.Environments = envs
			errSlice.add(errs...)
		default:
			step, errs := visitStageStepNode(key, stepNode)
			stage.Steps = append(stage.Steps, step)
			errSlice.add(errs...)
		}
	}
	return
}

func visitStageEnvironmentsNode(node *ast.MappingValueNode) ([]string, errorSlice) {
	envs, errs := visitEnvironmentStringsNode(node.Value)
	errs.fmtErrorfAll("environments: %w", fmtErrorfPlaceholder)
	return envs, errs
}

func visitStageStepNode(key *ast.StringNode, node *ast.MappingValueNode) (Step2, errorSlice) {
	step, errs := visitStepNode(key, node.Value)
	errs.fmtErrorfAll("step %q: %w", key.Value, fmtErrorfPlaceholder)
	return step, errs
}

func stageBodyAsNodes(body ast.Node) ([]*ast.MappingValueNode, error) {
	n, ok := getMappingValueNodes(body)
	if !ok {
		return nil, newParseErrorNode(fmt.Errorf("stage type: %s: %w",
			body.Type(), ErrStageNotMap), body)
	}
	return n, nil
}
