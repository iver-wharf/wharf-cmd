package wharfyml

import (
	"testing"

	"github.com/iver-wharf/wharf-cmd/internal/testutil"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	"github.com/stretchr/testify/assert"
)

func TestVisitDocInputs_ErrIfNotArray(t *testing.T) {
	_, errs := visitInputsNode(testutil.NewNode(t, `123`))
	testutil.RequireContainsErr(t, errs, visit.ErrInvalidFieldType)
}

func TestVisitDocInputs_ErrIfNameCollision(t *testing.T) {
	_, errs := visitInputsNode(testutil.NewNode(t, `
- name: myVar
  type: string
- name: myVar
  type: string
`))
	testutil.RequireContainsErr(t, errs, ErrInputNameCollision)
}

func TestVisitInputType_ErrIfUnknownType(t *testing.T) {
	_, errs := visitInputTypeNode(testutil.NewNode(t, `
name: myVar
type: unvariable
default: foo bar
`))
	testutil.RequireContainsErr(t, errs, ErrInputUnknownType)
}

func TestVisitInputType_AllValid(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		want    Input
	}{
		{
			name: "string",
			content: `
name: myVar
type: string
default: hello there`,
			want: InputString{
				Source:  visit.Pos{2, 1},
				Name:    "myVar",
				Default: "hello there",
			},
		},
		{
			name: "password",
			content: `
name: myVar
type: password
default: hello there`,
			want: InputPassword{
				Source:  visit.Pos{2, 1},
				Name:    "myVar",
				Default: "hello there",
			},
		},
		{
			name: "number int",
			content: `
name: myVar
type: number
default: 12345`,
			want: InputNumber{
				Source:  visit.Pos{2, 1},
				Name:    "myVar",
				Default: 12345,
			},
		},
		{
			name: "number float",
			content: `
name: myVar
type: number
default: 123.45`,
			want: InputNumber{
				Source:  visit.Pos{2, 1},
				Name:    "myVar",
				Default: 123.45,
			},
		},
		{
			name: "choice",
			content: `
name: myVar
type: choice
default: optionA
values: [optionA, optionB, optionC]`,
			want: InputChoice{
				Source:  visit.Pos{2, 1},
				Name:    "myVar",
				Default: "optionA",
				Values: []string{
					"optionA", "optionB", "optionC",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			node := testutil.NewNode(t, tc.content)
			got, errs := visitInputTypeNode(node)
			assert.Equal(t, tc.want, got)
			testutil.RequireNotContainsErr(t, errs, visit.ErrInvalidFieldType)
		})
	}
}

func TestVisitInputType_ErrIfMissingRequiredFields(t *testing.T) {
	testCases := []struct {
		name    string
		content string
	}{
		{
			name: "missing type",
			content: `
name: myVar
default: hello there`,
		},
		{
			name: "string missing name",
			content: `
type: string
default: hello there`,
		},
		{
			name: "password missing name",
			content: `
type: password
default: hello there`,
		},
		{
			name: "number missing name",
			content: `
type: number
default: 1234`,
		},
		{
			name: "choice missing name",
			content: `
type: choice
default: hello there
values: [hello there]`,
		},
		{
			name: "choice missing default",
			content: `
name: myVar
type: choice
values: [hello there]`,
		},
		{
			name: "choice missing values",
			content: `
name: myVar
type: choice
default: [hello there]`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, errs := visitInputTypeNode(testutil.NewNode(t, tc.content))
			testutil.RequireContainsErr(t, errs, visit.ErrMissingRequired)
		})
	}
}

func TestVisitInputChoice_ErrIfUndefinedDefaultValue(t *testing.T) {
	_, errs := visitInputTypeNode(testutil.NewNode(t, `
name: myVar
type: choice
default: optionF
values: [optionA, optionB, optionC]
`))
	testutil.RequireContainsErr(t, errs, ErrInputChoiceUnknownValue)
}
