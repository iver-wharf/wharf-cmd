package wharfyml

import (
	"errors"

	"github.com/goccy/go-yaml/ast"
)

// Errors related to parsing environments.
var (
	ErrInputsNotArray            = errors.New("inputs should be a YAML array")
	ErrInputTypeMissingRequired  = errors.New("missing required field")
	ErrInputTypeInvalidFieldType = errors.New("invalid field type")
	ErrInputChoiceUnknownValue   = errors.New("default value is missing from values array")
)

func visitDocInputsNode(node ast.Node) ([]Input, Errors) {
	// TODO:
	return nil, nil
}

func visitInputTypeNode(node ast.Node) (Input, Errors) {
	return nil, nil
}

// Input is an interface that is implemented by all input types.
type Input interface {
	InputTypeName() string
}

// InputString represents a string (text) input value.
type InputString struct {
	Name    string
	Default string
}

// InputTypeName returns the name of this input type.
func (InputString) InputTypeName() string {
	return "string"
}

// InputPassword represents a string (text) input value, but where the value
// should be concealed in user interfaces.
type InputPassword struct {
	Name    string
	Default string
}

// InputTypeName returns the name of this input type.
func (InputPassword) InputTypeName() string {
	return "password"
}

// InputNumber represents a number (integer or float) input value.
type InputNumber struct {
	Name    string
	Default float64
}

// InputTypeName returns the name of this input type.
func (InputNumber) InputTypeName() string {
	return "number"
}

// InputChoice represents a choice of multiple string inputs.
type InputChoice struct {
	Name    string
	Values  []string
	Default string
}

// InputTypeName returns the name of this input type.
func (InputChoice) InputTypeName() string {
	return "choice"
}
