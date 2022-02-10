package wharfyml

import (
	"errors"
	"fmt"
	"math"

	"github.com/goccy/go-yaml/ast"
)

// Errors related to parsing map of nodes.
var (
	ErrInvalidFieldType = errors.New("invalid field type")
	ErrMissingRequired  = errors.New("missing required field")
)

type nodeMapParser struct {
	parent ast.Node
	nodes  map[string]ast.Node
}

func (p nodeMapParser) parentPos() Pos {
	return newPosNode(p.parent)
}

func (p nodeMapParser) unmarshalNumber(key string, target *float64) error {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	switch n := node.(type) {
	case *ast.NanNode:
		*target = math.NaN()
	case *ast.FloatNode:
		*target = n.Value
	case *ast.InfinityNode:
		*target = n.Value
	case *ast.IntegerNode:
		// By documentation of ast.IntegerNode, it will only be
		// either uint64 or int64
		switch num := n.Value.(type) {
		case uint64:
			*target = float64(num)
		case int64:
			*target = float64(num)
		}
	default:
		return newInvalidFieldTypeErr(key, "number", node)
	}
	return nil
}

func (p nodeMapParser) unmarshalString(key string, target *string) error {
	node, ok := p.nodes[key]
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

func (p nodeMapParser) unmarshalStringSlice(key string, target *[]string) Errors {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	arrayNode, ok := node.(*ast.SequenceNode)
	if !ok {
		return Errors{newInvalidFieldTypeErr(key, "string array", node)}
	}
	strs := make([]string, 0, len(arrayNode.Values))
	var errSlice Errors
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

func (p nodeMapParser) unmarshalStringStringMap(key string, target *map[string]string) Errors {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	nodes, err := parseMappingValueNodes(node)
	if err != nil {
		return Errors{err}
	}
	strMap := make(map[string]string, len(nodes))
	var errSlice Errors
	for _, n := range nodes {
		mapKey, err := parseMapKeyNonEmpty(n.Key)
		if err != nil {
			errSlice.add(err)
			continue
		}
		valNode, ok := n.Value.(*ast.StringNode)
		if !ok {
			errSlice.add(newInvalidFieldTypeErr(fmt.Sprintf("%s.%s", key, mapKey),
				"string", n.Value))
			continue
		}
		strMap[mapKey] = valNode.Value
	}
	*target = strMap
	return errSlice
}

func (p nodeMapParser) unmarshalBool(key string, target *bool) error {
	node, ok := p.nodes[key]
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

func (p nodeMapParser) validateRequiredString(key string) error {
	node, ok := p.nodes[key]
	if ok {
		strNode, ok := node.(*ast.StringNode)
		if !ok || strNode.Value != "" {
			return nil
		}
	}
	return p.newRequiredError(key)
}

func (p nodeMapParser) validateRequiredSlice(key string) error {
	node, ok := p.nodes[key]
	if ok {
		seqNode, ok := node.(*ast.SequenceNode)
		if !ok || len(seqNode.Values) > 0 {
			return nil
		}
	}
	return p.newRequiredError(key)
}

func (p nodeMapParser) newRequiredError(key string) error {
	inner := fmt.Errorf("%w: %q", ErrMissingRequired, key)
	return wrapPosErrorNode(inner, p.parent)
}

func newInvalidFieldTypeErr(key string, wantType string, node ast.Node) error {
	gotType := prettyNodeTypeName(node)
	err := wrapPosErrorNode(fmt.Errorf("%w: expected %s, but found %s",
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
