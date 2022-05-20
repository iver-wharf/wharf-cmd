package visit

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"gopkg.in/yaml.v3"
)

// Generic errors related to visiting YAML nodes.
var (
	ErrInvalidFieldType = errors.New("invalid field type")
	ErrKeyCollision     = errors.New("map key appears more than once")
	ErrKeyEmpty         = errors.New("map key must not be empty")
	ErrKeyNotString     = errors.New("map key must be string")
	ErrMissingDoc       = errors.New("empty document")
	ErrTooManyDocs      = errors.New("only 1 document is allowed")
)

// String will try to read this YAML node as a string.
func String(node *yaml.Node) (string, error) {
	if err := VerifyKindAndTag(node, "string", yaml.ScalarNode, ShortTagString); err != nil {
		return "", err
	}
	return node.Value, nil
}

// Int will try to read this YAML node as an integer.
func Int(node *yaml.Node) (int, error) {
	if err := VerifyKindAndTag(node, "integer", yaml.ScalarNode, ShortTagInt); err != nil {
		return 0, err
	}
	num, err := parseInt(node.Value)
	if err != nil {
		return 0, errutil.NewPosFromNode(err, node)
	}
	return num, nil
}

// Uint will try to read this YAML node as an integer.
func Uint(node *yaml.Node) (uint, error) {
	if err := VerifyKindAndTag(node, "unsigned integer", yaml.ScalarNode, ShortTagInt); err != nil {
		return 0, err
	}
	num, err := parseUint(node.Value)
	if err != nil {
		return 0, errutil.NewPosFromNode(err, node)
	}
	return num, nil
}

// Float64 will try to read this YAML node as a float64.
func Float64(node *yaml.Node) (float64, error) {
	if node.Kind == yaml.ScalarNode && node.ShortTag() == ShortTagInt {
		num, err := Int(node)
		if err != nil {
			return 0, err
		}
		return float64(num), nil
	}
	if err := VerifyKindAndTag(node, "float", yaml.ScalarNode, ShortTagFloat); err != nil {
		return 0, err
	}
	num, err := parseFloat64(node.Value)
	if err != nil {
		return 0, errutil.NewPosFromNode(err, node)
	}
	return num, nil
}

// Bool will try to read this YAML node as a bool.
func Bool(node *yaml.Node) (bool, error) {
	if err := VerifyKindAndTag(node, "boolean", yaml.ScalarNode, ShortTagBool); err != nil {
		return false, err
	}
	b, err := parseBool(node.Value)
	if err != nil {
		return false, errutil.NewPosFromNode(err, node)
	}
	return b, nil
}

// Map will try to read this YAML node as a map of nodes with string keys.
func Map(node *yaml.Node) (map[string]*yaml.Node, errutil.Slice) {
	pairs, errs := MapSlice(node)
	m := make(map[string]*yaml.Node, len(pairs))
	for _, pair := range pairs {
		m[pair.Key.Value] = pair.Value
	}
	return m, errs
}

// MapItem is a key-value pair item from a YAML map node.
type MapItem struct {
	Key   StringNode
	Value *yaml.Node
}

// StringNode is a node with known string type, with both the string value and
// original node available.
type StringNode struct {
	Node  *yaml.Node
	Value string
}

// MapSlice will try to read this YAML node as a map of nodes with string keys,
// returned as a slice of key-value pairs, keeping the original order of the
// nodes from the YAML file.
func MapSlice(node *yaml.Node) ([]MapItem, errutil.Slice) {
	var errSlice errutil.Slice

	if err := VerifyKind(node, "map", yaml.MappingNode); err != nil {
		errSlice.Add(err)
		return nil, errSlice
	}

	pairs := make([]MapItem, 0, len(node.Content)/2)
	keys := make(map[string]struct{}, len(pairs))
	for i := 0; i < len(node.Content)-1; i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		if keyNode.Kind == yaml.ScalarNode && keyNode.ShortTag() == ShortTagMerge {
			merged, errs := MapSlice(valueNode)
			errSlice.Add(errs...)
			pairs = append(pairs, merged...)
			continue
		}

		key, err := String(keyNode)
		if err != nil {
			errSlice.Add(errutil.NewPosFromNode(fmt.Errorf("%w: %v", ErrKeyNotString, err), keyNode))
			// non fatal error
		} else if key == "" {
			errSlice.Add(errutil.NewPosFromNode(ErrKeyEmpty, keyNode))
			// non fatal error
		}
		if _, ok := keys[key]; ok {
			errSlice.Add(errutil.Scope(
				errutil.NewPosFromNode(ErrKeyCollision, keyNode),
				key))
			continue
		}
		keys[key] = struct{}{}
		pairs = append(pairs, MapItem{StringNode{keyNode, key}, valueNode})
	}
	return pairs, errSlice
}

