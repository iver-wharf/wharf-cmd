package main

import "github.com/iver-wharf/wharf-cmd/pkg/wharfyml"

type mockBuildStepLister struct {
}

func (b *mockBuildStepLister) ListBuildSteps() []wharfyml.Step {
	return []wharfyml.Step{
		{
			Pos: wharfyml.Pos{
				Line:   34,
				Column: 35,
			},
			Name: "step-1",
			Type: wharfyml.StepContainer{
				Image: "ubuntu:20.04",
			},
		},
		{
			Pos: wharfyml.Pos{
				Line:   399,
				Column: 21,
			},
			Name: "step-2",
			Type: wharfyml.StepContainer{
				Image: "ubuntu:20.04",
			},
		},
	}
}
