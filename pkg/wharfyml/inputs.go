package wharfyml

import (
	"errors"
	"strconv"

	"github.com/goccy/go-yaml/ast"
)

// Errors related to parsing environments.
var (
	ErrInputUnknownType        = errors.New("unknown input type")
	ErrInputChoiceUnknownValue = errors.New("default value is missing from values array")
)

func visitDocInputsNode(node ast.Node) (inputs []Input, errSlice Errors) {
	seqNode, err := parseSequenceNode(node)
	if err != nil {
		errSlice.add(err)
		return
	}
	for i, inputNode := range seqNode.Values {
		input, errs := visitInputTypeNode(inputNode)
		if len(errs) > 0 {
			errSlice.add(wrapPathErrorSlice(strconv.Itoa(i), errs)...)
		}
		if input != nil {
			inputs = append(inputs, input)
		}
	}
	return
}

func visitInputTypeNode(node ast.Node) (input Input, errSlice Errors) {
	nodes, err := parseMappingValueNodes(node)
	if err != nil {
		errSlice.add(err)
		return
	}
	nodeMap, errs := mappingValueNodeSliceToMap(nodes)
	errSlice.add(errs...)
	p := nodeMapParser{
		parent: node,
		nodes:  nodeMap,
	}
	var inputName string
	var inputType string
	errSlice.addNonNils(
		p.unmarshalString("name", &inputName),
		p.unmarshalString("type", &inputType),
		p.validateRequiredString("name"),
		p.validateRequiredString("type"),
	)
	switch inputType {
	case "":
		// validate required has already added error for it
		return
	case "string":
		inputString := InputString{Name: inputName}
		p.unmarshalString("default", &inputString.Default)
		input = inputString
	case "password":
		inputPassword := InputPassword{Name: inputName}
		p.unmarshalString("default", &inputPassword.Default)
		input = inputPassword
	case "number":
		inputNumber := InputNumber{Name: inputName}
		p.unmarshalNumber("default", &inputNumber.Default)
		input = inputNumber
	case "choice":
		inputChoice := InputChoice{Name: inputName}
		p.unmarshalString("default", &inputChoice.Default)
		p.unmarshalStringSlice("values", &inputChoice.Values)
		input = inputChoice
		errSlice.addNonNils(
			p.validateRequiredString("default"),
			p.validateRequiredSlice("values"),
			inputChoice.validate(),
		)
	default:
		errSlice.add(ErrInputUnknownType)
	}
	return
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

func (i InputChoice) validate() error {
	for _, v := range i.Values {
		if v == i.Default {
			return nil
		}
	}
	return ErrInputChoiceUnknownValue
}
