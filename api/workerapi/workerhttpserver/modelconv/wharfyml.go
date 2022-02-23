package modelconv

import (
	"github.com/iver-wharf/wharf-cmd/api/workerapi/workerhttpserver/model/response"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
)

func StepsToResponseSteps(steps []wharfyml.Step) []response.Step {
	result := make([]response.Step, 0, len(steps))
	for _, v := range steps {
		result = append(result, StepToResponseStep(v))
	}

	return result
}

func StepToResponseStep(step wharfyml.Step) response.Step {
	return response.Step{
		Pos:      PosToResponsePos(step.Pos),
		Name:     step.Name,
		StepType: StepTypeToResponseStepType(step.Type),
	}
}

func StepTypeToResponseStepType(stepType wharfyml.StepType) response.StepType {
	return response.StepType{
		Name: stepType.StepTypeName(),
	}
}

func PosToResponsePos(pos wharfyml.Pos) response.Pos {
	return response.Pos{
		Line:   pos.Line,
		Column: pos.Column,
	}
}
