package tarutil

import (
	"archive/tar"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// Dir will recursively tar the contents of an entire directory. Hidden files
// (files that start with a dot) are included. The name of the target directory
// is not included in the tarball, but instead only the children.
func Dir(w io.Writer, filesFromDir string) error {
	return DirIgnore(w, filesFromDir, nil)
}

// Ignorer is an interface for conditionally ignoring files or directory trees
// when creating a tarball.
type Ignorer interface {
	// Ignore returns true to ignore a file, and false to include the file.
	Ignore(path string) bool
}

// DirIgnore will recursively tar the contents of an entire directory, and allow
// ignoring directory trees using the Ignorer interface. Hidden files
// (files that start with a dot) are included. The name of the target directory
// is not included in the tarball, but instead only the children.
func DirIgnore(w io.Writer, filesFromDir string, ignorer Ignorer) error {
	tw := tar.NewWriter(w)
	fileSys := os.DirFS(filesFromDir)
	err := fs.WalkDir(fileSys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		realPath := filepath.Join(filesFromDir, path)
		info, err := os.Stat(realPath)
		if err != nil {
			return err
		}
		if ignorer != nil && ignorer.Ignore(path) {
			if info.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		name := path
		if d.IsDir() {
			name += "/"
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
			file, err := os.Open(realPath)
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
	return tw.Close()
}