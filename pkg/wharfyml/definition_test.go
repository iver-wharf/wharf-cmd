package wharfyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterStagesOnEnv(t *testing.T) {
	tests := []struct {
		name      string
		envFilter string
		stages    []Stage
		want      []string
	}{
		{
			name:      "no filter skips all with env refs",
			envFilter: "",
			stages: []Stage{
				{Name: "with-envs", Envs: []EnvRef{{Name: "env1"}}},
				{Name: "with-empty-envs", Envs: []EnvRef{{Name: ""}}},
				{Name: "no-envs", Envs: nil},
			},
			want: []string{"no-envs"},
		},
		{
			name:      "with filter includes without env refs",
			envFilter: "env1",
			stages: []Stage{
				{Name: "no-envs1", Envs: nil},
				{Name: "no-envs2", Envs: nil},
				{Name: "no-envs3", Envs: nil},
			},
			want: []string{"no-envs1", "no-envs2", "no-envs3"},
		},
		{
			name:      "with filter only includes matching stages",
			envFilter: "env1",
			stages: []Stage{
				{Name: "with-env1", Envs: []EnvRef{{Name: "env1"}}},
				{Name: "with-env2", Envs: []EnvRef{{Name: "env2"}}},
				{Name: "with-env1-and-env2", Envs: []EnvRef{
					{Name: "env1"},
					{Name: "env2"},
				}},
			},
			want: []string{"with-env1", "with-env1-and-env2"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := filterStagesOnEnv(tc.stages, tc.envFilter)
			var gotNames []string
			for _, stage := range got {
				gotNames = append(gotNames, stage.Name)
			}
			assert.Equal(t, tc.want, gotNames)
		})
	}
}
