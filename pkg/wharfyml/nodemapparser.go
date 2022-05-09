package wharfyml

import (
	"errors"
	"fmt"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"gopkg.in/yaml.v3"
)

// errutil.Slice related to parsing map of nodes.
var (
	ErrMissingRequired = errors.New("missing required field")
)

func newNodeMapParser(parent *yaml.Node, nodes map[string]*yaml.Node) nodeMapParser {
	return nodeMapParser{
		parent:    parent,
		nodes:     nodes,
		positions: make(map[string]Pos),
	}
}

type nodeMapParser struct {
	parent    *yaml.Node
	nodes     map[string]*yaml.Node
	positions map[string]Pos
}

func (p nodeMapParser) parentPos() Pos {
	return newPosNode(p.parent)
}

func (p nodeMapParser) hasNode(key string) bool {
	_, ok := p.nodes[key]
	return ok
}

func (p nodeMapParser) unmarshalNumber(key string, target *float64) error {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	p.positions[key] = newPosNode(node)
	num, err := visitFloat64(node)
	if err != nil {
		return errutil.Scope(err, key)
	}
	*target = num
	return nil
}

func (p nodeMapParser) unmarshalString(key string, target *string) error {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	p.positions[key] = newPosNode(node)
	str, err := visitString(node)
	if err != nil {
		return errutil.Scope(err, key)
	}
	*target = str
	return nil
}

func (p nodeMapParser) unmarshalStringSlice(key string, target *[]string) errutil.Slice {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	p.positions[key] = newPosNode(node)
	seq, err := visitSequence(node)
	if err != nil {
		return errutil.Slice{errutil.Scope(err, key)}
	}
	strs := make([]string, 0, len(seq))
	var errSlice errutil.Slice
	for i, n := range seq {
		str, err := visitString(n)
		if err != nil {
			errSlice.Add(errutil.Scope(err, fmt.Sprintf("%s[%d]", key, i)))
			continue
		}
		strs = append(strs, str)
	}
	*target = strs
	return errSlice
}

func (p nodeMapParser) unmarshalStringStringMap(key string, target *map[string]string) errutil.Slice {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	p.positions[key] = newPosNode(node)
	var errSlice errutil.Slice
	nodes, errs := visitMapSlice(node)
	errSlice.Add(errutil.ScopeSlice(errs, key)...)

	strMap := make(map[string]string, len(nodes))
	for _, n := range nodes {
		val, err := visitString(n.value)
		if err != nil {
			errSlice.Add(errutil.Scope(err, fmt.Sprintf("%s.%s", key, n.key.value)))
			continue
		}
		strMap[n.key.value] = val
	}
	*target = strMap
	return errSlice
}

func (p nodeMapParser) unmarshalStringWithVarSub(
	nodeKey, varLookup string, source varsub.Source, target *string) error {

	err := p.loadFromVarSubIfUnset(nodeKey, varLookup, source)
	if err != nil {
		return err
	}
	return p.unmarshalString(nodeKey, target)
}

func (p nodeMapParser) unmarshalStringFromVarSub(
	varLookup string, source varsub.Source, target *string) error {
	node, err := p.lookupFromVarSub(varLookup, source)
	if err != nil {
		return err
	}
	p.nodes["__tmp"] = node
	err = p.unmarshalString("__tmp", target)
	delete(p.nodes, "__tmp")
	return err
}

func (p nodeMapParser) unmarshalBool(key string, target *bool) error {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	p.positions[key] = newPosNode(node)
	b, err := visitBool(node)
	if err != nil {
		return errutil.Scope(err, key)
	}
	*target = b
	return nil
}

func (p nodeMapParser) validateRequiredString(key string) error {
	node, ok := p.nodes[key]
	if !ok {
		return p.newRequiredError(key)
	}
	isStr := node.Kind == yaml.ScalarNode && node.ShortTag() == shortTagString
	if isStr && node.Value == "" {
		return p.newRequiredError(key)
	}
	return nil
}

func (p nodeMapParser) validateRequiredSlice(key string) error {
	node, ok := p.nodes[key]
	if !ok {
		return p.newRequiredError(key)
	}
	isSeq := node.Kind == yaml.SequenceNode
	if isSeq && len(node.Content) == 0 {
		return p.newRequiredError(key)
	}
	return nil
}

func (p nodeMapParser) newRequiredError(key string) error {
	inner := fmt.Errorf("%w: %q", ErrMissingRequired, key)
	return wrapPosErrorNode(inner, p.parent)
}

func (p nodeMapParser) loadFromVarSubIfUnset(
	nodeKey, varLookup string, source varsub.Source) error {
	if _, ok := p.nodes[nodeKey]; ok {
		return nil
	}
	node, err := p.lookupFromVarSub(varLookup, source)
	if err != nil {
		return err
	}
	p.nodes[nodeKey] = node
	return nil
}

func (p nodeMapParser) lookupFromVarSub(
	varLookup string, source varsub.Source) (*yaml.Node, error) {
	varValue, ok := safeLookupVar(source, varLookup)
	if !ok {
		err := fmt.Errorf("%w: need %s", ErrMissingBuiltinVar, varLookup)
		return nil, wrapPosErrorNode(err, p.parent)
	}
	newNode, err := newNodeWithValue(p.parent, varValue)
	if err != nil {
		err := fmt.Errorf("read %s: %w", varLookup, err)
		return nil, wrapPosErrorNode(err, p.parent)
	}
	return newNode, nil
}

func safeLookupVar(source varsub.Source, name string) (interface{}, bool) {
	if source == nil {
		return nil, false
	}
	return source.Lookup(name)
}
