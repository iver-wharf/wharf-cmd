package tarutil

import (
	"archive/tar"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/iver-wharf/wharf-core/pkg/logger"
)

var log = logger.NewScoped("TAR")

var fileSeparatorString = string(filepath.Separator)

// Dir will recursively tar the contents of an entire directory. Hidden files
// (files that start with a dot) are included. The name of the target directory
// is not included in the tarball, but instead only the children.
func Dir(w io.Writer, dirPath string) error {
	rootDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		return err
	}
	tw := tar.NewWriter(w)
	defer tw.Close()
	fileSys := os.DirFS(rootDirPath)
	return fs.WalkDir(fileSys, ".", func(path string, d fs.DirEntry, err error) error {
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
			file, err := os.Open(absPath)
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
}
