package filecopy

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/iver-wharf/wharf-cmd/internal/ignorer"
)

// Copier is an interface that can copy bytes from one reader to a writer.
type Copier interface {
	Copy(dst io.Writer, src io.Reader) error
}

type ioCopier struct{}

func (ioCopier) Copy(dst io.Writer, src io.Reader) error {
	_, err := io.Copy(dst, src)
	return err
}

// IOCopier is a copier implementation that uses the built-in io.Copy()
// to copy the content in a loop.
var IOCopier Copier = ioCopier{}

// CopyFile takes a destination and source path and copies the content via the
// copier implementation.
func CopyFile(dst, src string, copier Copier) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src: %w", err)
	}
	defer srcFile.Close()

	srcFileStat, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("stat src: %w", err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, srcFileStat.Mode())
	if err != nil {
		return fmt.Errorf("open dst: %w", err)
	}
	defer dstFile.Close()

	return copier.Copy(dstFile, srcFile)
}

// CopyDirIgnorer will recursively copy all files from the source path over to
// the destination path. Both dst and src are expected to be existing
// directories, and then all content of one directory is copied into the other.
//
// The passed ignorer is used to conditionally omit files or directory trees
// from the copy operation.
//
// Only files and directories are copied. Any sockets, devices, symlinks, or
// other file types are silently ignored.
func CopyDirIgnorer(dst, src string, copier Copier, ignorer ignorer.Ignorer) error {
	// TODO: Could possibly multithread this for better performance.
	// Would need some benchmarks to actually prove it's faster.

	srcFS := os.DirFS(src)
	return fs.WalkDir(srcFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		srcPath := filepath.Join(src, path)
		info, err := os.Stat(srcPath)
		if err != nil {
			return err
		}
		if ignorer != nil {
			if ignorer.Ignore(path) {
				if info.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
		}
		dstPath := filepath.Join(dst, path)
		if info.IsDir() {
			err := os.Mkdir(dstPath, info.Mode())
			if os.IsExist(err) {
				return nil
			}
			return err
		}
		isFile := info.Mode().Type() == 0
		if isFile {
			return CopyFile(dstPath, srcPath, copier)
		}
		// Just skip other file types, e.g symlinks, devices, and sockets.
		return nil
	})
}

// CopyDir will recursively copy all files from the source path over to
// the destination path. Both dst and src are expected to be existing
// directories, and then all content of one directory is copied into the other.
//
// Only files and directories are copied. Any sockets, devices, symlinks, or
// other file types are silently ignored.
func CopyDir(dst, src string, copier Copier) error {
	return CopyDirIgnorer(dst, src, copier, nil)
}
