package wharfyml

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

func (i InputChoice) validate() error {
	for _, v := range i.Values {
		if v == i.Default {
			return nil
		}
	}
	return ErrInputChoiceUnknownValue
}
