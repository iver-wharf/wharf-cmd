package wharfyml

import (
	"errors"
	"fmt"

	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"gopkg.in/yaml.v3"
)

// Errors related to parsing map of nodes.
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
		return wrapPathError(err, key)
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
		return wrapPathError(err, key)
	}
	*target = str
	return nil
}

func (p nodeMapParser) unmarshalStringSlice(key string, target *[]string) Errors {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	p.positions[key] = newPosNode(node)
	seq, err := visitSequence(node)
	if err != nil {
		return Errors{wrapPathError(err, key)}
	}
	strs := make([]string, 0, len(seq))
	var errSlice Errors
	for i, n := range seq {
		str, err := visitString(n)
		if err != nil {
			errSlice.add(wrapPathError(err, fmt.Sprintf("%s[%d]", key, i)))
			continue
		}
		strs = append(strs, str)
	}
	*target = strs
	return errSlice
}

func (p nodeMapParser) unmarshalStringStringMap(key string, target *map[string]string) Errors {
	node, ok := p.nodes[key]
	if !ok {
		return nil
	}
	p.positions[key] = newPosNode(node)
	var errSlice Errors
	nodes, errs := visitMapSlice(node)
	errSlice.add(wrapPathErrorSlice(errs, key)...)

	strMap := make(map[string]string, len(nodes))
	for _, n := range nodes {
		val, err := visitString(n.value)
		if err != nil {
			errSlice.add(wrapPathError(err, fmt.Sprintf("%s.%s", key, n.key.value)))
			continue
		}
		strMap[n.key.value] = val
	}
	*target = strMap
	return errSlice
}

func (p nodeMapParser) unmarshalStringFromNodeOrVarSubForOther(
	nodeKey, varLookup, other string, source varsub.Source, target *string) error {

	err := p.readFromVarSubForOther(nodeKey, varLookup, other, source)
	if err != nil {
		return err
	}
	return p.unmarshalString(nodeKey, target)
}

func (p nodeMapParser) unmarshalStringFromVarSubForOther(
	varLookup, other string, source varsub.Source, target *string) error {
	node, err := p.lookupFromVarSubForOther(
		varLookup, other, source)
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
		return wrapPathError(err, key)
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

func (p nodeMapParser) readFromVarSubForOther(
	nodeKey, varLookup, other string, source varsub.Source) error {
	if _, ok := p.nodes[nodeKey]; ok {
		return nil
	}
	value, ok := source.Lookup(varLookup)
	if !ok {
		return wrapPosErrorNode(fmt.Errorf(
			"%w: need %s or %q to construct %q",
			ErrMissingBuiltinVar, varLookup, nodeKey, other),
			p.parent)
	}
	newNode, err := newNodeWithValue(p.parent, value)
	if err != nil {
		return wrapPosErrorNode(fmt.Errorf(
			"read %s to construct %q: %w", varLookup, other, err),
			p.parent)
	}
	p.nodes["registry"] = newNode
	return nil
}

func (p nodeMapParser) lookupFromVarSubForOther(
	varLookup, other string, source varsub.Source) (*yaml.Node, error) {
	repoNameVar, ok := source.Lookup(varLookup)
	if !ok {
		err := fmt.Errorf("%w: need %s to construct %q",
			ErrMissingBuiltinVar, varLookup, other)
		return nil, wrapPosErrorNode(err, p.parent)
	}
	newNode, err := newNodeWithValue(p.parent, repoNameVar)
	if err != nil {
		err := fmt.Errorf("read %s to construct %q: %w",
			varLookup, other, err)
		return nil, wrapPosErrorNode(err, p.parent)
	}
	return newNode, nil
}
