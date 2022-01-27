package builder

import (
	"errors"

	"github.com/iver-wharf/wharf-cmd/pkg/core/wharfyml"
)

type k8sStepRunner struct {
}

func NewK8sStepRunner() StepRunner {
	return k8sStepRunner{}
}

func (r k8sStepRunner) RunStep(step wharfyml.Step) (StepResult, error) {
	// TODO: Run step as Kubernetes pod
	return StepResult{}, errors.New("not implemented")
}
