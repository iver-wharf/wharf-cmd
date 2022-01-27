package tarutil

import (
	"archive/tar"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func Dir(w io.Writer, filesFromDir string) error {
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
		name := path
		size := info.Size()
		if d.IsDir() {
			name += "/"
			size = 0
		}
		tw.WriteHeader(&tar.Header{
			Name:    name,
			Mode:    int64(info.Mode()),
			Size:    size,
			ModTime: info.ModTime(),
		})
		if info.Mode().Type() == 0 {
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
	//return err
}
