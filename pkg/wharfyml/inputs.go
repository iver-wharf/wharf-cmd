package wharfyml

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"gopkg.in/yaml.v3"
)

// Errors related to parsing environments.
var (
	ErrInputNameCollision      = errors.New("input variable name is already used")
	ErrInputUnknownType        = errors.New("unknown input type")
	ErrInputChoiceUnknownValue = errors.New("default value is missing from values array")
)

// Inputs is a map of Input field definitions, keyed on their names.
type Inputs map[string]Input

// DefaultsVarSource returns a varsub.Source of the default values from this
// .wharf-ci.yml's input definitions.
func (i Inputs) DefaultsVarSource() varsub.Source {
	source := make(varsub.SourceMap, len(i))
	for k, input := range i {
		source[k] = varsub.Val{
			Value:  input.DefaultValue(),
			Source: ".wharf-ci.yml, input defaults",
		}
	}
	return source
}

// Input is an interface that is implemented by all input types.
type Input interface {
	InputTypeName() string
	InputVarName() string
	DefaultValue() any
}

func visitInputsNode(node *yaml.Node) (inputs Inputs, errSlice Errors) {
	nodes, err := visitSequence(node)
	if err != nil {
		errSlice.add(err)
		return
	}
	inputs = make(Inputs, len(nodes))
	for i, inputNode := range nodes {
		input, errs := visitInputTypeNode(inputNode)
		if len(errs) > 0 {
			errSlice.add(wrapPathErrorSlice(errs, strconv.Itoa(i))...)
		}
		if input != nil {
			name := input.InputVarName()
			if _, ok := inputs[name]; ok {
				err := wrapPosErrorNode(
					fmt.Errorf("%w: %q", ErrInputNameCollision, name), inputNode)
				errSlice.add(wrapPathError(err, strconv.Itoa(i)))
			}
			inputs[name] = input
		}
	}
	return
}

func visitInputTypeNode(node *yaml.Node) (input Input, errSlice Errors) {
	nodeMap, errs := visitMap(node)
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
		inputString := InputString{Name: inputName, Source: pos}
		p.unmarshalString("default", &inputString.Default)
		input = inputString
	case "password":
		inputPassword := InputPassword{Name: inputName, Source: pos}
		p.unmarshalString("default", &inputPassword.Default)
		input = inputPassword
	case "number":
		inputNumber := InputNumber{Name: inputName, Source: pos}
		p.unmarshalNumber("default", &inputNumber.Default)
		input = inputNumber
	case "choice":
		inputChoice := InputChoice{Name: inputName, Source: pos}
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
