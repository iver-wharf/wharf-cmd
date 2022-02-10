package wharfyml

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDocInputs_ErrIfNotArray(t *testing.T) {
	_, errs := visitDocInputsNode(getNode(t, `123`))
	requireContainsErr(t, errs, ErrInputsNotArray)
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
			node := getNode(t, fmt.Sprintf(`
name: myVar
type: %s
default: %s
`, tc.name, tc.content))
			inputType, errs := visitInputTypeNode(node)
			assert.IsType(t, tc.want, inputType)
			requireNotContainsErr(t, errs, ErrInputTypeInvalidFieldType)
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
			requireNotContainsErr(t, errs, ErrInputTypeMissingRequired)
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
