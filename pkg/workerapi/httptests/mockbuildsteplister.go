package httptests

import "github.com/iver-wharf/wharf-cmd/pkg/wharfyml"

type mockBuildStepLister struct{}

const (
	buildStepName1 = "step-1"
	buildStepName2 = "step-2"
	buildStepType1 = "container"
	buildStepType2 = "container"
)

func (b *mockBuildStepLister) ListBuildSteps() []wharfyml.Step {
	return []wharfyml.Step{
		{
			Name: buildStepName1,
			Type: wharfyml.StepContainer{},
		},
		{
			Name: buildStepName2,
			Type: wharfyml.StepContainer{},
		},
	}
}
