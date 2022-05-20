package tarstore

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/iver-wharf/wharf-cmd/internal/filecopy"
	"github.com/iver-wharf/wharf-cmd/internal/ignorer"
	"github.com/iver-wharf/wharf-cmd/internal/tarutil"
	"github.com/iver-wharf/wharf-core/v2/pkg/logger"
	"gopkg.in/typ.v4/sync2"
)

var log = logger.NewScoped("TARSTORE")

const dirFileMode fs.FileMode = 0775

// Tarball is an identifier for a tarball file containing a repository.
type Tarball string

// Open creates a file handle to the tarball.
func (t Tarball) Open() (io.ReadCloser, error) {
	return os.Open(string(t))
}

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
	if id == "" {
		return "", errors.New("tarball name cannot be empty")
	}
	once, _ := s.onceMap.LoadOrStore(id, new(sync2.Once2[Tarball, error]))
	return once.Do(func() (Tarball, error) {
		return s.prepare(copier, ignorer, id)
	})
}

func (s *store) prepare(copier filecopy.Copier, ignorer ignorer.Ignorer, id string) (Tarball, error) {
	dstPath := filepath.Join(s.tmpPath, id)
	if err := os.MkdirAll(dstPath, dirFileMode); err != nil {
		return "", err
	}
	log.Info().
		WithString("src", s.srcPath).
		WithString("dst", dstPath).
		Message("Copying files.")
	if err := filecopy.CopyDirIgnorer(dstPath, s.srcPath, copier, ignorer); err != nil {
		return "", err
	}
	log.Debug().
		WithString("src", s.srcPath).
		WithString("dst", dstPath).
		Message("Done copying files.")
	tarPath := filepath.Join(s.tmpPath, id+".tar")
	tarFile, err := os.Create(tarPath)
	if err != nil {
		return "", err
	}
	defer tarFile.Close()
	log.Info().
		WithString("path", tarPath).
		Message("Creating tarball.")
	if err := tarutil.Dir(tarFile, dstPath); err != nil {
		return "", err
	}
	log.Debug().
		WithString("path", tarPath).
		Message("Done creating tarball.")
	return Tarball(tarPath), nil
}
