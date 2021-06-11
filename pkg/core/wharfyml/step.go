package wharfyml

import (
	"fmt"
	"strings"
)

type Step struct {
	Name      string
	Type      StepType
	Variables map[string]interface{}
}

func parseStep(name string, content map[string]interface{}) (Step, error) {
	if len(content) != 1 {
		return Step{}, fmt.Errorf("expected single step-type, got %d", len(content))
	}

	var stepType StepType
	var variables map[string]interface{}
	for k, v := range content {
		stepType = ParseStepType(k)
		variables = v.(map[string]interface{})
	}

	return Step{Name: name, Type: stepType, Variables: variables}, nil
}

func (step Step) GetImage() (string, error) {
	switch step.Type {
	case Container:
		return step.Variables["image"].(string), nil
	case Kaniko:
		return "boolman/kaniko:busybox-latest", nil
	case HelmPackage:
		return "wharfse/helm:latest", nil
	case HelmDeploy:
		return "wharfse/helm:latest", nil
	}

	return "", fmt.Errorf("cannot translate %s step type to image name", step.Type)
}

func (step Step) GetCommand() ([]string, error) {
	switch step.Type {
	case Container:
		vars := step.Variables["cmds"].([]interface{})
		strVars := []string{}
		for _, v := range vars {
			strVars = append(strVars, v.(string))
		}
		return []string{"/bin/sh", "-c", strings.Join(strVars, ";")}, nil
	}

	return []string{}, nil
}
