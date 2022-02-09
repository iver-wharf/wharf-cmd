package wharfyml

import (
	"strings"
	"testing"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
// - error on invalid input variable def
// - error on use of undeclared variable
// - error on multiple YAML documents (sep by three dashes)
// - can use aliases and anchors on stages
// - can use aliases and anchors on steps
//
// TODO: Create issue on using https://pkg.go.dev/github.com/goccy/go-yaml
// instead to be able to annotate errors with line numbers, to be able
// to add a `wharf-cmd lint` option

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
