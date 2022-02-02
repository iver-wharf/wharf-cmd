package resultstore

import (
	"io"
)

type nopWriteCloser struct {
	writer io.Writer
}

func (w nopWriteCloser) Write(p []byte) (n int, err error) {
	if w.writer == nil {
		return len(p), nil
	}
	return w.writer.Write(p)
}

func (nopWriteCloser) Close() error {
	return nil
}

type mockFS struct {
	openAppend func(name string) (io.WriteCloser, error)
	openWrite  func(name string) (io.WriteCloser, error)
	openRead   func(name string) (io.ReadCloser, error)
}

func (fs mockFS) OpenAppend(name string) (io.WriteCloser, error) {
	return fs.openAppend(name)
}

func (fs mockFS) OpenWrite(name string) (io.WriteCloser, error) {
	return fs.openWrite(name)
}

func (fs mockFS) OpenRead(name string) (io.ReadCloser, error) {
	return fs.openRead(name)
}
