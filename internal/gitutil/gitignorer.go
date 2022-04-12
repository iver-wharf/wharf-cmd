package gitutil

import (
	"path/filepath"

	"github.com/denormal/go-gitignore"
	"github.com/iver-wharf/wharf-cmd/internal/ignorer"
)

// NewIgnorer creates a new ignorer that checks if a file should be ignored or
// not depending on the .gitignore files found inside the repository.
func NewIgnorer(currentDir, repoRoot string) (ignorer.Ignorer, error) {
	repo, err := gitignore.NewRepository(repoRoot)
	if err != nil {
		return nil, err
	}
	return &gitIgnorer{currentDir, repo}, nil
}

type gitIgnorer struct {
	currentDir string
	repo       gitignore.GitIgnore
}

func (i *gitIgnorer) Ignore(relPath string) bool {
	// NOTE: gitignore.GitIgnore has a .Ignore function, but that function
	// isn't implemented on the repository level. Therefore we need to
	// re-implement the gitignore.GitIgnore.Ignore() function here via the
	// .Match function that is implemented on the repository level.
	match := i.repo.Match(filepath.Join(i.currentDir, relPath))
	if match == nil {
		return false
	}
	return match.Ignore()
}
