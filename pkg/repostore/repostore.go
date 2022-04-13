package repostore

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/iver-wharf/wharf-cmd/internal/filecopy"
	"github.com/iver-wharf/wharf-cmd/internal/ignorer"
	"github.com/iver-wharf/wharf-cmd/internal/tarutil"
	"gopkg.in/typ.v3/pkg/sync2"
)

const dirFileMode fs.FileMode = 0775

// Store is an interface for copying and tar'ing repositories.
type Store interface {
	io.Closer

	GetPreparedTarball(copier filecopy.Copier, ignorer ignorer.Ignorer, id string) (Tarball, error)
}

// New creates a new Store with a given directory path as the repo root.
// The repo root does not have to be a Git root.
func New(srcPath string) (Store, error) {
	tmpPath, err := os.MkdirTemp("", "wharf-cmd-repo-")
	if err != nil {
		return nil, err
	}
	return &store{
		tmpPath: tmpPath,
		srcPath: srcPath,
	}, nil
}

type store struct {
	tmpPath string
	srcPath string
	onceMap sync2.Map[string, *sync2.Once2[Tarball, error]]
}

func (s *store) Close() error {
	return os.RemoveAll(s.tmpPath)
}

func (s *store) GetPreparedTarball(copier filecopy.Copier, ignorer ignorer.Ignorer, id string) (Tarball, error) {
	once, _ := s.onceMap.LoadOrStore(id, new(sync2.Once2[Tarball, error]))
	return once.Do(func() (Tarball, error) {
		return s.prepare(copier, ignorer, id)
	})
}

func (s *store) prepare(copier filecopy.Copier, ignorer ignorer.Ignorer, id string) (Tarball, error) {
	dstDir := filepath.Join(s.tmpPath, id)
	if err := os.MkdirAll(dstDir, dirFileMode); err != nil {
		return "", err
	}
	if err := filecopy.CopyDirIgnorer(dstDir, s.srcPath, copier, ignorer); err != nil {
		return "", err
	}
	tarPath := filepath.Join(s.tmpPath, id+".tar")
	tarFile, err := os.Create(tarPath)
	if err != nil {
		return "", err
	}
	defer tarFile.Close()
	if err := tarutil.Dir(tarFile, dstDir); err != nil {
		return "", err
	}
	return Tarball(tarPath), nil
}
