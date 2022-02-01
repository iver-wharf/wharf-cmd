package resultstore

import (
	"io"
	"os"
	"path/filepath"
)

type FS interface {
	OpenAppend(name string) (io.WriteCloser, error)
	OpenWrite(name string) (io.WriteCloser, error)
	OpenRead(name string) (io.ReadCloser, error)
}

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
