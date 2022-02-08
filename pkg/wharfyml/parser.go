package wharfyml

import (
	"errors"
	"fmt"
	"io"

	"github.com/goccy/go-yaml"
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
	Stages []Stage2
}

type yamlDefinition struct {
	Environments map[string]Environment `yaml:"environments"`
	Stages       yaml.MapSlice          `yaml:",inline"`
}

func Parse2(reader io.Reader) (Definition, errorSlice) {
	def, errs := parse(reader)
	if len(errs) == 0 {
		return def, nil
	}
	for i, err := range errs {
		var parseErr ParseError
		if errors.As(err, &parseErr) {
			errs[i] = fmt.Errorf("%d:%d: %w",
				parseErr.Position.Line, parseErr.Position.Column, err)
		}
	}
	return Definition{}, errs
}

func parse(reader io.Reader) (def Definition, errSlice errorSlice) {
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
	def, errs = parseDocNodes(nodes)
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
		err := fmt.Errorf("documents: %d: %w", len(file.Docs), ErrTooManyDocs)
		// Continue, but only parse the first doc
		return doc, err
	}
	return doc, nil
}

func parseDocNodes(nodes []*ast.MappingValueNode) (def Definition, errSlice errorSlice) {
	for _, n := range nodes {
		key, err := parseMapKey(n.Key)
		if err != nil {
			errSlice.add(fmt.Errorf("%q: %w", n.Key, err))
			continue
		}
		errs := parseDocNodeIntoDefinition(&def, key, n)
		errSlice.add(errs...)
	}
	return
}

func parseDocNodeIntoDefinition(def *Definition, key *ast.StringNode, node *ast.MappingValueNode) errorSlice {
	switch key.Value {
	case propEnvironments:
		var errs errorSlice
		def.Envs, errs = parseDocEnvironmentsNode(node)
		return errs
	case propInput:
		// TODO: support inputs
		return errorSlice{errors.New("does not support input vars yet")}
	default:
		stage, errs := parseDocStageNode(key, node)
		def.Stages = append(def.Stages, stage)
		return errs
	}
}

func parseDocEnvironmentsNode(node *ast.MappingValueNode) (map[string]Env, errorSlice) {
	envs, errs := parseDefEnvironments(node.Value)
	errs.fmtErrorfAll("environments: %w", fmtErrorfPlaceholder)
	return envs, errs
}

func parseDocStageNode(key *ast.StringNode, node *ast.MappingValueNode) (Stage2, errorSlice) {
	stage, errs := parseStage2(key, node.Value)
	errs.fmtErrorfAll("stage %q: %w", key.Value, fmtErrorfPlaceholder)
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
