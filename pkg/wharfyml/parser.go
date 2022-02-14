package wharfyml

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
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
