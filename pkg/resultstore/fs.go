package resultstore

import (
	"io"
	"io/fs"
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
	// ListDirEntries will list all files, directories, symlinks, and other entries
	// inside a directory, non-recursively. It does not include the current "."
	// or parent ".." directory names.
	ListDirEntries(name string) ([]fs.DirEntry, error)
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
	return fs.openFileMkdirAll(name, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
}

func (fs osFS) OpenWrite(name string) (io.WriteCloser, error) {
	return fs.openFileMkdirAll(name, os.O_WRONLY|os.O_CREATE, 0644)
}

func (fs osFS) OpenRead(name string) (io.ReadCloser, error) {
	path := filepath.Join(fs.dir, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		writer, err := fs.openFileMkdirAll(name, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
		if err := writer.Close(); err != nil {
			log.Warn().WithError(err).Message("Failed closing file handle.")
		}
	}
	return os.OpenFile(path, os.O_RDONLY, 0644)
}

func (fs osFS) ListDirEntries(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(filepath.Join(fs.dir, name))
}

func (fs osFS) openFileMkdirAll(name string, flags int, perm fs.FileMode) (io.WriteCloser, error) {
	path := filepath.Join(fs.dir, name)
	log.Debug().WithString("path", path).Message("Opening file.")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Debug().WithString("path", path).Message("File does not exist. Creating.")
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, perm); err != nil {
			log.Error().WithError(err).WithString("dir", dir).Message("Failed to create parent directory recursively.")
		}
	}
	return os.OpenFile(path, flags, perm)
}