// Sequence will try to read this YAML node as a sequence of nodes (array).
func Sequence(node *yaml.Node) ([]*yaml.Node, error) {
	if err := VerifyKind(node, "sequence", yaml.SequenceNode); err != nil {
		return nil, err
	}
	return node.Content, nil
}

// Document will try to read the root node of this document YAML node.
func Document(node *yaml.Node) (*yaml.Node, error) {
	if err := VerifyKind(node, "document", yaml.DocumentNode); err != nil {
		return nil, err
	}
	return node.Content[0], nil
}

func yamlKindString(kind yaml.Kind) string {
	switch kind {
	case yaml.DocumentNode:
		return "document"
	case yaml.SequenceNode:
		return "sequence"
	case yaml.MappingNode:
		return "mapping"
	case yaml.ScalarNode:
		return "scalar"
	case yaml.AliasNode:
		return "alias"
	default:
		return fmt.Sprintf("unknown (%d)", kind)
	}
}

func yamlShortTagName(tag string) string {
	switch tag {
	case ShortTagString:
		return "string"
	case ShortTagNull:
		return "null"
	case ShortTagInt:
		return "integer"
	case ShortTagFloat:
		return "float"
	case ShortTagBool:
		return "boolean"
	case ShortTagMap:
		return "map"
	case ShortTagSeq:
		return "sequence"
	case ShortTagTimestamp:
		return "timestamp"
	case "":
		return "undefined"
	default:
		return strings.TrimLeft(tag, "!")
	}
}

// PrettyNodeTypeName will return a human-readable type name of a node.
func PrettyNodeTypeName(node *yaml.Node) string {
	switch node.Kind {
	case yaml.ScalarNode:
		return yamlShortTagName(node.ShortTag())
	default:
		return yamlKindString(node.Kind)
	}
}

func removeUnderscores(str string) string {
	// YAML supports underscore delimiters for readability, while
	// strconv.ParseFloat does not.
	return strings.ReplaceAll(str, "_", "")
}

func parseInt(str string) (int, error) {
	num, err := strconv.ParseInt(removeUnderscores(str), 0, 0)
	if err != nil {
		return 0, err
	}
	return int(num), nil
}

func parseUint(str string) (uint, error) {
	num, err := strconv.ParseUint(removeUnderscores(str), 0, 0)
	if err != nil {
		return 0, err
	}
	return uint(num), nil
}

func parseFloat64(str string) (float64, error) {
	// https://yaml.org/type/float.html
	switch str {
	case ".inf", ".Inf", ".INF", "+.inf", "+.Inf", "+.INF":
		return math.Inf(1), nil
	case "-.inf", "-.Inf", "-.INF":
		return math.Inf(-1), nil
	case ".nan", ".NaN", ".NAN":
		return math.NaN(), nil
	}
	num, err := strconv.ParseFloat(removeUnderscores(str), 64)
	if err != nil {
		return 0, err
	}
	return num, nil
}

func parseBool(val string) (bool, error) {
	// Got damn, YAML has too many boolean alternatives...
	// https://yaml.org/type/bool.html
	switch val {
	case "y", "Y", "yes", "Yes", "YES",
		"true", "True", "TRUE",
		"on", "On", "ON":
		return true, nil
	case "n", "N", "no", "No", "NO",
		"off", "Off", "OFF",
		"false", "False", "FALSE":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %q", val)
	}
}
