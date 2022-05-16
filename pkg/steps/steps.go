package steps

import (
	"errors"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/config"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	v1 "k8s.io/api/core/v1"
)

// Errors specific to parsing step types.
var (
	ErrStepTypeUnknown = errors.New("unknown step type")
)

var (
	commonContainerName   = "step"
	commonRepoVolumeMount = v1.VolumeMount{
		Name:      "repo",
		MountPath: "/mnt/repo",
	}

	podInitWaitArgs     = []string{"/bin/sh", "-c", "sleep infinite || true"}
	podInitContinueArgs = []string{"killall", "-s", "SIGINT", "sleep"}

	errIllegalParentDirAccess = errors.New("illegal parent directory access")

	basePodSpec = v1.PodSpec{
		ServiceAccountName: "wharf-cmd",
		RestartPolicy:      v1.RestartPolicyNever,
		InitContainers: []v1.Container{
			{
				Name:            "init",
				Image:           "alpine:3",
				ImagePullPolicy: v1.PullIfNotPresent,
				Command:         podInitWaitArgs,
				VolumeMounts: []v1.VolumeMount{
					commonRepoVolumeMount,
				},
			},
		},
		Volumes: []v1.Volume{
			{
				Name: commonRepoVolumeMount.Name,
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				},
			},
		},
	}
)

func newBasePodSpec() v1.PodSpec {
	return basePodSpec
}

// StepType is an interface that is implemented by all step types.
type StepType interface {
	StepTypeName() string

	PodSpecer
}

// PodSpecer is a type that can return a Kubernetes Pod specification.
type PodSpecer interface {
	PodSpec() *v1.PodSpec
}

// Factory is the default factory implementation using the default hardcoded
// configs.
var Factory wharfyml.StepTypeFactory = factory{config: &config.DefaultConfig}

// NewFactory creates a new step type factory using the provided config.
func NewFactory(config *config.Config) wharfyml.StepTypeFactory {
	return factory{config: config}
}

type factory struct {
	config *config.Config
}

func (f factory) NewStepType(stepTypeName, stepName string, v visit.MapVisitor) (wharfyml.StepType, errutil.Slice) {
	var step interface {
		init(stepName string, v visit.MapVisitor) (StepType, errutil.Slice)
	}
	switch stepTypeName {
	case "container":
		step = Container{instanceID: f.config.InstanceID}
	case "docker":
		step = Docker{config: &f.config.Worker.Steps.Docker, instanceID: f.config.InstanceID}
	case "helm":
		step = Helm{config: &f.config.Worker.Steps.Helm}
	case "helm-package":
		step = HelmPackage{}
	case "kubectl":
		step = Kubectl{config: &f.config.Worker.Steps.Kubectl}
	case "nuget-package":
		step = NuGetPackage{}
	default:
		return nil, errutil.Slice{ErrStepTypeUnknown}
	}
	return step.init(stepName, v)
}
