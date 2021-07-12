package wharfyml

import (
	"fmt"
	"io/ioutil"

	"github.com/iver-wharf/wharf-cmd/pkg/core/utils"
	"sigs.k8s.io/yaml"
)

type BuildDefinition struct {
	Inputs       []interface{}
	Environments map[string]Environment
	Stages       map[string]Stage
}

func (b BuildDefinition) GetStageWithReplacement(stageName string, environmentName string) (Stage, error) {
	if environmentName == "" {
		return Stage{}, fmt.Errorf("environment cannot be empty string")
	}

	stage, ok := b.Stages[stageName]
	if ok == false {
		return Stage{}, fmt.Errorf("stage not found in definition: %s", stageName)
	}

	if !stage.HasEnvironments() {
		return stage, nil
	}

	if !stage.ContainsEnvironment(environmentName) {
		return Stage{}, fmt.Errorf("environment referenced in stage %q is not declared in the build definition: %s", stageName, environmentName)
	}

	envs, ok := b.Environments[environmentName]
	if ok == false {
		log.Warn().WithString("environment", environmentName).Message("Environment not found in build definition.")
		return stage, nil
	}

	for _, step := range stage.Steps {
		for i, v := range step.Variables {
			variable, ok := v.(string)
			if ok {
				step.Variables[i] = utils.ReplaceVariables(variable, envs.Variables)
				continue
			}

			varSlice, ok := v.([]string)
			if ok {
				for j, el := range varSlice {
					varSlice[j] = utils.ReplaceVariables(el, envs.Variables)
				}
				step.Variables[i] = varSlice
			}
		}
	}

	return stage, nil
}

func Parse(path string, builtinVars map[BuiltinVar]string) (BuildDefinition, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return BuildDefinition{}, err
	}

	after := utils.ReplaceVariables(string(data), ConvertToParams(builtinVars))

	return parseContent(after)
}

func parseContent(content string) (BuildDefinition, error) {
	var definition map[string]interface{}
	err := yaml.Unmarshal([]byte(content), &definition)
	if err != nil {
		return BuildDefinition{}, err
	}

	var inputs []interface{}
	envs := map[string]Environment{}
	stages := map[string]Stage{}
	for k, v := range definition {
		if k == propEnvironments {
			for envName, variables := range v.(map[string]interface{}) {
				envs[envName] = Environment{Variables: variables.(map[string]interface{})}
			}
			continue
		}

		if k == propInput {
			for _, inputElement := range v.([]interface{}) {
				inputMap := inputElement.(map[string]interface{})
				input, err := parseInput(inputMap)
				if err != nil {
					return BuildDefinition{}, err
				}
				inputs = append(inputs, input)
			}
			continue
		}

		stages[k], err = parseStage(k, v.(map[string]interface{}))
	}

	parsed := BuildDefinition{
		Stages:       stages,
		Inputs:       inputs,
		Environments: envs,
	}

	return parsed, err
}
