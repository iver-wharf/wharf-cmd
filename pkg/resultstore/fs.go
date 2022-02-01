package resultstore

import (
	"io"
	"os"
)

type FS interface {
	OpenAppend(name string) (io.WriteCloser, error)
	OpenWrite(name string) (io.WriteCloser, error)
	OpenReadWrite(name string) (io.ReadWriteCloser, error)
	OpenRead(name string) (io.ReadCloser, error)
}

type osFS struct{}

func (osFS) OpenAppend(name string) (io.WriteCloser, error) {
	return os.OpenFile(name, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
}

func (osFS) OpenWrite(name string) (io.WriteCloser, error) {
	return os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0644)
}

func (osFS) OpenReadWrite(name string) (io.ReadWriteCloser, error) {
	return os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0644)
}

func (osFS) OpenRead(name string) (io.ReadCloser, error) {
	return os.OpenFile(name, os.O_RDONLY, 0644)
}
