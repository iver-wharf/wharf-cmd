package visit

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// DecodeFirstRootNode returns the first YAML root node from the first parsed
// document. If there are multiple documents then the first document is returned
// together with an error.
func DecodeFirstRootNode(reader io.Reader) (*yaml.Node, error) {
	rootNodes, err := DecodeRootNodes(reader)
	if err != nil {
		return nil, err
	}
	if len(rootNodes) == 0 {
		return nil, ErrMissingDoc
	}
	if len(rootNodes) > 1 {
		return nil, fmt.Errorf("%w: expected 1, found %d", ErrTooManyDocs, len(rootNodes))
	}
	return rootNodes[0], nil
}

// DecodeRootNodes returns the list of YAML root nodes from all documents
// in the parsed input.
func DecodeRootNodes(reader io.Reader) ([]*yaml.Node, error) {
	dec := yaml.NewDecoder(reader)
	var rootNodes []*yaml.Node
	for {
		var doc yaml.Node
		if err := dec.Decode(&doc); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("document %d: %w", len(rootNodes)+1, err)
		}
		root, err := Document(&doc)
		if err != nil {
			return nil, fmt.Errorf("document %d: %w", len(rootNodes)+1, err)
		}
		root = unwrapNodeRec(root)
		rootNodes = append(rootNodes, root)
	}
	return rootNodes, nil
}

func unwrapNodeRec(node *yaml.Node) *yaml.Node {
	for node.Alias != nil {
		node = node.Alias
	}
	for i, child := range node.Content {
		node.Content[i] = unwrapNodeRec(child)
	}
	return node
}
