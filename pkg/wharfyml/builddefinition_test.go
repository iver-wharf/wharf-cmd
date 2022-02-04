package wharfyml

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type getStageWithReplacementTestSuite struct {
	suite.Suite
	sut BuildDefinition
}

func TestGetStageWithReplacement(t *testing.T) {
	suite.Run(t, new(getStageWithReplacementTestSuite))
}

func (suite *getStageWithReplacementTestSuite) SetupSuite() {
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

	sut, err := parseContent(buildDef)
	require.Nil(suite.T(), err)

	suite.sut = sut
}

func (suite *getStageWithReplacementTestSuite) TestGetStageWithReplacement() {
	stage, err := suite.sut.GetStageWithReplacement("deploy", "stage")
	require.Nil(suite.T(), err)

	require.Equal(suite.T(), "deploy", stage.Name)
	require.Equal(suite.T(), 2, len(stage.Steps))

	for _, s := range stage.Steps {
		assert.Equal(suite.T(), "stage", s.Variables["namespace"])
		assert.Equal(suite.T(), "stage-config", s.Variables["cluster"])
		if s.Name == "helm" {
			assert.Equal(suite.T(), "wharf-stage", s.Variables["name"])
			assert.Equal(suite.T(), "${CHART_REPO}/tools", s.Variables["repo"])
		}
	}
}

func (suite *getStageWithReplacementTestSuite) TestGetStageWithReplacementInvalidEnv() {
	_, err := suite.sut.GetStageWithReplacement("deploy", "test")
	require.NotNil(suite.T(), err)
}

func (suite *getStageWithReplacementTestSuite) TestGetStageWithReplacementInvalidStage() {
	_, err := suite.sut.GetStageWithReplacement("build", "stage")
	require.NotNil(suite.T(), err)
}

func TestParse_PreservesStageOrder(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		wantOrder []string
	}{
		{
			name: "A-B-C",
			input: `
A:
	myStep:
    helm-package: {}
B:
	myStep:
    helm-package: {}
C:
	myStep:
    helm-package: {}
`,
			wantOrder: []string{"A", "B", "C"},
		},
		{
			name: "B-A-C",
			input: `
B:
	myStep:
    helm-package: {}
A:
	myStep:
    helm-package: {}
C:
	myStep:
    helm-package: {}
`,
			wantOrder: []string{"B", "A", "C"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, errs := parse(strings.NewReader(tc.input))
			require.Empty(t, errs)
			var gotOrder []string
			for _, s := range got.Stages {
				gotOrder = append(gotOrder, s.Name)
			}
			assert.Equal(t, tc.wantOrder, gotOrder)
		})
	}
}

func TestParse_PreservesStepOrder(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		wantOrder []string
	}{
		{
			name: "A-B-C",
			input: `
myStage:
  A:
    helm-package: {}
  B:
    helm-package: {}
  C:
    helm-package: {}
`,
			wantOrder: []string{"A", "B", "C"},
		},
		{
			name: "B-A-C",
			input: `
myStage:
  B:
    helm-package: {}
  A:
    helm-package: {}
  C:
    helm-package: {}
`,
			wantOrder: []string{"B", "A", "C"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, errs := parse(strings.NewReader(tc.input))
			require.Empty(t, errs)
			require.Len(t, got.Stages, 1)
			var gotOrder []string
			for _, s := range got.Stages[0].Steps {
				gotOrder = append(gotOrder, s.Name)
			}
			assert.Equal(t, tc.wantOrder, gotOrder)
		})
	}
}

// TODO: Test the following:
// - error on stage not YAML map
// - error on step not YAML map
// - error on step with multiple step types
// - error on empty step
// - error on missing required fields on a per-step-type basis
// - error on missing required fields on a per-step-type basis
// - error on empty stages
// - error on unused environment
// - error on use of undeclared environment
// - error on invalid environment variable type
// - error on invalid input variable def
// - error on use of undeclared variable
// - error on multiple YAML documents (sep by three dashes)
//
// TODO: Create issue on using https://pkg.go.dev/github.com/goccy/go-yaml
// instead to be able to annotate errors with line numbers, to be able
// to add a `wharf-cmd lint` option
