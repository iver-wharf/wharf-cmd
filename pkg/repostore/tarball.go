package repostore

import (
	"io"
	"os"
)

// Tarball is an identifier for a tarball file containing a repository.
type Tarball string

// Open creates a file handle to the tarball.
func (t Tarball) Open() (io.ReadCloser, error) {
	return os.Open(string(t))
}
