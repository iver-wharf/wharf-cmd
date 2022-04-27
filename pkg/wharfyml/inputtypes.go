package wharfyml

import (
	"fmt"
	"strconv"
)

// InputString represents a string (text) input value.
type InputString struct {
	Source  Pos
	Name    string
	Default string
}

// InputTypeName returns the name of this input type.
func (InputString) InputTypeName() string {
	return "string"
}

// InputVarName returns the name of this input variable.
func (i InputString) InputVarName() string {
	return i.Name
}

// DefaultValue returns the default value for this input variable.
func (i InputString) DefaultValue() any {
	return i.Default
}

// Pos returns the position of where this variable was defined.
func (i InputString) Pos() Pos {
	return i.Source
}

// ParseValue will try to parse the value and return the input-compatible value
// for this input variable type, or returns an error if the type isn't
// valid.
func (i InputString) ParseValue(value any) (any, error) {
	return fmt.Sprint(value), nil
}

// InputPassword represents a string (text) input value, but where the value
// should be concealed in user interfaces.
type InputPassword struct {
	Source  Pos
	Name    string
	Default string
}

// InputTypeName returns the name of this input type.
func (InputPassword) InputTypeName() string {
	return "password"
}

// InputVarName returns the name of this input variable.
func (i InputPassword) InputVarName() string {
	return i.Name
}

// DefaultValue returns the default value for this input variable.
func (i InputPassword) DefaultValue() any {
	return i.Default
}

// Pos returns the position of where this variable was defined.
func (i InputPassword) Pos() Pos {
	return i.Source
}

// ParseValue will try to parse the value and return the input-compatible value
// for this input variable type, or returns an error if the type isn't
// valid.
func (i InputPassword) ParseValue(value any) (any, error) {
	return fmt.Sprint(value), nil
}

// InputNumber represents a number (integer or float) input value.
type InputNumber struct {
	Source  Pos
	Name    string
	Default float64
}

// InputTypeName returns the name of this input type.
func (InputNumber) InputTypeName() string {
	return "number"
}

// InputVarName returns the name of this input variable.
func (i InputNumber) InputVarName() string {
	return i.Name
}

// DefaultValue returns the default value for this input variable.
func (i InputNumber) DefaultValue() any {
	return i.Default
}

// Pos returns the position of where this variable was defined.
func (i InputNumber) Pos() Pos {
	return i.Source
}

// ParseValue will try to parse the value and return the input-compatible value
// for this input variable type, or returns an error if the type isn't
// valid.
func (i InputNumber) ParseValue(value any) (any, error) {
	switch value := value.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return value, nil
	case string:
		return strconv.ParseFloat(value, 64)
	case fmt.Stringer:
		return strconv.ParseFloat(value.String(), 64)
	default:
		return nil, fmt.Errorf("cannot parse type as number: %T", value)
	}
}

// InputChoice represents a choice of multiple string inputs.
type InputChoice struct {
	Source  Pos
	Name    string
	Values  []string
	Default string
}

// InputTypeName returns the name of this input type.
func (InputChoice) InputTypeName() string {
	return "choice"
}

// InputVarName returns the name of this input variable.
func (i InputChoice) InputVarName() string {
	return i.Name
}

// DefaultValue returns the default value for this input variable.
func (i InputChoice) DefaultValue() any {
	return i.Default
}

// ParseValue will try to parse the value and return the input-compatible value
// for this input variable type, or returns an error if the type isn't
// valid.
func (i InputChoice) ParseValue(value any) (any, error) {
	str := fmt.Sprint(value)
	if err := i.validateValue(str); err != nil {
		return nil, err
	}
	return str, nil
}

// Pos returns the position of where this variable was defined.
func (i InputChoice) Pos() Pos {
	return i.Source
}

func (i InputChoice) validateDefault() error {
	return i.validateValue(i.Default)
}

func (i InputChoice) validateValue(value string) error {
	for _, v := range i.Values {
		if v == value {
			return nil
		}
	}
	return ErrInputChoiceUnknownValue
}
