package resultstore

import (
	"io"
	"io/fs"
)

type nopCloser struct{}

func (nopCloser) Close() error { return nil }

type nopWriteCloser struct {
	writer io.Writer
}

func (w nopWriteCloser) Write(p []byte) (n int, err error) {
	if w.writer == nil {
		return len(p), nil
	}
	return w.writer.Write(p)
}

func (nopWriteCloser) Close() error { return nil }

type mockFS struct {
	openAppend     func(name string) (io.WriteCloser, error)
	openWrite      func(name string) (io.WriteCloser, error)
	openRead       func(name string) (io.ReadCloser, error)
	listDirEntries func(name string) ([]fs.DirEntry, error)
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

func (fs mockFS) ListDirEntries(name string) ([]fs.DirEntry, error) {
	return fs.listDirEntries(name)
}

type mockDirEntry struct {
	name  func() string
	isDir func() bool
	typ   func() fs.FileMode
	info  func() (fs.FileInfo, error)
}

func (m mockDirEntry) Name() string {
	return m.name()
}

func (m mockDirEntry) IsDir() bool {
	return m.isDir()
}

func (m mockDirEntry) Type() fs.FileMode {
	return m.typ()
}

func (m mockDirEntry) Info() (fs.FileInfo, error) {
	return m.info()
}

func newMockDirEntryFile(name string) mockDirEntry {
	return mockDirEntry{
		name:  func() string { return name },
		isDir: func() bool { return false },
	}
}

func newMockDirEntryDir(name string) mockDirEntry {
	return mockDirEntry{
		name:  func() string { return name },
		isDir: func() bool { return true },
	}
}
