package wharfyml

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	ErrInvalidFieldType = errors.New("invalid field type")
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

type visitor struct {
	errs Errors
}

func unwrapNode(node *yaml.Node) *yaml.Node {
	for node.Alias != nil {
		node = node.Alias
	}
	return node
}

func verifyKind(node *yaml.Node, wantStr string, wantKind yaml.Kind) error {
	if node.Kind != wantKind {
		return wrapPosErrorNode2(fmt.Errorf("%w: expected %s, but was %s",
			ErrInvalidFieldType, wantStr, prettyNodeTypeName2(node)), node)
	}
	return nil
}

func verifyTag(node *yaml.Node, wantStr string, wantTag string) error {
	gotTag := node.ShortTag()
	if gotTag != wantTag {
		return wrapPosErrorNode2(fmt.Errorf("%w: expected %s, but was %s",
			ErrInvalidFieldType, wantStr, prettyNodeTypeName2(node)), node)
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
	node = unwrapNode(node)
	if err := verifyKindAndTag(node, "string", yaml.ScalarNode, shortTagString); err != nil {
		return "", err
	}
	return node.Value, nil
}

func visitInt(node *yaml.Node) (int, error) {
	node = unwrapNode(node)
	if err := verifyKindAndTag(node, "integer", yaml.ScalarNode, shortTagInt); err != nil {
		return 0, err
	}
	num, err := parseInt(node.Value)
	if err != nil {
		return 0, wrapPosErrorNode2(err, node)
	}
	return num, nil
}

func parseInt(str string) (int, error) {
	num, err := strconv.ParseInt(removeUnderscores(str), 0, 0)
	if err != nil {
		return 0, err
	}
	return int(num), nil
}

func visitUInt(node *yaml.Node) (uint, error) {
	node = unwrapNode(node)
	if err := verifyKindAndTag(node, "integer", yaml.ScalarNode, shortTagInt); err != nil {
		return 0, err
	}
	num, err := parseUint(node.Value)
	if err != nil {
		return 0, wrapPosErrorNode2(err, node)
	}
	return num, nil
}

func parseUint(str string) (uint, error) {
	num, err := strconv.ParseUint(removeUnderscores(str), 0, 0)
	if err != nil {
		return 0, err
	}
	return uint(num), nil
}

func visitFloat64(node *yaml.Node) (float64, error) {
	node = unwrapNode(node)
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
		return 0, wrapPosErrorNode2(err, node)
	}
	return num, nil
}

func parseFloat64(str string) (float64, error) {
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

func removeUnderscores(str string) string {
	// YAML supports underscore delimiters for readability, while
	// strconv.ParseFloat does not.
	return strings.ReplaceAll(str, "_", "")
}

func visitBool(node *yaml.Node) (bool, error) {
	node = unwrapNode(node)
	if err := verifyKindAndTag(node, "boolean", yaml.ScalarNode, shortTagBool); err != nil {
		return false, err
	}
	b, err := parseBool(node.Value)
	if err != nil {
		return false, wrapPosErrorNode2(err, node)
	}
	return b, nil
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

func visitMap(node *yaml.Node) (map[string]*yaml.Node, Errors) {
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

func visitMapSlice(node *yaml.Node) ([]mapItem, Errors) {
	node = unwrapNode(node)
	var errSlice Errors

	if err := verifyKind(node, "map", yaml.MappingNode); err != nil {
		errSlice.add(err)
		return nil, errSlice
	}

	pairs := make([]mapItem, 0, len(node.Content)/2)
	keys := make(map[string]struct{}, len(pairs))
	for i := 0; i < len(node.Content)-1; i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		if keyNode.Kind == yaml.ScalarNode && keyNode.ShortTag() == shortTagMerge {
			merged, errs := visitMapSlice(valueNode)
			errSlice.add(errs...)
			pairs = append(pairs, merged...)
			continue
		}

		key, err := visitString(keyNode)
		if err != nil {
			errSlice.add(wrapPosErrorNode2(fmt.Errorf("%w: %v", ErrKeyNotString, err), keyNode))
			// non fatal error
		} else if key == "" {
			errSlice.add(wrapPosErrorNode2(ErrKeyEmpty, keyNode))
			// non fatal error
		}
		if _, ok := keys[key]; ok {
			errSlice.add(wrapPathError(key,
				wrapPosErrorNode2(ErrKeyCollision, keyNode)))
			continue
		}
		keys[key] = struct{}{}
		fmt.Printf("node %q value kind: %[2]T(%[2]d)\n", key, valueNode.Kind)
		pairs = append(pairs, mapItem{strNode{keyNode, key}, valueNode})
	}
	return pairs, errSlice
}

func visitSequence(node *yaml.Node) ([]*yaml.Node, error) {
	node = unwrapNode(node)
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
