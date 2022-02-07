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

func parseStage2(key *ast.StringNode, node ast.Node) (stage Stage2, errSlice []error) {
	stage.Name = key.Value
	if key.Value == "" {
		errSlice = append(errSlice, wrapParseErrNode(ErrStageEmptyName, key))
		// Continue, its not a fatal issue
	}
	nodes, err := stageBodyAsNodes(node)
	if err != nil {
		errSlice = append(errSlice, err)
		return
	}
	if len(nodes) == 0 {
		errSlice = append(errSlice, wrapParseErrNode(ErrStageEmpty, node))
		return
	}
	for _, stepNode := range nodes {
		key, err := parseMapKey(stepNode.Key)
		if err != nil {
			errSlice = append(errSlice, err)
			continue
		}
		errs := parseStepNodeIntoStage(&stage, key, stepNode)
		errSlice = append(errSlice, errs...)
	}
	return
}

func parseStepNodeIntoStage(stage *Stage2, key *ast.StringNode, node *ast.MappingValueNode) []error {
	var errSlice []error
	switch key.Value {
	case propEnvironments:
		stage.Environments, errSlice = parseStageEnvironmentsNode(node)
	default:
		step, errs := parseStageStepNode(key, node)
		stage.Steps = append(stage.Steps, step)
		errSlice = errs
	}
	return errSlice
}

func parseStageEnvironmentsNode(node *ast.MappingValueNode) ([]string, []error) {
	envs, errs := parseStageEnvironments2(node.Value)
	for i, err := range errs {
		errs[i] = fmt.Errorf("environments: %w", err)
	}
	return envs, errs
}

func parseStageStepNode(key *ast.StringNode, node *ast.MappingValueNode) (Step2, []error) {
	step, errs := parseStep2(key, node.Value)
	for i, err := range errs {
		errs[i] = fmt.Errorf("step %q: %w", key.Value, err)
	}
	return step, errs
}

func stageBodyAsNodes(body ast.Node) ([]*ast.MappingValueNode, error) {
	n, ok := getMappingValueNodes(body)
	if !ok {
		return nil, wrapParseErrNode(fmt.Errorf("stage type: %s: %w", body.Type(), ErrStageNotMap), body)
	}
	return n, nil
}
