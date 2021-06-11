package wharfyml

import "fmt"

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