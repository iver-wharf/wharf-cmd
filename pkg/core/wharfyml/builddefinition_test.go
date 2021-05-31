package wharfyml

import (
	"github.com/iver-wharf/wharf-cmd/pkg/core/containercreator"
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

	sut, err := parseContent([]byte(buildDef))
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
