package wharfyml

import (
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"gopkg.in/yaml.v3"
)

// Args specify arguments used when parsing the .wharf-ci.yml file, such as what
// environment to use for variable substitution.
type Args struct {
	Env       string
	VarSource varsub.Source
}

// ParseFile will parse the file at the given path.
// Multiple errors may be returned, one for each validation or parsing error.
func ParseFile(path string, args Args) (Definition, Errors) {
	file, err := os.Open(path)
	if err != nil {
		return Definition{}, Errors{err}
	}
	defer file.Close()
	return Parse(file, args)
}

// Parse will parse the YAML content as a .wharf-ci.yml definition structure.
// Multiple errors may be returned, one for each validation or parsing error.
func Parse(reader io.Reader, args Args) (def Definition, errSlice Errors) {
	def, errs := parse(reader, args)
	sortErrorsByPosition(errs)
	return def, errs
}

func parse(reader io.Reader, args Args) (def Definition, errSlice Errors) {
	doc, err := decodeFirstDoc(reader)
	if err != nil {
		errSlice.add(err)
	}
	if doc == nil {
		return
	}
	var errs Errors
	def, errs = visitDefNode(doc, args)
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
	body = unwrapNodeRec(body)
	return body, nil
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

func parseInt(str string) (int, error) {
	num, err := strconv.ParseInt(removeUnderscores(str), 0, 0)
	if err != nil {
		return 0, err
	}
	return int(num), nil
}

func parseFloat64(str string) (float64, error) {
	// https://yaml.org/type/float.html
	switch str {
	case ".inf", ".Inf", ".INF", "+.inf", "+.Inf", "+.INF":
		return math.Inf(1), nil
	case "-.inf", "-.Inf", "-.INF":
		return math.Inf(-1), nil
	case ".nan", ".NaN", ".NAN":
		return math.NaN(), nil
	}
	num, err := strconv.ParseFloat(removeUnderscores(str), 64)
	if err != nil {
		return 0, err
	}
	return num, nil
}

func removeUnderscores(str string) string {
	// YAML supports underscore delimiters for readability, while
	// strconv.ParseFloat does not.
	return strings.ReplaceAll(str, "_", "")
}

func parseBool(val string) (bool, error) {
	// Got damn, YAML has too many boolean alternatives...
	// https://yaml.org/type/bool.html
	switch val {
	case "y", "Y", "yes", "Yes", "YES",
		"true", "True", "TRUE",
		"on", "On", "ON":
		return true, nil
	case "n", "N", "no", "No", "NO",
		"off", "Off", "OFF",
		"false", "False", "FALSE":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %q", val)
	}
}
