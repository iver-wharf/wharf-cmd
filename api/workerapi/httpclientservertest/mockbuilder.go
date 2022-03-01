package main

import (
	"context"

	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/worker"
)

type mockBuilder struct{}

func (b *mockBuilder) Build(ctx context.Context, def wharfyml.Definition, opt worker.BuildOptions) (worker.Result, error) {
	return worker.Result{}, nil
}

func (b *mockBuilder) ListBuildSteps() []wharfyml.Step {
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
