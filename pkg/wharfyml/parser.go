package wharfyml

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

var (
	ErrKeyNotString  = errors.New("map key must be string")
	ErrMissingDoc    = errors.New("empty document")
	ErrTooManyDocs   = errors.New("only 1 document is allowed")
	ErrDocNotMap     = errors.New("document must be a map")
	ErrMissingStages = errors.New("missing stages")
)

type Definition struct {
	Envs   map[string]Env
	Stages []Stage
}

func ParseFile(path string) (Definition, errorSlice) {
	file, err := os.Open(path)
	if err != nil {
		return Definition{}, errorSlice{err}
	}
	defer file.Close()
	return Parse(file)
}

func Parse(reader io.Reader) (def Definition, errSlice errorSlice) {
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
	var errs errorSlice
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

func visitDocNodes(nodes []*ast.MappingValueNode) (def Definition, errSlice errorSlice) {
	for _, n := range nodes {
		key, err := parseMapKey(n.Key)
		if err != nil {
			errSlice.add(fmt.Errorf("%q: %w", n.Key, err))
			continue
		}
		switch key.Value {
		case propEnvironments:
			var errs errorSlice
			def.Envs, errs = visitDocEnvironmentsNodes(n)
			errSlice.add(errs...)
		case propInput:
			// TODO: support inputs
			errSlice.add(errors.New("does not support input vars yet"))
		default:
			stage, errs := visitDocStageNode(key, n)
			def.Stages = append(def.Stages, stage)
			errSlice.add(errs...)
		}
	}
	return
}

func visitDocEnvironmentsNodes(node *ast.MappingValueNode) (map[string]Env, errorSlice) {
	envs, errs := visitEnvironmentMapsNode(node.Value)
	errs = wrapPathErrorSlice(propEnvironments, errs)
	return envs, errs
}

func visitDocStageNode(key *ast.StringNode, node *ast.MappingValueNode) (Stage, errorSlice) {
	stage, errs := visitStageNode(key, node.Value)
	errs = wrapPathErrorSlice(key.Value, errs)
	return stage, errs
}

func parseMapKey(keyNode ast.Node) (*ast.StringNode, error) {
	switch key := keyNode.(type) {
	case *ast.StringNode:
		return key, nil
	default:
		return nil, newParseErrorNode(ErrKeyNotString, keyNode)
	}
}

func docBodyAsNodes(body ast.Node) ([]*ast.MappingValueNode, error) {
	n, ok := getMappingValueNodes(body)
	if !ok {
		return nil, newParseErrorNode(fmt.Errorf("document type: %s: %w", body.Type(), ErrDocNotMap), body)
	}
	return n, nil
}

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

func mappingValueNodeSliceToMap(slice []*ast.MappingValueNode) (map[string]ast.Node, errorSlice) {
	m := make(map[string]ast.Node, len(slice))
	var errSlice errorSlice
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
