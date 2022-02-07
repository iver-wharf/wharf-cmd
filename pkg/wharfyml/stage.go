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

func parseStage2(key *ast.StringNode, node ast.Node) (stage Stage, errSlice []error) {
	if key.Value == "" {
		errSlice = append(errSlice, wrapParseErrNode(ErrStageEmptyName, key))
		// Continue, its not a fatal issue
	}
	if node.Type() != ast.MappingType {
		errSlice = append(errSlice, wrapParseErrNode(fmt.Errorf("stage type: %s: %w", node.Type(), ErrStageNotMap), node))
		return
	}
	m := node.(*ast.MappingNode)
	if len(m.Values) == 0 {
		errSlice = append(errSlice, wrapParseErrNode(ErrStageEmpty, node))
	}
	return
}
