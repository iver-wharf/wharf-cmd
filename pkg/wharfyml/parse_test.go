package wharfyml

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParse_AcceptanceTest(t *testing.T) {
	got, errs := Parse(strings.NewReader(`
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
`), Args{Env: "myEnvA"})
	requireNoErr(t, errs)

	assert.Len(t, got.Inputs, 4)
	if assert.IsType(t, InputString{}, got.Inputs["myStringVar"], `Inputs["myStringVar"]`) {
		v := got.Inputs["myStringVar"].(InputString)
		assert.Equal(t, "foo bar", v.Default, "myStringVar.Default")
	}
	if assert.IsType(t, InputPassword{}, got.Inputs["myPasswordVar"], `Inputs["myPasswordVar"]`) {
		v := got.Inputs["myPasswordVar"].(InputPassword)
		assert.Equal(t, "supersecret", v.Default, "myPasswordVar.Default")
	}
	if assert.IsType(t, InputNumber{}, got.Inputs["myNumberVar"], `Inputs["myNumberVar"]`) {
		v := got.Inputs["myNumberVar"].(InputNumber)
		assert.Equal(t, float64(123), v.Default, "myNumberVar.Default")
	}
	if assert.IsType(t, InputChoice{}, got.Inputs["myChoiceVar"], `Inputs["myChoiceVar"]`) {
		v := got.Inputs["myChoiceVar"].(InputChoice)
		assert.Equal(t, "A", v.Default, "myChoiceVar.Default")
		assert.Equal(t, []string{"A", "B", "C"}, v.Values, "myChoiceVar.Values")
	}

	assert.Len(t, got.Envs, 2)
	if myEnvA, ok := got.Envs["myEnvA"]; assert.True(t, ok, "myEnvA") {
		assert.Equal(t, "foo bar", myEnvA.Vars["myString"], `myEnvA.Vars["myString"]`)
		assert.Equal(t, 123, myEnvA.Vars["myUint"], `myEnvA.Vars["myUint"]`)
		assert.Equal(t, -123, myEnvA.Vars["myInt"], `myEnvA.Vars["myInt"]`)
		assert.Equal(t, 123.45, myEnvA.Vars["myFloat"], `myEnvA.Vars["myFloat"]`)
		assert.Equal(t, true, myEnvA.Vars["myBool"], `myEnvA.Vars["myBool"]`)
	}

	if myEnvB, ok := got.Envs["myEnvB"]; assert.True(t, ok, "myEnvB") {
		assert.Equal(t, "foo bar", myEnvB.Vars["myString"], `myEnvB.Vars["myString"]`)
		assert.Equal(t, 123, myEnvB.Vars["myUint"], `myEnvB.Vars["myUint"]`)
		assert.Equal(t, -123, myEnvB.Vars["myInt"], `myEnvB.Vars["myInt"]`)
		assert.Equal(t, 123.45, myEnvB.Vars["myFloat"], `myEnvB.Vars["myFloat"]`)
		assert.Equal(t, true, myEnvB.Vars["myBool"], `myEnvB.Vars["myBool"]`)
	}

	if assert.Len(t, got.Stages, 2) {
		myStage1 := got.Stages[0]
		assert.Equal(t, "myStage1", myStage1.Name, "myStage1.Name")
		if assert.Len(t, myStage1.Envs, 1, "myStage1.Envs") {
			assert.Equal(t, "myEnvA", myStage1.Envs[0].Name, "myStage1.Envs[0].Name")
		}

		if assert.Len(t, myStage1.Steps, 2, "myStage1.Steps") {
			assert.Equal(t, "myDockerStep", myStage1.Steps[0].Name, "myStage1.myDockerStep.Name")
			if assert.IsType(t, StepDocker{}, myStage1.Steps[0].Type, "myStage1.myDockerStep") {
				s := myStage1.Steps[0].Type.(StepDocker)
				assert.Equal(t, "Dockerfile", s.File)
				assert.Equal(t, "latest", s.Tag)
			}

			assert.Equal(t, "myContainerStep", myStage1.Steps[1].Name, "myStage1.myContainerStep.Name")
			if assert.IsType(t, StepContainer{}, myStage1.Steps[1].Type, "myStage1.myContainerStep") {
				s := myStage1.Steps[1].Type.(StepContainer)
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
			if assert.IsType(t, StepKubectl{}, myStage2.Steps[0].Type, "myStage2.myContainerStep") {
				s := myStage2.Steps[0].Type.(StepKubectl)
				assert.Equal(t, "deploy/pod.yaml", s.File)
			}
		}
	}
}

func TestParse_SupportsTags(t *testing.T) {
	def, errs := Parse(strings.NewReader(`
environments:
  myEnv:
    myStr: !!str 123
    myInt: !!int 123
`), Args{})
	requireNoErr(t, errs)
	myEnv, ok := def.Envs["myEnv"]
	require.True(t, ok, "myEnv environment exists")

	assert.Equal(t, "123", myEnv.Vars["myStr"], "myStr env var")
	assert.Equal(t, 123, myEnv.Vars["myInt"], "myInt env var")
}

func TestParse_SupportsAnchoringStages(t *testing.T) {
	def, errs := Parse(strings.NewReader(`
myStage1: &reused
  myStep:
    helm-package: {}

myStage2: *reused
`), Args{})
	requireNoErr(t, errs)
	require.Len(t, def.Stages, 2)
	assert.Equal(t, "myStage1", def.Stages[0].Name, "stage 1 name")
	assert.Equal(t, "myStage2", def.Stages[1].Name, "stage 2 name")

	require.Len(t, def.Stages[0].Steps, 1, "stage 1 steps")
	require.Len(t, def.Stages[1].Steps, 1, "stage 2 steps")
	assert.IsType(t, StepHelmPackage{}, def.Stages[0].Steps[0].Type, "stage 1 step 1")
	assert.IsType(t, StepHelmPackage{}, def.Stages[1].Steps[0].Type, "stage 2 step 1")
}

func TestParse_SupportsMergingStages(t *testing.T) {
	def, errs := Parse(strings.NewReader(`
myStage1: &reused
  myStep:
    helm-package: {}

myStage2:
  <<: *reused
  myOtherStep:
    helm-package: {}
`), Args{})
	requireNoErr(t, errs)
	require.Len(t, def.Stages, 2)
	assert.Equal(t, "myStage1", def.Stages[0].Name, "stage 1 name")
	assert.Equal(t, "myStage2", def.Stages[1].Name, "stage 2 name")

	require.Len(t, def.Stages[0].Steps, 1, "stage 1 steps")
	require.Len(t, def.Stages[1].Steps, 2, "stage 2 steps")
	assert.IsType(t, StepHelmPackage{}, def.Stages[0].Steps[0].Type, "stage 1 step 1")
	assert.IsType(t, StepHelmPackage{}, def.Stages[1].Steps[0].Type, "stage 2 step 1")
	assert.IsType(t, StepHelmPackage{}, def.Stages[1].Steps[1].Type, "stage 2 step 2")
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
			got, errs := Parse(strings.NewReader(tc.input), Args{})
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
			got, errs := Parse(strings.NewReader(tc.input), Args{})
			requireNoErr(t, errs)
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
	_, errs := Parse(strings.NewReader(`
a: 1
---
b: 2
---
c: 3
`), Args{})
	requireContainsErr(t, errs, ErrTooManyDocs)
}

func TestParse_OneDocWithDocSeparator(t *testing.T) {
	_, errs := Parse(strings.NewReader(`
---
c: 3
`), Args{})
	requireNotContainsErr(t, errs, ErrTooManyDocs)
}

func TestParse_MissingDoc(t *testing.T) {
	_, errs := Parse(strings.NewReader(``), Args{})
	requireContainsErr(t, errs, ErrMissingDoc)
}

func TestParse_ErrIfDocNotMap(t *testing.T) {
	_, errs := Parse(strings.NewReader(`123`), Args{})
	requireContainsErr(t, errs, ErrInvalidFieldType)
}

func TestParse_ErrIfNonStringKey(t *testing.T) {
	_, errs := Parse(strings.NewReader(`
123: {}
`), Args{})
	requireContainsErr(t, errs, ErrKeyNotString)
}

func TestParse_ErrIfEmptyStageName(t *testing.T) {
	_, errs := Parse(strings.NewReader(`
"": {}
`), Args{})
	requireContainsErr(t, errs, ErrKeyEmpty)
}

func TestParse_ErrIfUseOfUnknownEnv(t *testing.T) {
	_, errs := Parse(strings.NewReader(`
myStage:
  environments: [myEnv]
`), Args{})
	requireContainsErr(t, errs, ErrUseOfUndefinedEnv)
}

func TestParse_EnvVarSub(t *testing.T) {
	testCases := []struct {
		name      string
		args      Args
		wantImage string
		wantCmd   string
		input     string
	}{
		{
			name:      "no env",
			args:      Args{},
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
			args:      Args{Env: "myEnv"},
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
			def, errs := Parse(strings.NewReader(tc.input), tc.args)
			require.Empty(t, errs)
			require.Len(t, def.Stages, 1, "stage count")
			require.Len(t, def.Stages[0].Steps, 1, "step count")
			myStep, ok := def.Stages[0].Steps[0].Type.(StepContainer)
			require.True(t, ok, "step type is container")

			assert.Equal(t, tc.wantImage, myStep.Image, "container.image")
			assert.Equal(t, []string{tc.wantCmd}, myStep.Cmds, "container.cmds")
		})
	}
}

func getKeyedNode(t *testing.T, content string) (strNode, *yaml.Node) {
	node := getNode(t, content)
	require.Equal(t, yaml.MappingNode, node.Kind, "keyed node")
	require.Len(t, node.Content, 2, "keyed node")
	require.Equal(t, yaml.ScalarNode, node.Content[0].Kind, "key node kind in keyed node")
	require.Equal(t, shortTagString, node.Content[0].ShortTag(), "key node tag in keyed node")
	return strNode{node: node.Content[0], value: node.Content[0].Value}, node.Content[1]
}

func getNode(t *testing.T, content string) *yaml.Node {
	var doc yaml.Node
	err := yaml.Unmarshal([]byte(content), &doc)
	require.NoError(t, err, "parse node")
	require.Equal(t, yaml.DocumentNode, doc.Kind, "document node")
	require.Len(t, doc.Content, 1, "document node count")
	return doc.Content[0]
}
