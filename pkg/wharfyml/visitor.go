package wharfyml

import (
	"errors"
	"fmt"
	"strings"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"gopkg.in/yaml.v3"
)

// Generic errors related to parsing.
var (
	ErrInvalidFieldType = errors.New("invalid field type")
	ErrKeyCollision     = errors.New("map key appears more than once")
	ErrKeyEmpty         = errors.New("map key must not be empty")
	ErrKeyNotString     = errors.New("map key must be string")
	ErrMissingDoc       = errors.New("empty document")
	ErrTooManyDocs      = errors.New("only 1 document is allowed")
)

const (
	shortTagString    = "!!str"
	shortTagNull      = "!!null"
	shortTagInt       = "!!int"
	shortTagFloat     = "!!float"
	shortTagBool      = "!!bool"
	shortTagMap       = "!!map"
	shortTagSeq       = "!!seq"
	shortTagTimestamp = "!!timestamp"
	shortTagMerge     = "!!merge"
)

func verifyKind(node *yaml.Node, wantStr string, wantKind yaml.Kind) error {
	if node.Kind != wantKind {
		return wrapPosErrorNode(fmt.Errorf("%w: expected %s, but was %s",
			ErrInvalidFieldType, wantStr, prettyNodeTypeName(node)), node)
	}
	return nil
}

func verifyTag(node *yaml.Node, wantStr string, wantTag string) error {
	gotTag := node.ShortTag()
	if gotTag != wantTag {
		return wrapPosErrorNode(fmt.Errorf("%w: expected %s, but was %s",
			ErrInvalidFieldType, wantStr, prettyNodeTypeName(node)), node)
	}
	return nil
}

func verifyKindAndTag(node *yaml.Node, wantStr string, wantKind yaml.Kind, wantTag string) error {
	if err := verifyKind(node, wantStr, wantKind); err != nil {
		return err
	}
	return verifyTag(node, wantStr, wantTag)
}

func visitString(node *yaml.Node) (string, error) {
	if err := verifyKindAndTag(node, "string", yaml.ScalarNode, shortTagString); err != nil {
		return "", err
	}
	return node.Value, nil
}

func visitInt(node *yaml.Node) (int, error) {
	if err := verifyKindAndTag(node, "integer", yaml.ScalarNode, shortTagInt); err != nil {
		return 0, err
	}
	num, err := parseInt(node.Value)
	if err != nil {
		return 0, wrapPosErrorNode(err, node)
	}
	return num, nil
}

func visitFloat64(node *yaml.Node) (float64, error) {
	if node.Kind == yaml.ScalarNode && node.ShortTag() == shortTagInt {
		num, err := visitInt(node)
		if err != nil {
			return 0, err
		}
		return float64(num), nil
	}
	if err := verifyKindAndTag(node, "float", yaml.ScalarNode, shortTagFloat); err != nil {
		return 0, err
	}
	num, err := parseFloat64(node.Value)
	if err != nil {
		return 0, wrapPosErrorNode(err, node)
	}
	return num, nil
}

func visitBool(node *yaml.Node) (bool, error) {
	if err := verifyKindAndTag(node, "boolean", yaml.ScalarNode, shortTagBool); err != nil {
		return false, err
	}
	b, err := parseBool(node.Value)
	if err != nil {
		return false, wrapPosErrorNode(err, node)
	}
	return b, nil
}

func visitMap(node *yaml.Node) (map[string]*yaml.Node, errutil.Slice) {
	pairs, errs := visitMapSlice(node)
	m := make(map[string]*yaml.Node, len(pairs))
	for _, pair := range pairs {
		m[pair.key.value] = pair.value
	}
	return m, errs
}

type mapItem struct {
	key   strNode
	value *yaml.Node
}

type strNode struct {
	node  *yaml.Node
	value string
}

func visitMapSlice(node *yaml.Node) ([]mapItem, errutil.Slice) {
	var errSlice errutil.Slice

	if err := verifyKind(node, "map", yaml.MappingNode); err != nil {
		errSlice.Add(err)
		return nil, errSlice
	}

	pairs := make([]mapItem, 0, len(node.Content)/2)
	keys := make(map[string]struct{}, len(pairs))
	for i := 0; i < len(node.Content)-1; i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		if keyNode.Kind == yaml.ScalarNode && keyNode.ShortTag() == shortTagMerge {
			merged, errs := visitMapSlice(valueNode)
			errSlice.Add(errs...)
			pairs = append(pairs, merged...)
			continue
		}

		key, err := visitString(keyNode)
		if err != nil {
			errSlice.Add(wrapPosErrorNode(fmt.Errorf("%w: %v", ErrKeyNotString, err), keyNode))
			// non fatal error
		} else if key == "" {
			errSlice.Add(wrapPosErrorNode(ErrKeyEmpty, keyNode))
			// non fatal error
		}
		if _, ok := keys[key]; ok {
			errSlice.Add(errutil.Scope(
				wrapPosErrorNode(ErrKeyCollision, keyNode),
				key))
			continue
		}
		keys[key] = struct{}{}
		pairs = append(pairs, mapItem{strNode{keyNode, key}, valueNode})
	}
	return pairs, errSlice
}

func visitSequence(node *yaml.Node) ([]*yaml.Node, error) {
	if err := verifyKind(node, "sequence", yaml.SequenceNode); err != nil {
		return nil, err
	}
	return node.Content, nil
}

func visitDocument(node *yaml.Node) (*yaml.Node, error) {
	if err := verifyKind(node, "document", yaml.DocumentNode); err != nil {
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
	case shortTagString:
		return "string"
	case shortTagNull:
		return "null"
	case shortTagInt:
		return "integer"
	case shortTagFloat:
		return "float"
	case shortTagBool:
		return "boolean"
	case shortTagMap:
		return "map"
	case shortTagSeq:
		return "sequence"
	case shortTagTimestamp:
		return "timestamp"
	case "":
		return "undefined"
	default:
		return strings.TrimLeft(tag, "!")
	}
}

func prettyNodeTypeName(node *yaml.Node) string {
	switch node.Kind {
	case yaml.ScalarNode:
		return yamlShortTagName(node.ShortTag())
	default:
		return yamlKindString(node.Kind)
	}
}
