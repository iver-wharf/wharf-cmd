package steps

import (
	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
)

// StepType is an interface that is implemented by all step types.
type StepType interface {
	StepTypeName() string
}

type Factory struct{}

func (Factory) NewStepType(stepTypeName, stepName string, v visit.MapVisitor) (wharfyml.StepType, errutil.Slice) {
	var step interface {
		wharfyml.StepType
		init(stepName string, v visit.MapVisitor) errutil.Slice
	}
	switch stepTypeName {
	case "container":
		step = &Container{}
	case "docker":
		step = &Docker{}
	case "helm":
		step = &Helm{}
	case "helm-package":
		step = &HelmPackage{}
	case "kubectl":
		step = &Kubectl{}
	case "nuget-package":
		step = &NuGetPackage{}
	}
	return step, step.init(stepName, v)
}
