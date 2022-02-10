package wharfyml

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/goccy/go-yaml/ast"
)

// Errors related to parsing environments.
var (
	ErrInputNameCollision      = errors.New("input variable name is already used")
	ErrInputUnknownType        = errors.New("unknown input type")
	ErrInputChoiceUnknownValue = errors.New("default value is missing from values array")
)

func visitInputsNode(node ast.Node) (inputs map[string]Input, errSlice Errors) {
	seqNode, err := parseSequenceNode(node)
	if err != nil {
		errSlice.add(err)
		return
	}
	inputs = make(map[string]Input, len(seqNode.Values))
	for i, inputNode := range seqNode.Values {
		input, errs := visitInputTypeNode(inputNode)
		if len(errs) > 0 {
			errSlice.add(wrapPathErrorSlice(strconv.Itoa(i), errs)...)
		}
		if input != nil {
			name := input.InputVarName()
			if _, ok := inputs[name]; ok {
				err := wrapPosErrorNode(fmt.Errorf("%w: %q", ErrInputNameCollision, name), inputNode)
				errSlice.add(wrapPathError(strconv.Itoa(i), err))
			}
			inputs[name] = input
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
	p := newNodeMapParser(node, nodeMap)
	var inputName string
	var inputType string
	errSlice.addNonNils(
		p.unmarshalString("name", &inputName),
		p.unmarshalString("type", &inputType),
		p.validateRequiredString("name"),
		p.validateRequiredString("type"),
	)
	pos := newPosNode(node)
	switch inputType {
	case "":
		// validate required has already added error for it
		return
	case "string":
		inputString := InputString{Name: inputName, Pos: pos}
		p.unmarshalString("default", &inputString.Default)
		input = inputString
	case "password":
		inputPassword := InputPassword{Name: inputName, Pos: pos}
		p.unmarshalString("default", &inputPassword.Default)
		input = inputPassword
	case "number":
		inputNumber := InputNumber{Name: inputName, Pos: pos}
		p.unmarshalNumber("default", &inputNumber.Default)
		input = inputNumber
	case "choice":
		inputChoice := InputChoice{Name: inputName, Pos: pos}
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
	InputVarName() string
}

// InputString represents a string (text) input value.
type InputString struct {
	Name    string
	Default string
	Pos     Pos
}

// InputTypeName returns the name of this input type.
func (InputString) InputTypeName() string {
	return "string"
}

// InputVarName returns the name of this input variable.
func (i InputString) InputVarName() string {
	return i.Name
}

// InputPassword represents a string (text) input value, but where the value
// should be concealed in user interfaces.
type InputPassword struct {
	Name    string
	Default string
	Pos     Pos
}

// InputTypeName returns the name of this input type.
func (InputPassword) InputTypeName() string {
	return "password"
}

// InputVarName returns the name of this input variable.
func (i InputPassword) InputVarName() string {
	return i.Name
}

// InputNumber represents a number (integer or float) input value.
type InputNumber struct {
	Name    string
	Default float64
	Pos     Pos
}

// InputTypeName returns the name of this input type.
func (InputNumber) InputTypeName() string {
	return "number"
}

// InputVarName returns the name of this input variable.
func (i InputNumber) InputVarName() string {
	return i.Name
}

// InputChoice represents a choice of multiple string inputs.
type InputChoice struct {
	Name    string
	Values  []string
	Default string
	Pos     Pos
}

// InputTypeName returns the name of this input type.
func (InputChoice) InputTypeName() string {
	return "choice"
}

// InputVarName returns the name of this input variable.
func (i InputChoice) InputVarName() string {
	return i.Name
}

func (i InputChoice) validate() error {
	for _, v := range i.Values {
		if v == i.Default {
			return nil
		}
	}
	return ErrInputChoiceUnknownValue
}
