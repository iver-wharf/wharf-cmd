package wharfyml

import (
	"errors"
	"fmt"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"gopkg.in/yaml.v3"
)

// Errors specific to performing variable substitution on nodes.
var (
	ErrUnsupportedVarSubType = errors.New("unsupported variable substitution value")
)

// VarSubNode is a custom varsub variable type that envelops a YAML node.
// Mostly only used internally inside the wharfyml package.
type VarSubNode struct {
	Node *yaml.Node
}

// String implements the fmt.Stringer interface.
func (v VarSubNode) String() string {
	return v.Node.Value
}

func varSubNodeRec(node *yaml.Node, source varsub.Source) (*yaml.Node, error) {
	if source == nil {
		return node, nil
	}
	if node.Kind == yaml.ScalarNode {
		if node.Tag != shortTagString {
			return node, nil
		}
		return varSubStringNode(strNode{node, node.Value}, source)
	}
	if len(node.Content) == 0 {
		return node, nil
	}
	clone := *node
	clone.Content = make([]*yaml.Node, len(node.Content))
	for i, child := range node.Content {
		child, err := varSubNodeRec(child, source)
		if err != nil {
			return nil, err
		}
		clone.Content[i] = child
	}
	return &clone, nil
}

func varSubStringNode(str strNode, source varsub.Source) (*yaml.Node, error) {
	val, err := varsub.Substitute(str.value, source)
	if err != nil {
		return nil, wrapPosErrorNode(err, str.node)
	}
	return newNodeWithValue(str.node, val)
}

func newNodeWithValue(node *yaml.Node, val any) (*yaml.Node, error) {
	clone := *node
	clone.Kind = yaml.ScalarNode
	switch val := val.(type) {
	case nil:
		clone.Tag = shortTagNull
		clone.Value = ""
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		clone.Tag = shortTagInt
		clone.Value = fmt.Sprint(val)
	case float32, float64:
		clone.Tag = shortTagFloat
		clone.Value = fmt.Sprint(val)
	case time.Time:
		clone.Tag = shortTagTimestamp
		clone.Value = val.Format(time.RFC3339Nano)
	case bool:
		clone.Tag = shortTagBool
		if val {
			clone.Value = "true"
		} else {
			clone.Value = "false"
		}
	case *yaml.Node:
		return val, nil
	case VarSubNode:
		return val.Node, nil
	case string:
		clone.SetString(val)
	case fmt.Stringer:
		clone.SetString(val.String())
	default:
		err := fmt.Errorf("%w: %T", ErrUnsupportedVarSubType, val)
		return nil, wrapPosErrorNode(err, node)
	}
	return &clone, nil
}
