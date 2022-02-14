package wharfyml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDocInputs_ErrIfNotArray(t *testing.T) {
	_, errs := visitInputsNode(getNode(t, `123`))
	requireContainsErr(t, errs, ErrInvalidFieldType)
}

func TestParseDocInputs_ErrIfNameCollision(t *testing.T) {
	_, errs := visitInputsNode(getNode(t, `
- name: myVar
  type: string
- name: myVar
  type: string
`))
	requireContainsErr(t, errs, ErrInputNameCollision)
}

func TestParseInputType_ErrIfUnknownType(t *testing.T) {
	_, errs := visitInputTypeNode(getNode(t, `
name: myVar
type: unvariable
default: foo bar
`))
	requireContainsErr(t, errs, ErrInputUnknownType)
}

func TestParseInputType_AllValid(t *testing.T) {
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
				Source:  Pos{2, 1},
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
				Source:  Pos{2, 1},
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
				Source:  Pos{2, 1},
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
				Source:  Pos{2, 1},
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
				Source:  Pos{2, 1},
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
			node := getNode(t, tc.content)
			got, errs := visitInputTypeNode(node)
			assert.Equal(t, tc.want, got)
			requireNotContainsErr(t, errs, ErrInvalidFieldType)
		})
	}
}

func TestParseInputType_ErrIfMissingRequiredFields(t *testing.T) {
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
			_, errs := visitInputTypeNode(getNode(t, tc.content))
			requireContainsErr(t, errs, ErrMissingRequired)
		})
	}
}

func TestParseInputChoice_ErrIfUndefinedDefaultValue(t *testing.T) {
	_, errs := visitInputTypeNode(getNode(t, `
name: myVar
type: choice
default: optionF
values: [optionA, optionB, optionC]
`))
	requireContainsErr(t, errs, ErrInputChoiceUnknownValue)
}
