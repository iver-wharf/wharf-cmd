package wharfyml

import (
	"errors"
	"fmt"
	"strings"
	"testing"

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
			got, errs := Parse2(strings.NewReader(tc.input))
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
			got, errs := Parse2(strings.NewReader(tc.input))
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
	_, errs := Parse2(strings.NewReader(`
a: 1
---
b: 2
---
c: 3
`))
	requireContainsErr(t, errs, ErrTooManyDocs)
}

func TestParser_MissingDoc(t *testing.T) {
	_, errs := Parse2(strings.NewReader(``))
	requireContainsErr(t, errs, ErrMissingDoc)
}

func TestParser_ErrIfDocNotMap(t *testing.T) {
	_, errs := Parse2(strings.NewReader(`123`))
	requireContainsErr(t, errs, ErrDocNotMap)
}

func TestParser_ErrIfNonStringKey(t *testing.T) {
	_, errs := Parse2(strings.NewReader(`
123: {}
`))
	requireContainsErr(t, errs, ErrKeyNotString)
}

func TestParser_ErrIfNoStages(t *testing.T) {
	_, errs := Parse2(strings.NewReader(`
environments:
  foo:
    var: 123
`))
	requireContainsErr(t, errs, ErrMissingStages)
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

func requireContainsErr(t *testing.T, errs []error, err error) {
	for _, e := range errs {
		if errors.Is(e, err) {
			return
		}
	}
	t.Fatalf("\nexpected contains error: %q\nactual: (len=%d) %v",
		err, len(errs), errs)
}

func requireContainsErrf(t *testing.T, errs []error, err error, format string, args ...interface{}) {
	for _, e := range errs {
		if errors.Is(e, err) {
			return
		}
	}
	t.Fatalf("\nexpected contains error: %q\nactual: (len=%d) %v\nmessage: %s",
		err, len(errs), errs, fmt.Sprintf(format, args...))
}
