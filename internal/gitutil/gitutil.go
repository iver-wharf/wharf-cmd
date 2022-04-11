package gitutil

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

var (
	ErrNotAGitDir = errors.New("not a Git directory")
)

// GitRepoRoot looks recursively upwards for the Git repository root directory
// using a fs.StatFS.
func GitRepoRootFS(dir string, statFS fs.StatFS) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	currentDir := absDir
	for {
		info, err := statFS.Stat(filepath.Join(currentDir, ".git"))
		if err == nil && info.IsDir() {
			return currentDir, nil
		}
		oldDir := currentDir
		currentDir = filepath.Dir(currentDir)
		if oldDir == currentDir {
			return "", ErrNotAGitDir
		}
	}
}

// IsGitRepoFS checks recursively upwards if a directory is inside a Git
// repository using a fs.StatFS.
func IsGitRepoFS(dir string, statFS fs.StatFS) (bool, error) {
	_, err := GitRepoRootFS(dir, statFS)
	if err == ErrNotAGitDir {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

type osStatFS struct{}

func (osStatFS) Open(name string) (fs.File, error) {
	return os.Open(name)
}

func (osStatFS) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

// GitRepoRoot looks recursively upwards for the Git repository root directory
// using the file system from the OS.
func GitRepoRoot(dir string) (string, error) {
	return GitRepoRootFS(dir, osStatFS{})
}

// IsGitRepo checks recursively upwards if a directory is inside a Git
// repository using the file system from the OS.
func IsGitRepo(dir string) (bool, error) {
	return IsGitRepoFS(dir, osStatFS{})
}
