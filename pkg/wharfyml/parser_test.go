package wharfyml

import (
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
