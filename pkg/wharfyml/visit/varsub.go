package visit

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

func VarSubNodeRec(node *yaml.Node, source varsub.Source) (*yaml.Node, error) {
	if source == nil {
		return node, nil
	}
	if node.Kind == yaml.ScalarNode {
		if node.Tag != ShortTagString {
			return node, nil
		}
		return VarSubStringNode(StringNode{node, node.Value}, source)
	}
	if len(node.Content) == 0 {
		return node, nil
	}
	clone := *node
	clone.Content = make([]*yaml.Node, len(node.Content))
	for i, child := range node.Content {
		child, err := VarSubNodeRec(child, source)
		if err != nil {
			return nil, err
		}
		clone.Content[i] = child
	}
	return &clone, nil
}

func VarSubStringNode(str StringNode, source varsub.Source) (*yaml.Node, error) {
	val, err := varsub.Substitute(str.Value, source)
	if err != nil {
		return nil, WrapPosErrorNode(err, str.Node)
	}
	return NewNodeWithValue(str.Node, val)
}

func NewNodeWithValue(node *yaml.Node, val any) (*yaml.Node, error) {
	clone := *node
	clone.Kind = yaml.ScalarNode
	switch val := val.(type) {
	case nil:
		clone.Tag = ShortTagNull
		clone.Value = ""
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		clone.Tag = ShortTagInt
		clone.Value = fmt.Sprint(val)
	case float32, float64:
		clone.Tag = ShortTagFloat
		clone.Value = fmt.Sprint(val)
	case time.Time:
		clone.Tag = ShortTagTimestamp
		clone.Value = val.Format(time.RFC3339Nano)
	case bool:
		clone.Tag = ShortTagBool
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
		return nil, WrapPosErrorNode(err, node)
	}
	return &clone, nil
}

func safeLookupVar(source varsub.Source, name string) (varsub.Var, bool) {
	if source == nil {
		return varsub.Var{}, false
	}
	return source.Lookup(name)
}
