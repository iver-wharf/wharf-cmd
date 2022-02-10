package wharfyml

import (
	"strings"
	"testing"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
  myKubectlStep:
    kubectl:
      file: deploy/pod.yaml
`))
	assertNoErr(t, errs)

	want := Definition{
		Inputs: []Input{
			InputString{
				Name:    "myStringVar",
				Default: "foo bar",
			},
			InputPassword{
				Name:    "myPasswordVar",
				Default: "supersecret",
			},
			InputNumber{
				Name:    "myNumberVar",
				Default: 123,
			},
			InputChoice{
				Name:    "myChoiceVar",
				Default: "A",
				Values:  []string{"A", "B", "C"},
			},
		},
		Envs: map[string]Env{
			"myEnvA": {
				Name: "myEnvA",
				Vars: map[string]interface{}{
					"myString": "foo bar",
					"myUint":   uint64(123),
					"myInt":    int64(-123),
					"myFloat":  123.45,
					"myBool":   true,
				},
			},
			"myEnvB": {
				Name: "myEnvB",
				Vars: map[string]interface{}{
					"myString": "foo bar",
					"myUint":   uint64(123),
					"myInt":    int64(-123),
					"myFloat":  123.45,
					"myBool":   true,
				},
			},
		},
		Stages: []Stage{
			{
				Name: "myStage1",
				Envs: []string{"myEnvA"},
				Steps: []Step{
					{
						Name: "myDockerStep",
						Type: StepDocker{
							File: "Dockerfile",
							Tag:  "latest",
							// Values from defaults:
							AppendCert: true,
							Push:       true,
							Secret:     "gitlab-registry",
						},
					},
					{
						Name: "myContainerStep",
						Type: StepContainer{
							Image: "alpine:latest",
							Cmds:  []string{"echo hello"},
							// Values from defaults:
							OS:             "linux",
							ServiceAccount: "default",
							Shell:          "/bin/sh",
						},
					},
				},
			},
			{
				Name: "myStage2",
				Steps: []Step{
					{
						Name: "myKubectlStep",
						Type: StepKubectl{
							File: "deploy/pod.yaml",
							// Values from defaults:
							Cluster: "kubectl-config",
							Action:  "apply",
						},
					},
				},
			},
		},
	}
	assert.Equal(t, want, got)
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
			got, errs := Parse(strings.NewReader(tc.input))
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
			got, errs := Parse(strings.NewReader(tc.input))
			assertNoErr(t, errs)
			require.Len(t, got.Stages, 1)
			var gotOrder []string
			for _, s := range got.Stages[0].Steps {
				gotOrder = append(gotOrder, s.Name)
			}
			assert.Equal(t, tc.wantOrder, gotOrder)
		})
	}
}

func TestParser_TooManyDocs(t *testing.T) {
	_, errs := Parse(strings.NewReader(`
a: 1
---
b: 2
---
c: 3
`))
	requireContainsErr(t, errs, ErrTooManyDocs)
}

func TestParser_OneDocWithDocSeparator(t *testing.T) {
	_, errs := Parse(strings.NewReader(`
---
c: 3
`))
	requireNotContainsErr(t, errs, ErrTooManyDocs)
}

func TestParser_MissingDoc(t *testing.T) {
	_, errs := Parse(strings.NewReader(``))
	requireContainsErr(t, errs, ErrMissingDoc)
}

func TestParser_ErrIfDocNotMap(t *testing.T) {
	_, errs := Parse(strings.NewReader(`123`))
	requireContainsErr(t, errs, ErrNotMap)
}

func TestParser_ErrIfNonStringKey(t *testing.T) {
	_, errs := Parse(strings.NewReader(`
123: {}
`))
	requireContainsErr(t, errs, ErrKeyNotString)
}

func TestParser_ErrIfEmptyStageName(t *testing.T) {
	_, errs := Parse(strings.NewReader(`
"": {}
`))
	requireContainsErr(t, errs, ErrKeyEmpty)
}

// TODO: Test the following:
// - error on unused environment
// - error on use of undeclared environment
// - error on use of undeclared variable
// - error on multiple YAML documents (sep by three dashes)
// - can use aliases and anchors on stages
// - can use aliases and anchors on steps

func getKeyedNode(t *testing.T, content string) *ast.MappingValueNode {
	mapValNode, ok := getNode(t, content).(*ast.MappingValueNode)
	require.True(t, ok, "testing content did not parse as a MappingValueNode")
	return mapValNode
}

func getNode(t *testing.T, content string) ast.Node {
	file, err := parser.ParseBytes([]byte(content), parser.Mode(0))
	require.NoError(t, err, "parse keyed node")
	require.Len(t, file.Docs, 1, "document count")
	return file.Docs[0].Body
}
