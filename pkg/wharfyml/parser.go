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
	Stages []Stage
}

type yamlDefinition struct {
	Environments map[string]Environment `yaml:"environments"`
	Stages       yaml.MapSlice          `yaml:",inline"`
}

func Parse2(reader io.Reader) (Definition, []error) {
	def, errs := parse(reader)
	if len(errs) == 0 {
		return def, nil
	}
	for i, err := range errs {
		var parseErr ParseError
		if errors.As(err, &parseErr) {
			errs[i] = fmt.Errorf("%d:%d: %w", parseErr.Position.Line, parseErr.Position.Column, err)
		}
	}
	return Definition{}, errs
}

func parse(reader io.Reader) (def Definition, errSlice []error) {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		errSlice = append(errSlice, err)
		return
	}
	file, err := parser.ParseBytes(bytes, parser.ParseComments)
	if err != nil {
		errSlice = append(errSlice, err)
		return
	}
	if len(file.Docs) == 0 {
		errSlice = append(errSlice, ErrMissingDoc)
		return
	}
	if len(file.Docs) > 1 {
		errSlice = append(errSlice, fmt.Errorf("documents: %d: %w", len(file.Docs), ErrTooManyDocs))
		// Continue, but only parse the first doc
	}
	nodes, err := docBodyAsNodes(file.Docs[0].Body)
	if err != nil {
		errSlice = append(errSlice, err)
		return
	}
	for _, n := range nodes {
		key, err := parseDocMapKey(n.Key)
		if err != nil {
			errSlice = append(errSlice, fmt.Errorf("%q: %w", n.Key, err))
			continue
		}
		switch key.Value {
		case propEnvironments:
			envs, errs := parseEnvironments(n.Value)
			if len(errs) > 0 {
				for i, err := range errs {
					errs[i] = fmt.Errorf("environments: %w", err)
				}
				errSlice = append(errSlice, errs...)
				continue
			}
			def.Envs = envs
		case propInput:
			// TODO: support inputs
			errSlice = append(errSlice, errors.New("does not support input vars yet"))
		default:
			stage, errs := parseStage2(key, n.Value)
			if len(errs) > 0 {
				for i, err := range errs {
					errs[i] = fmt.Errorf("stage %q: %w", key, err)
				}
				errSlice = append(errSlice, errs...)
				continue
			}
			def.Stages = append(def.Stages, stage)
		}
	}
	// TODO: second pass to validate environment usage:
	// - error on unused environment
	// - error on use of undeclared environment
	return
}

func parseDocMapKey(keyNode ast.Node) (*ast.StringNode, error) {
	switch key := keyNode.(type) {
	case *ast.StringNode:
		return key, nil
	default:
		return nil, wrapParseErrNode(ErrKeyNotString, keyNode)
	}
}

func docBodyAsNodes(body ast.Node) ([]*ast.MappingValueNode, error) {
	switch b := body.(type) {
	case *ast.MappingValueNode:
		return []*ast.MappingValueNode{b}, nil
	case *ast.MappingNode:
		return b.Values, nil
	default:
		return nil, wrapParseErrNode(fmt.Errorf("document type: %s: %w", body.Type(), ErrDocNotMap), body)
	}
}
