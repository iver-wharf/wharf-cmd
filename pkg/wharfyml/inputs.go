package wharfyml

import "fmt"

type InputString struct {
	Name    string
	Type    InputType
	Default string
}

type InputPassword struct {
	Name    string
	Type    InputType
	Default string
}

type InputNumber struct {
	Name    string
	Type    InputType
	Default int
}

type InputChoice struct {
	Name    string
	Type    InputType
	Values  []interface{}
	Default interface{}
}

func parseInput(inputMap map[string]interface{}) (interface{}, error) {
	inputTypeName, ok := inputMap[inputType].(string)
	if !ok {
		return nil, fmt.Errorf("invalid input type: %v", inputMap)
	}

	inputType, ok := ParseInputType(inputTypeName)
	if !ok {
		return nil, fmt.Errorf("invalid input type: %v", inputMap)
	}

	inputName, ok := inputMap[inputName].(string)
	if inputName == "" || !ok {
		return nil, fmt.Errorf("invalid input name: %v", inputMap)
	}

	switch inputType {
	case String:
		def, ok := inputMap[inputDefault].(string)
		if !ok {
			def = ""
		}

		return InputString{
			Name:    inputName,
			Type:    inputType,
			Default: def,
		}, nil
	case Choice:
		def := inputMap[inputDefault]
		if def == "" {
			return nil, fmt.Errorf("invalid input, missing default: %v", inputMap)
		}

		values, ok := inputMap[inputValues].([]interface{})
		if len(values) == 0 || !ok {
			return nil, fmt.Errorf("invalid input, missing default: %v", inputMap)
		}

		return InputChoice{
			Name:    inputName,
			Type:    inputType,
			Values:  values,
			Default: def,
		}, nil
	case Number:
		defNumber, ok := inputMap[inputDefault].(float64)
		if !ok {
			defNumber = 0
		}

		return InputNumber{
			Name:    inputName,
			Type:    inputType,
			Default: int(defNumber),
		}, nil
	case Password:
		def, ok := inputMap[inputDefault].(string)
		if !ok {
			def = ""
		}

		return InputPassword{
			Name:    inputName,
			Type:    inputType,
			Default: def,
		}, nil
	}

	return nil, fmt.Errorf("invalid input type, %v", inputMap)
}
