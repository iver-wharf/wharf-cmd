package wharfyml

import (
	"io"
	"os"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
)

// Args specify arguments used when parsing the .wharf-ci.yml file, such as what
// environment to use for variable substitution.
type Args struct {
	Env                string
	VarSource          varsub.Source
	SkipStageFiltering bool
	Inputs             map[string]any
}

// ParseFile will parse the file at the given path.
// Multiple errors may be returned, one for each validation or parsing error.
func ParseFile(path string, args Args) (Definition, errutil.Slice) {
	file, err := os.Open(path)
	if err != nil {
		return Definition{}, errutil.Slice{err}
	}
	defer file.Close()
	return Parse(file, args)
}

// Parse will parse the YAML content as a .wharf-ci.yml definition structure.
// Multiple errors may be returned, one for each validation or parsing error.
func Parse(reader io.Reader, args Args) (def Definition, errSlice errutil.Slice) {
	def, errs := parse(reader, args)
	errutil.SortByPos(errs)
	return def, errs
}

func parse(reader io.Reader, args Args) (def Definition, errSlice errutil.Slice) {
	doc, err := visit.DecodeFirstRootNode(reader)
	if err != nil {
		errSlice.Add(err)
	}
	if doc == nil {
		return
	}
	var errs errutil.Slice
	def, errs = visitDefNode(doc, args)
	errSlice.Add(errs...)
	return
}
