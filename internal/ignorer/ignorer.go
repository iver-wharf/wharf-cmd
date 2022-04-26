package ignorer

import (
	"path/filepath"
	"strings"

	"gopkg.in/typ.v3/pkg/slices"
)

// Ignorer is an interface for conditionally ignoring files or directory trees
// when creating a tarball.
type Ignorer interface {
	// Ignore returns true to ignore a file, and false to include the file.
	Ignore(relPath string) bool
}

// Merge returns an Ignorer implementation that returns true if any of the
// provided ignorers return true.
func Merge(ignorers ...Ignorer) Ignorer {
	return merge(ignorers)
}

type merge []Ignorer

func (m merge) Ignore(relPath string) bool {
	for _, i := range m {
		if i.Ignore(relPath) {
			return true
		}
	}
	return false
}

// NewFileIncluder creates an Ignorer that will ignore all files except the
// given files. Parent directories of the given files are also included.
func NewFileIncluder(relPaths []string) Ignorer {
	return fileIncluder(slices.Map(relPaths, filepath.Clean))
}

type fileIncluder []string

var separatorStr = string(filepath.Separator)

func (fi fileIncluder) Ignore(relPath string) bool {
	relPath = filepath.Clean(relPath)
	relPathDir := relPath + separatorStr
	for _, want := range fi {
		if want == relPath {
			return false
		}
		if strings.HasPrefix(want, relPathDir) {
			return false
		}
	}
	return true
}
