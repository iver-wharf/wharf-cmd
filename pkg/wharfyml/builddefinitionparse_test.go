package wharfyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSimpleBuildDefinition(t *testing.T) {
	type testCase struct {
		name     string
		buildDef string
		env      bool
	}

	tests := []testCase{
		{
			name: "Parse simple build definition",
			buildDef: `
stage1:
  step1:
    container:
      image: busybox:latest
      cmds:
      - echo hello world
`,
			env: false,
		},
		{
			name: "Parse build definition with environments",
			buildDef: `
stage1:
  environments: [prod]
  step1:
    container:
      image: busybox:latest
      cmds:
      - echo hello world
`,
			env: true,
		}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseContent(tc.buildDef)
			require.Nil(t, err)

			require.Equal(t, 1, len(got.Stages))

			stage1 := got.Stages["stage1"]
			require.Equal(t, 1, len(stage1.Steps))

			step1 := stage1.Steps[0]

			if tc.env {
				require.Equal(t, 1, len(stage1.Environments))
				assert.Equal(t, "prod", stage1.Environments[0])
			}

			assert.Equal(t, StepType(Container), step1.Type)
			assert.Equal(t, "step1", step1.Name)

			require.Equal(t, 2, len(step1.Variables))
			assert.Equal(t, "busybox:latest", step1.Variables["image"])

			require.NotNil(t, step1.Variables["cmds"])

			commands := step1.Variables["cmds"].([]interface{})
			require.Equal(t, 1, len(commands))
			assert.Equal(t, "echo hello world", commands[0])
		})
	}
}

func TestParseInputs(t *testing.T) {
	str := `
inputs:
- name: test1
  type: string
  default: foobar
- name: test2
  type: choice
  values:
  - 1
  - 2
  - 3
  default: 1
- name: test3
  type: password
- name: test4
  type: number
  default: 5
`

	expected := []interface{}{
		InputString{
			Name:    "test1",
			Type:    String,
			Default: "foobar",
		},
		InputChoice{
			Name:    "test2",
			Type:    Choice,
			Default: float64(1),
			Values:  []interface{}{float64(1), float64(2), float64(3)},
		},
		InputPassword{
			Name:    "test3",
			Type:    Password,
			Default: "",
		},
		InputNumber{
			Name:    "test4",
			Type:    Number,
			Default: 5,
		},
	}

	definition, err := parseContent(str)
	require.Nil(t, err)
	assert.ElementsMatch(t, expected, definition.Inputs)
}

func TestWithInvalidEnvironmentsShouldFailParse(t *testing.T) {
	str := `
stage1:
  environments: [1, 2, 3]
  step1:
    container:
      image: busybox:latest
      cmds:
      - echo hello world
`

	_, err := parseContent(str)

	require.NotNil(t, err)
}

func TestParseContentWithEnvironments(t *testing.T) {
	buildDef := `
environments:
  stage:
    namespace: stage
    isProduction: false
    cluster: stage-config
    url: wharf.test.local
  prod:
    namespace: prod
    isProduction: true
    cluster: prod-config
  dev:
    namespace: wharf
    isProduction: false
    cluster: dev-config

deploy:
  environments: [dev, stage, prod]

  wharf:
    helm:
      name: wharf-${namespace}
      namespace: ${namespace}
      chart: wharf-helm
      chartVersion: 0.7.7
      repo: ${CHART_REPO}/tools
      files: ['deploy/wharf.yaml']
      cluster: ${cluster}
      helmVersion: v3.0.2

  postgres:
    kubectl:
      file: 'deploy/postgres.yml'
      namespace: ${namespace}
      cluster: ${cluster}`

	got, err := parseContent(buildDef)
	require.Nil(t, err)

	require.Equal(t, 3, len(got.Environments))

	require.Equal(t, 4, len(got.Environments["stage"].Variables))
	assert.Equal(t, "stage", got.Environments["stage"].Variables["namespace"])
	assert.Equal(t, false, got.Environments["stage"].Variables["isProduction"])
	assert.Equal(t, "stage-config", got.Environments["stage"].Variables["cluster"])
	assert.Equal(t, "wharf.test.local", got.Environments["stage"].Variables["url"])

	require.Equal(t, 3, len(got.Environments["prod"].Variables))
	assert.Equal(t, "prod", got.Environments["prod"].Variables["namespace"])
	assert.Equal(t, true, got.Environments["prod"].Variables["isProduction"])
	assert.Equal(t, "prod-config", got.Environments["prod"].Variables["cluster"])

	require.Equal(t, 3, len(got.Environments["dev"].Variables))
	assert.Equal(t, "wharf", got.Environments["dev"].Variables["namespace"])
	assert.Equal(t, false, got.Environments["dev"].Variables["isProduction"])
	assert.Equal(t, "dev-config", got.Environments["dev"].Variables["cluster"])

	require.Equal(t, 0, len(got.Inputs))

	require.Equal(t, 1, len(got.Stages))
	assert.Equal(t, "deploy", got.Stages["deploy"].Name)
	require.Equal(t, 3, len(got.Stages["deploy"].Environments))
	require.Equal(t, 2, len(got.Stages["deploy"].Steps))
}
