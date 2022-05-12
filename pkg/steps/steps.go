package steps

import (
	"errors"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
)

var (
	ErrStepTypeUnknown = errors.New("unknown step type")
)

// StepType is an interface that is implemented by all step types.
type StepType interface {
	StepTypeName() string
}

var Factory wharfyml.StepTypeFactory = factory{}

type factory struct{}

func (factory) NewStepType(stepTypeName, stepName string, v visit.MapVisitor) (wharfyml.StepType, errutil.Slice) {
	var step interface {
		init(stepName string, v visit.MapVisitor) (wharfyml.StepType, errutil.Slice)
	}
	switch stepTypeName {
	case "container":
		step = Container{}
	case "docker":
		step = Docker{}
	case "helm":
		step = Helm{}
	case "helm-package":
		step = HelmPackage{}
	case "kubectl":
		step = Kubectl{}
	case "nuget-package":
		step = NuGetPackage{}
	default:
		return nil, errutil.Slice{ErrStepTypeUnknown}
	}
	return step.init(stepName, v)
}
