package modelconv

import (
	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerhttpserver/model/response"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
)

// StepsToResponseSteps converts a slice of wharfyml Steps so it can be used
// in an HTTP response.
func StepsToResponseSteps(steps []wharfyml.Step) []response.Step {
	result := make([]response.Step, 0, len(steps))
	for _, v := range steps {
		result = append(result, StepToResponseStep(v))
	}

	return result
}

// StepToResponseStep converts a wharfyml Step so it can be used in an HTTP
// response.
func StepToResponseStep(step wharfyml.Step) response.Step {
	return response.Step{
		Name:     step.Name,
		StepType: StepTypeToResponseStepType(step.Type),
	}
}

// StepTypeToResponseStepType converts a wharfyml StepType so it can be used in
// an HTTP response.
func StepTypeToResponseStepType(stepType wharfyml.StepType) response.StepType {
	return response.StepType{
		Name: stepType.StepTypeName(),
	}
}
