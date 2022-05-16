package wharfyml_test

import (
	"strings"
	"testing"

	"github.com/iver-wharf/wharf-cmd/internal/errtesting"
	"github.com/iver-wharf/wharf-cmd/internal/yamltesting"
	"github.com/iver-wharf/wharf-cmd/pkg/steps"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testVarSource = varsub.SourceMap{
	"REPO_GROUP": varsub.Val{Value: "iver-wharf"},
	"REPO_NAME":  varsub.Val{Value: "wharf-cmd"},
	"REG_URL":    varsub.Val{Value: "http://harbor.example.com"},
	"REG_SECRET": varsub.Val{Value: "docker-secret"},
	"CHART_REPO": varsub.Val{Value: "http://charts.example.com"},
	"REG_USER":   varsub.Val{Value: "admin"},
	"REG_PASS":   varsub.Val{Value: "nimda"},
}

var testArgs = wharfyml.Args{
	VarSource:       testVarSource,
	StepTypeFactory: steps.DefaultFactory,
}

func TestParse_AcceptanceTest(t *testing.T) {
	got, errs := wharfyml.Parse(strings.NewReader(`
inputs:
  - name: myStringVar
    type: string
    default: foo bar
  - name: myPasswordVar
    type: password
    default: supersecret
  - name: myNumberVar
    type: number
    default: 123
  - name: myChoiceVar
    type: choice
    default: A
    values: [A, B, C]

environments:
  myEnvA:
    myString: foo bar
    myUint: 123
    myInt: -123
    myFloat: 123.45
    myBool: true
  myEnvB:
    myString: foo bar
    myUint: 123
    myInt: -123
    myFloat: 123.45
    myBool: true

myStage1:
  environments: [myEnvA]
  myDockerStep:
    docker:
      file: Dockerfile
      tag: latest
  myContainerStep:
    container:
      image: alpine:latest
      cmds: [echo hello]

myStage2:
  environments: [myEnvA]
  myKubectlStep:
    kubectl:
      file: deploy/pod.yaml
`), wharfyml.Args{Env: "myEnvA", VarSource: testVarSource, StepTypeFactory: steps.DefaultFactory})
	errtesting.RequireNoErr(t, errs)

	assert.Len(t, got.Inputs, 4)
	if assert.IsType(t, wharfyml.InputString{}, got.Inputs["myStringVar"], `Inputs["myStringVar"]`) {
		v := got.Inputs["myStringVar"].(wharfyml.InputString)
		assert.Equal(t, "foo bar", v.Default, "myStringVar.Default")
	}
	if assert.IsType(t, wharfyml.InputPassword{}, got.Inputs["myPasswordVar"], `Inputs["myPasswordVar"]`) {
		v := got.Inputs["myPasswordVar"].(wharfyml.InputPassword)
		assert.Equal(t, "supersecret", v.Default, "myPasswordVar.Default")
	}
	if assert.IsType(t, wharfyml.InputNumber{}, got.Inputs["myNumberVar"], `Inputs["myNumberVar"]`) {
		v := got.Inputs["myNumberVar"].(wharfyml.InputNumber)
		assert.Equal(t, float64(123), v.Default, "myNumberVar.Default")
	}
	if assert.IsType(t, wharfyml.InputChoice{}, got.Inputs["myChoiceVar"], `Inputs["myChoiceVar"]`) {
		v := got.Inputs["myChoiceVar"].(wharfyml.InputChoice)
		assert.Equal(t, "A", v.Default, "myChoiceVar.Default")
		assert.Equal(t, []string{"A", "B", "C"}, v.Values, "myChoiceVar.Values")
	}

	assert.Len(t, got.Envs, 2)
	if myEnvA, ok := got.Envs["myEnvA"]; assert.True(t, ok, "myEnvA") {
		yamltesting.AssertVarSubNode(t, "foo bar", myEnvA.Vars["myString"], `myEnvA.Vars["myString"]`)
		yamltesting.AssertVarSubNode(t, 123, myEnvA.Vars["myUint"], `myEnvA.Vars["myUint"]`)
		yamltesting.AssertVarSubNode(t, -123, myEnvA.Vars["myInt"], `myEnvA.Vars["myInt"]`)
		yamltesting.AssertVarSubNode(t, 123.45, myEnvA.Vars["myFloat"], `myEnvA.Vars["myFloat"]`)
		yamltesting.AssertVarSubNode(t, true, myEnvA.Vars["myBool"], `myEnvA.Vars["myBool"]`)
	}

	if myEnvB, ok := got.Envs["myEnvB"]; assert.True(t, ok, "myEnvB") {
		yamltesting.AssertVarSubNode(t, "foo bar", myEnvB.Vars["myString"], `myEnvB.Vars["myString"]`)
		yamltesting.AssertVarSubNode(t, 123, myEnvB.Vars["myUint"], `myEnvB.Vars["myUint"]`)
		yamltesting.AssertVarSubNode(t, -123, myEnvB.Vars["myInt"], `myEnvB.Vars["myInt"]`)
		yamltesting.AssertVarSubNode(t, 123.45, myEnvB.Vars["myFloat"], `myEnvB.Vars["myFloat"]`)
		yamltesting.AssertVarSubNode(t, true, myEnvB.Vars["myBool"], `myEnvB.Vars["myBool"]`)
	}

	if assert.Len(t, got.Stages, 2) {
		myStage1 := got.Stages[0]
		assert.Equal(t, "myStage1", myStage1.Name, "myStage1.Name")
		if assert.Len(t, myStage1.Envs, 1, "myStage1.Envs") {
			assert.Equal(t, "myEnvA", myStage1.Envs[0].Name, "myStage1.Envs[0].Name")
		}

		if assert.Len(t, myStage1.Steps, 2, "myStage1.Steps") {
			assert.Equal(t, "myDockerStep", myStage1.Steps[0].Name, "myStage1.myDockerStep.Name")
			if assert.IsType(t, steps.Docker{}, myStage1.Steps[0].Type, "myStage1.myDockerStep") {
				s := myStage1.Steps[0].Type.(steps.Docker)
				assert.Equal(t, "Dockerfile", s.File)
				assert.Equal(t, "latest", s.Tag)
			}

			assert.Equal(t, "myContainerStep", myStage1.Steps[1].Name, "myStage1.myContainerStep.Name")
			if assert.IsType(t, steps.Container{}, myStage1.Steps[1].Type, "myStage1.myContainerStep") {
				s := myStage1.Steps[1].Type.(steps.Container)
				assert.Equal(t, "alpine:latest", s.Image)
				assert.Equal(t, []string{"echo hello"}, s.Cmds)
			}
		}

		myStage2 := got.Stages[1]
		assert.Equal(t, "myStage2", myStage2.Name, "myStage2.Name")
		if assert.Len(t, myStage2.Envs, 1, "myStage2.Envs") {
			assert.Equal(t, "myEnvA", myStage2.Envs[0].Name, "myStage2.Envs[0].Name")
		}

		if assert.Len(t, myStage2.Steps, 1, "myStage2.Steps") {
			assert.Equal(t, "myKubectlStep", myStage2.Steps[0].Name, "myStage2.myKubectlStep.Name")
			if assert.IsType(t, steps.Kubectl{}, myStage2.Steps[0].Type, "myStage2.myContainerStep") {
				s := myStage2.Steps[0].Type.(steps.Kubectl)
				assert.Equal(t, "deploy/pod.yaml", s.File)
			}
		}
	}
}

func TestParse_SupportsTags(t *testing.T) {
	def, errs := wharfyml.Parse(strings.NewReader(`
environments:
  myEnv:
    myStr: !!str 123
    myInt: !!int 123
`), testArgs)
	errtesting.RequireNoErr(t, errs)
	myEnv, ok := def.Envs["myEnv"]
	require.True(t, ok, "myEnv environment exists")

	yamltesting.AssertVarSubNode(t, "123", myEnv.Vars["myStr"], "myStr env var")
	yamltesting.AssertVarSubNode(t, 123, myEnv.Vars["myInt"], "myInt env var")
}

func TestParse_SupportsAnchoringStages(t *testing.T) {
	def, errs := wharfyml.Parse(strings.NewReader(`
myStage1: &reused
  myStep:
    helm-package: {}

myStage2: *reused
`), testArgs)
	errtesting.RequireNoErr(t, errs)
	require.Len(t, def.Stages, 2)
	assert.Equal(t, "myStage1", def.Stages[0].Name, "stage 1 name")
	assert.Equal(t, "myStage2", def.Stages[1].Name, "stage 2 name")

	require.Len(t, def.Stages[0].Steps, 1, "stage 1 steps")
	require.Len(t, def.Stages[1].Steps, 1, "stage 2 steps")
	assert.IsType(t, steps.HelmPackage{}, def.Stages[0].Steps[0].Type, "stage 1 step 1")
	assert.IsType(t, steps.HelmPackage{}, def.Stages[1].Steps[0].Type, "stage 2 step 1")
}

func TestParse_SupportsMergingStages(t *testing.T) {
	def, errs := wharfyml.Parse(strings.NewReader(`
myStage1: &reused
  myStep:
    helm-package: {}

myStage2:
  <<: *reused
  myOtherStep:
    helm-package: {}
`), testArgs)
	errtesting.RequireNoErr(t, errs)
	require.Len(t, def.Stages, 2)
	assert.Equal(t, "myStage1", def.Stages[0].Name, "stage 1 name")
	assert.Equal(t, "myStage2", def.Stages[1].Name, "stage 2 name")

	require.Len(t, def.Stages[0].Steps, 1, "stage 1 steps")
	require.Len(t, def.Stages[1].Steps, 2, "stage 2 steps")
	assert.IsType(t, steps.HelmPackage{}, def.Stages[0].Steps[0].Type, "stage 1 step 1")
	assert.IsType(t, steps.HelmPackage{}, def.Stages[1].Steps[0].Type, "stage 2 step 1")
	assert.IsType(t, steps.HelmPackage{}, def.Stages[1].Steps[1].Type, "stage 2 step 2")
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
			got, errs := wharfyml.Parse(strings.NewReader(tc.input), testArgs)
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
			got, errs := wharfyml.Parse(strings.NewReader(tc.input), testArgs)
			errtesting.RequireNoErr(t, errs)
			require.Len(t, got.Stages, 1)
			var gotOrder []string
			for _, s := range got.Stages[0].Steps {
				gotOrder = append(gotOrder, s.Name)
			}
			assert.Equal(t, tc.wantOrder, gotOrder)
		})
	}
}

