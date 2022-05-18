package wharfyml

import (
	"testing"

	"github.com/iver-wharf/wharf-cmd/internal/errtesting"
	"github.com/iver-wharf/wharf-cmd/internal/yamltesting"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVisitDefEnvironments_ErrIfNotMap(t *testing.T) {
	_, errs := visitDocEnvironmentsNode(yamltesting.NewNode(t, `123`))
	errtesting.RequireContainsErr(t, errs, visit.ErrInvalidFieldType)
}

func TestVisitDefEnvironments_ErrIfKeyNotString(t *testing.T) {
	_, errs := visitDocEnvironmentsNode(yamltesting.NewNode(t, `123: {}`))
	errtesting.RequireContainsErr(t, errs, visit.ErrKeyNotString)
}

func TestVisitDefEnvironments_ErrIfKeyEmpty(t *testing.T) {
	_, errs := visitDocEnvironmentsNode(yamltesting.NewNode(t, `"": {}`))
	errtesting.RequireContainsErr(t, errs, visit.ErrKeyEmpty)
}

func TestVisitDefEnvironments_ValidMapOfEnvs(t *testing.T) {
	envs, _ := visitDocEnvironmentsNode(yamltesting.NewNode(t, `
myEnv1: {}
myEnv2: {}
myEnv3: {}
`))
	require.Len(t, envs, 3)
	_, hasEnv1 := envs["myEnv1"]
	_, hasEnv2 := envs["myEnv2"]
	_, hasEnv3 := envs["myEnv3"]
	assert.True(t, hasEnv1, "has myEnv1")
	assert.True(t, hasEnv2, "has myEnv2")
	assert.True(t, hasEnv3, "has myEnv3")
}

func TestVisitEnvironment_SetsName(t *testing.T) {
	env, _ := visitEnvironmentNode(yamltesting.NewKeyedNode(t, `myEnv: {}`))
	assert.Equal(t, "myEnv", env.Name)
}

func TestVisitEnvironment_ErrIfEnvNotMap(t *testing.T) {
	_, errs := visitEnvironmentNode(yamltesting.NewKeyedNode(t, `myEnv: 123`))
	errtesting.RequireContainsErr(t, errs, visit.ErrInvalidFieldType)
}

func TestVisitEnvironment_ErrIfInvalidVarType(t *testing.T) {
	_, errs := visitEnvironmentNode(yamltesting.NewKeyedNode(t, `
myEnv:
  myVar: [123]
`))
	errtesting.RequireContainsErr(t, errs, visit.ErrInvalidFieldType)
}

func TestVisitEnvironment_ValidVarTypes(t *testing.T) {
	env, errs := visitEnvironmentNode(yamltesting.NewKeyedNode(t, `
myEnv:
  myIntNeg: -123
  myIntPos: 123
  myFloat: 456.789
  myString: foo bar
  myBool: true
`))
	errtesting.RequireNotContainsErr(t, errs, visit.ErrInvalidFieldType)
	want := map[string]any{
		"myIntNeg": -123,
		"myIntPos": 123,
		"myFloat":  456.789,
		"myString": "foo bar",
		"myBool":   true,
	}
	for k, wantValue := range want {
		yamltesting.AssertVarSubNode(t, wantValue, env.Vars[k], "env.Vars[%q]", k)
	}
}

func TestVisitStageEnvironments_ErrIfNotArray(t *testing.T) {
	_, errs := visitStageEnvironmentsNode(yamltesting.NewNode(t, `123`))
	errtesting.RequireContainsErr(t, errs, visit.ErrInvalidFieldType)
}

func TestVisitStageEnvironments_Valid(t *testing.T) {
	envs, errs := visitStageEnvironmentsNode(yamltesting.NewNode(t, `[a, b, c]`))
	if len(errs) > 0 {
		t.Logf("errs: %v", errs)
	}
	want := []string{"a", "b", "c"}
	got := make([]string, len(envs))
	for i, env := range envs {
		got[i] = env.Name
	}
	assert.Equal(t, want, got)
}
