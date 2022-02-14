package wharfyml

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/goccy/go-yaml/ast"
	"gopkg.in/yaml.v3"
)

// Generic errors related to parsing.
var (
	ErrNotMap       = errors.New("not a map")
	ErrNotArray     = errors.New("not an array")
	ErrKeyNotString = errors.New("map key must be string")
	ErrKeyEmpty     = errors.New("map key must not be empty")
	ErrKeyDuplicate = errors.New("map key appears more than once")
	ErrMissingDoc   = errors.New("empty document")
	ErrTooManyDocs  = errors.New("only 1 document is allowed")
)

// ParseFile will parse the file at the given path.
// Multiple errors may be returned, one for each validation or parsing error.
func ParseFile(path string) (Definition, Errors) {
	file, err := os.Open(path)
	if err != nil {
		return Definition{}, Errors{err}
	}
	defer file.Close()
	return Parse(file)
}

// Parse will parse the YAML content as a .wharf-ci.yml definition structure.
// Multiple errors may be returned, one for each validation or parsing error.
func Parse(reader io.Reader) (def Definition, errSlice Errors) {
	def, errs := parse(reader)
	sortErrorsByPosition(errs)
	return def, errs
}

func parse(reader io.Reader) (def Definition, errSlice Errors) {
	doc, err := decodeFirstDoc(reader)
	if err != nil {
		errSlice.add(err)
	}
	if doc == nil {
		return
	}
	var errs Errors
	def, errs = visitDefNode(doc)
	errSlice.add(errs...)
	return
}

func decodeFirstDoc(reader io.Reader) (*yaml.Node, error) {
	dec := yaml.NewDecoder(reader)
	var doc yaml.Node
	err := dec.Decode(&doc)
	if err == io.EOF {
		return nil, ErrMissingDoc
	}
	if err != nil {
		return nil, err
	}
	body, err := visitDocument(&doc)
	if err != nil {
		return nil, err
	}
	var unusedNode yaml.Node
	if err := dec.Decode(&unusedNode); err != io.EOF {
		// Continue, but only parse the first doc
		return body, ErrTooManyDocs
	}
	return body, nil
}

func parseMapKeyNonEmpty(node ast.Node) (string, error) {
	key, err := parseMapKey(node)
	if err != nil {
		return "", err
	}
	if key.Value == "" {
		return "", wrapPosErrorNode(ErrKeyEmpty, node)
	}
	return key.Value, nil
}

func parseMapKey(node ast.Node) (*ast.StringNode, error) {
	switch key := node.(type) {
	case *ast.StringNode:
		return key, nil
	default:
		return nil, wrapPosErrorNode(ErrKeyNotString, node)
	}
}

func parseMappingValueNodes(node ast.Node) ([]*ast.MappingValueNode, error) {
	switch n := node.(type) {
	case *ast.AnchorNode:
		return parseMappingValueNodes(n.Value)
	case *ast.AliasNode:
		return parseMappingValueNodes(n.Value)
	case *ast.MappingValueNode:
		return []*ast.MappingValueNode{n}, nil
	case *ast.MappingNode:
		return n.Values, nil
	default:
		return nil, wrapPosErrorNode(fmt.Errorf(
			"%w: expected map, but was %s", ErrNotMap, prettyNodeTypeName(node)), node)
	}
}

func parseSequenceNode(node ast.Node) (*ast.SequenceNode, error) {
	switch n := node.(type) {
	case *ast.SequenceNode:
		return n, nil
	default:
		return nil, wrapPosErrorNode(fmt.Errorf(
			"%w: expected array, but was %s", ErrNotArray, prettyNodeTypeName(node)), node)
	}
}

func parseMappingValueNodeSliceAsMap(slice []*ast.MappingValueNode) (map[string]ast.Node, Errors) {
	m := make(map[string]ast.Node, len(slice))
	var errSlice Errors
	for _, node := range slice {
		key, err := parseMapKeyNonEmpty(node.Key)
		if err != nil {
			errSlice.add(err)
			continue
		}
		m[key] = node.Value
	}
	return m, errSlice
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
	case ast.AnchorType:
		return "anchor"
	case ast.AliasType:
		return "alias"
	case ast.TagType:
		return "tag"
	default:
		return "unknown type"
	}
}

func prettyNodeTypeName2(node *yaml.Node) string {
	switch node.Kind {
	case yaml.ScalarNode:
		return yamlShortTagName(node.ShortTag())
	default:
		return yamlKindString(node.Kind)
	}
}
