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
	ErrKeyNotString = errors.New("map key must be string")
	ErrMissingDoc   = errors.New("empty document")
	ErrTooManyDocs  = errors.New("only 1 document is allowed")
	ErrDocNotMap    = errors.New("document must be a map")
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
	nodes, err := docBodyAsNodes(doc.Body)
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
		key, err := parseMapKey(n.Key)
		if err != nil {
			errSlice.add(fmt.Errorf("%q: %w", n.Key, err))
			continue
		}
		switch key.Value {
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
	envs, errs := visitEnvironmentMapsNode(node)
	errs = wrapPathErrorSlice(propEnvironments, errs)
	return envs, errs
}

func visitDocStageNode(key *ast.StringNode, node ast.Node) (Stage, Errors) {
	stage, errs := visitStageNode(key, node)
	errs = wrapPathErrorSlice(key.Value, errs)
	return stage, errs
}

func parseMapKey(keyNode ast.Node) (*ast.StringNode, error) {
	switch key := keyNode.(type) {
	case *ast.StringNode:
		return key, nil
	default:
		return nil, newPositionedErrorNode(ErrKeyNotString, keyNode)
	}
}

func docBodyAsNodes(body ast.Node) ([]*ast.MappingValueNode, error) {
	n, ok := getMappingValueNodes(body)
	if !ok {
		return nil, newPositionedErrorNode(fmt.Errorf("document type: %s: %w", body.Type(), ErrDocNotMap), body)
	}
	return n, nil
}

// TODO: return err instead and only have 1 error type for non-map errors
func getMappingValueNodes(node ast.Node) ([]*ast.MappingValueNode, bool) {
	switch n := node.(type) {
	case *ast.MappingValueNode:
		return []*ast.MappingValueNode{n}, true
	case *ast.MappingNode:
		return n.Values, true
	default:
		return nil, false
	}
}

func mappingValueNodeSliceToMap(slice []*ast.MappingValueNode) (map[string]ast.Node, Errors) {
	m := make(map[string]ast.Node, len(slice))
	var errSlice Errors
	for _, node := range slice {
		key, err := parseMapKey(node.Key)
		if err != nil {
			errSlice.add(err)
			continue
		}
		m[key.Value] = node.Value
	}
	return m, errSlice
}
