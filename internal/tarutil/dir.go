package tarutil

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/iver-wharf/wharf-cmd/internal/ignorer"
	"github.com/iver-wharf/wharf-core/pkg/logger"
)

var log = logger.NewScoped("TAR")

// Options contains options for when creating tarballs.
type Options struct {
	Path       string // empty string means current working directory
	Ignorer    ignorer.Ignorer
	FileOpener FileOpener
}

// FileOpener is an interface for opening files for reading, used in the Dir
// function when reading files.
type FileOpener interface {
	OpenFile(path string) (io.ReadCloser, error)
}

type osFileOpener struct{}

func (osFileOpener) OpenFile(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

var fileSeparatorString = string(filepath.Separator)

// Dir will recursively tar the contents of an entire directory. Hidden files
// (files that start with a dot) are included. The name of the target directory
// is not included in the tarball, but instead only the children.
func Dir(w io.Writer, opts Options) error {
	rootDirPath, err := filepath.Abs(opts.Path)
	if err != nil {
		return err
	}
	opener := opts.FileOpener
	if opener == nil {
		opener = osFileOpener{}
	}
	tw := tar.NewWriter(w)
	defer tw.Close()
	fileSys := os.DirFS(rootDirPath)
	var fileLogMsg bytes.Buffer
	err = fs.WalkDir(fileSys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		absPath := filepath.Join(rootDirPath, path)
		info, err := os.Stat(absPath)
		if err != nil {
			return err
		}
		name := path
		if d.IsDir() {
			name += fileSeparatorString
		}
		if opts.Ignorer != nil {
			if opts.Ignorer.Ignore(path) {
				fmt.Fprintln(&fileLogMsg, "- ", name)
				if info.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
		}
		fmt.Fprintln(&fileLogMsg, "+ ", name)
		isFile := info.Mode().Type() == 0
		var size int64
		if isFile {
			size = info.Size()
		}
		tw.WriteHeader(&tar.Header{
			Name:    name,
			Mode:    int64(info.Mode()),
			Size:    size,
			ModTime: info.ModTime(),
		})
		if isFile {
			// TODO: The size must be accurate... Must buffer the files
			file, err := opener.OpenFile(absPath)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tw, file)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	log.Debug().Messagef("Tardump includes:\n%s", fileLogMsg.String())
	return nil
}