func TestParse_TooManyDocs(t *testing.T) {
	_, errs := wharfyml.Parse(strings.NewReader(`
a: 1
---
b: 2
---
c: 3
`), wharfyml.Args{})
	errtesting.RequireContainsErr(t, errs, visit.ErrTooManyDocs)
}

func TestParse_OneDocWithDocSeparator(t *testing.T) {
	_, errs := wharfyml.Parse(strings.NewReader(`
---
c: 3
`), wharfyml.Args{})
	errtesting.RequireNotContainsErr(t, errs, visit.ErrTooManyDocs)
}

func TestParse_MissingDoc(t *testing.T) {
	_, errs := wharfyml.Parse(strings.NewReader(``), wharfyml.Args{})
	errtesting.RequireContainsErr(t, errs, visit.ErrMissingDoc)
}

func TestParse_ErrIfDocNotMap(t *testing.T) {
	_, errs := wharfyml.Parse(strings.NewReader(`123`), wharfyml.Args{})
	errtesting.RequireContainsErr(t, errs, visit.ErrInvalidFieldType)
}

func TestParse_ErrIfNonStringKey(t *testing.T) {
	_, errs := wharfyml.Parse(strings.NewReader(`
123: {}
`), wharfyml.Args{})
	errtesting.RequireContainsErr(t, errs, visit.ErrKeyNotString)
}

