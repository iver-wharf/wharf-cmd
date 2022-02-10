package wharfyml

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

// Generic errors related to parsing.
var (
	ErrNotMap       = errors.New("not a map")
	ErrKeyNotString = errors.New("map key must be string")
	ErrKeyEmpty     = errors.New("map key must not be empty")
	ErrMissingDoc   = errors.New("empty document")
	ErrTooManyDocs  = errors.New("only 1 document is allowed")
)

// Definition is the .wharf-ci.yml build definition structure.
type Definition struct {
	Inputs []Input
	Envs   map[string]Env
	Stages []Stage
}

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
	doc, err := parseFirstDocAsDocNode(reader)
	if err != nil {
		errSlice.add(err)
	}
	if doc == nil {
		return
	}
	nodes, err := parseMappingValueNodes(doc.Body)
	if err != nil {
		errSlice.add(err)
		return
	}
	var errs Errors
	def, errs = visitDocNodes(nodes)
	if len(errs) > 0 {
		errSlice.add(errs...)
	}
	// TODO: second pass to validate environment usage:
	// - error on unused environment
	// - error on use of undeclared environment
	return
}

func parseFirstDocAsDocNode(reader io.Reader) (*ast.DocumentNode, error) {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	file, err := parser.ParseBytes(bytes, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	if len(file.Docs) == 0 {
		return nil, ErrMissingDoc
	}
	doc := file.Docs[0]
	if len(file.Docs) > 1 {
		err := fmt.Errorf("found %d documents: %w", len(file.Docs), ErrTooManyDocs)
		// Continue, but only parse the first doc
		return doc, err
	}
	return doc, nil
}

func visitDocNodes(nodes []*ast.MappingValueNode) (def Definition, errSlice Errors) {
	for _, n := range nodes {
		key, err := parseMapKeyNonEmpty(n.Key)
		if err != nil {
			errSlice.add(fmt.Errorf("%q: %w", n.Key, err))
			// non-fatal error
		}
		switch key {
		case propEnvironments:
			var errs Errors
			def.Envs, errs = visitDocEnvironmentsNodes(n.Value)
			errSlice.add(errs...)
		case propInputs:
			var errs Errors
			def.Inputs, errs = visitDocInputsNode(n.Value)
			errSlice.add(wrapPathErrorSlice(propInputs, errs)...)
		default:
			stage, errs := visitDocStageNode(key, n.Value)
			def.Stages = append(def.Stages, stage)
			errSlice.add(errs...)
		}
	}
	return
}

func visitDocEnvironmentsNodes(node ast.Node) (map[string]Env, Errors) {
	envs, errs := visitDocEnvironmentsNode(node)
	errs = wrapPathErrorSlice(propEnvironments, errs)
	return envs, errs
}

func visitDocStageNode(key string, node ast.Node) (Stage, Errors) {
	stage, errs := visitStageNode(key, node)
	errs = wrapPathErrorSlice(key, errs)
	return stage, errs
}

func parseMapKeyNonEmpty(node ast.Node) (string, error) {
	key, err := parseMapKey(node)
	if err != nil {
		return "", err
	}
	if key.Value == "" {
		return "", newPositionedErrorNode(ErrKeyEmpty, node)
	}
	return key.Value, nil
}

func parseMapKey(node ast.Node) (*ast.StringNode, error) {
	switch key := node.(type) {
	case *ast.StringNode:
		return key, nil
	default:
		return nil, newPositionedErrorNode(ErrKeyNotString, node)
	}
}

func parseMappingValueNodes(node ast.Node) ([]*ast.MappingValueNode, error) {
	switch n := node.(type) {
	case *ast.MappingValueNode:
		return []*ast.MappingValueNode{n}, nil
	case *ast.MappingNode:
		return n.Values, nil
	default:
		return nil, newPositionedErrorNode(fmt.Errorf(
			"%w: expected map, but was %s", ErrNotMap, prettyNodeTypeName(node)), node)
	}
}

func mappingValueNodeSliceToMap(slice []*ast.MappingValueNode) (map[string]ast.Node, Errors) {
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
