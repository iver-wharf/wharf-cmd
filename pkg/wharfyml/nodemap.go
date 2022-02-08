package wharfyml

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml/ast"
)

var (
	ErrInvalidFieldType = errors.New("invalid field type")
)

type nodeMapUnmarshaller struct {
	parent ast.Node
	nodes  map[string]ast.Node
}

func (m nodeMapUnmarshaller) unmarshalString(key string, target *string) error {
	node, ok := m.nodes[key]
	if !ok {
		return nil
	}
	strNode, ok := node.(*ast.StringNode)
	if !ok {
		return newInvalidFieldTypeErr(key, "string", node)
	}
	*target = strNode.Value
	return nil
}

func (m nodeMapUnmarshaller) unmarshalStringSlice(key string, target *[]string) errorSlice {
	node, ok := m.nodes[key]
	if !ok {
		return nil
	}
	arrayNode, ok := node.(*ast.SequenceNode)
	if !ok {
		return errorSlice{newInvalidFieldTypeErr(key, "string array", node)}
	}
	strs := make([]string, 0, len(arrayNode.Values))
	var errSlice errorSlice
	for i, n := range arrayNode.Values {
		strNode, ok := n.(*ast.StringNode)
		if !ok {
			errSlice.add(newInvalidFieldTypeErr(fmt.Sprintf("%s[%d]", key, i),
				"string", n))
			continue
		}
		strs = append(strs, strNode.Value)
	}
	*target = strs
	return errSlice
}

func (m nodeMapUnmarshaller) unmarshalStringStringMap(key string, target *map[string]string) errorSlice {
	node, ok := m.nodes[key]
	if !ok {
		return nil
	}
	nodes, ok := getMappingValueNodes(node)
	if !ok {
		return errorSlice{newInvalidFieldTypeErr(key, "string to string map", node)}
	}
	strMap := make(map[string]string, len(nodes))
	var errSlice errorSlice
	for _, n := range nodes {
		keyNode, err := parseMapKey(n.Key)
		if err != nil {
			errSlice.add(err)
			continue
		}
		valNode, ok := n.Value.(*ast.StringNode)
		if !ok {
			errSlice.add(newInvalidFieldTypeErr(fmt.Sprintf("%s.%s", key, keyNode.Value),
				"string", n.Value))
			continue
		}
		strMap[keyNode.Value] = valNode.Value
	}
	*target = strMap
	return errSlice
}

func (m nodeMapUnmarshaller) unmarshalBool(key string, target *bool) error {
	node, ok := m.nodes[key]
	if !ok {
		return nil
	}
	strNode, ok := node.(*ast.BoolNode)
	if !ok {
		return newInvalidFieldTypeErr(key, "boolean", node)
	}
	*target = strNode.Value
	return nil
}

func (m nodeMapUnmarshaller) validateRequiredString(key string) error {
	node, ok := m.nodes[key]
	if ok {
		strNode, ok := node.(*ast.StringNode)
		if !ok || strNode.Value != "" {
			return nil
		}
	}
	return m.newRequiredError(key)
}

func (m nodeMapUnmarshaller) validateRequiredSlice(key string) error {
	node, ok := m.nodes[key]
	if ok {
		seqNode, ok := node.(*ast.SequenceNode)
		if !ok || len(seqNode.Values) > 0 {
			return nil
		}
	}
	return m.newRequiredError(key)
}

func (m nodeMapUnmarshaller) newRequiredError(key string) error {
	inner := fmt.Errorf("%w: %q", ErrStepTypeMissingRequired, key)
	return newParseErrorNode(inner, m.parent)
}

func newInvalidFieldTypeErr(key string, wantType string, node ast.Node) error {
	gotType := prettyNodeTypeName(node)
	err := newParseErrorNode(fmt.Errorf("%w: expected %s, but found %s",
		ErrInvalidFieldType, wantType, gotType), node)
	return wrapPathError(key, err)
}

func prettyNodeTypeName(node ast.Node) string {
	switch node.Type() {
	case ast.StringType:
		return "string"
	case ast.BoolType:
		return "boolean"
	case ast.FloatType:
		return "float"
	case ast.IntegerType:
		return "integer"
	case ast.NanType:
		return "NaN"
	case ast.InfinityType:
		return "infinity"
	case ast.MappingType:
		return "map"
	case ast.MappingKeyType:
		return "map key"
	case ast.MappingValueType:
		return "map value"
	case ast.NullType:
		return "null"
	case ast.SequenceType:
		return "array"
	default:
		return "unknown type"
	}
}