func TestParse_ErrIfEmptyStageName(t *testing.T) {
	_, errs := wharfyml.Parse(strings.NewReader(`
"": {}
`), wharfyml.Args{})
	errtesting.RequireContainsErr(t, errs, visit.ErrKeyEmpty)
}

func TestParse_ErrIfUseOfUnknownEnv(t *testing.T) {
	_, errs := wharfyml.Parse(strings.NewReader(`
myStage:
  environments: [myEnv]
`), wharfyml.Args{})
	errtesting.RequireContainsErr(t, errs, wharfyml.ErrUseOfUndefinedEnv)
}

func TestParse_EnvVarSub(t *testing.T) {
	testCases := []struct {
		name      string
		args      wharfyml.Args
		wantImage string
		wantCmd   string
		input     string
	}{
		{
			name:      "no env",
			args:      wharfyml.Args{StepTypeFactory: steps.DefaultFactory},
			wantImage: "${myImage}",
			wantCmd:   "${myCmd}",
			input: `
environments:
  myEnv:
    myImage: ubuntu:latest
    myCmd: echo hello world
myStage:
  myStep:
    container:
      image: ${myImage}
      cmds:
        - ${myCmd}
`,
		},
		{
			name:      "with env",
			args:      wharfyml.Args{Env: "myEnv", StepTypeFactory: steps.DefaultFactory},
			wantImage: "ubuntu:latest",
			wantCmd:   "echo hello world",
			input: `
environments:
  myEnv:
    myImage: ubuntu:latest
    myCmd: echo hello world
myStage:
  environments: [myEnv]
  myStep:
    container:
      image: ${myImage}
      cmds:
        - ${myCmd}
`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			def, errs := wharfyml.Parse(strings.NewReader(tc.input), tc.args)
			require.Empty(t, errs)
			require.Len(t, def.Stages, 1, "stage count")
			require.Len(t, def.Stages[0].Steps, 1, "step count")
			myStep, ok := def.Stages[0].Steps[0].Type.(steps.Container)
			require.True(t, ok, "step type is container")

			assert.Equal(t, tc.wantImage, myStep.Image, "container.image")
			assert.Equal(t, []string{tc.wantCmd}, myStep.Cmds, "container.cmds")
		})
	}
}
