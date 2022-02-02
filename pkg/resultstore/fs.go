package resultstore

import (
	"io"
	"os"
	"path/filepath"
)

// FS is a filesystem with ability to open files in either append-only,
// write-only, or read only mode.
type FS interface {
	// OpenAppend creates or opens a file in append-only mode, meaning all written
	// data is appended to the end.
	OpenAppend(name string) (io.WriteCloser, error)
	// OpenWrite creates or opens a file in write-only mode, meaning all written
	// data is written from the start.
	OpenWrite(name string) (io.WriteCloser, error)
	// OpenRead opens a file in read-only-mode, reading data from the start of
	// the file.
	OpenRead(name string) (io.ReadCloser, error)
}

// NewFS creates a filesystem that will use the given directory as the base
// directory when creating or reading files.
func NewFS(dir string) FS {
	return osFS{dir: dir}
}

type osFS struct {
	dir string
}

func (fs osFS) OpenAppend(name string) (io.WriteCloser, error) {
	return os.OpenFile(filepath.Join(fs.dir, name), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
}

func (fs osFS) OpenWrite(name string) (io.WriteCloser, error) {
	return os.OpenFile(filepath.Join(fs.dir, name), os.O_WRONLY|os.O_CREATE, 0644)
}

func (fs osFS) OpenRead(name string) (io.ReadCloser, error) {
	return os.OpenFile(filepath.Join(fs.dir, name), os.O_RDONLY, 0644)
}
