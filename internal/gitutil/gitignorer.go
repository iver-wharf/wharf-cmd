package gitutil

import (
	"github.com/denormal/go-gitignore"
	"github.com/iver-wharf/wharf-cmd/internal/ignorer"
)

// NewIgnorer creates a new ignorer that checks if a file should be ignored or
// not depending on the .gitignore files found inside the repository.
func NewIgnorer(repoRoot string) (ignorer.Ignorer, error) {
	repo, err := gitignore.NewRepository(repoRoot)
	if err != nil {
		return nil, err
	}
	return &gitIgnorer{repo}, nil
}

type gitIgnorer struct {
	repo gitignore.GitIgnore
}

func (i *gitIgnorer) Ignore(absPath, _ string) bool {
	// NOTE: gitignore.GitIgnore has a .Ignore function, but that function
	// isn't implemented on the repository level. Therefore we need to
	// re-implement the gitignore.GitIgnore.Ignore() function here via the
	// .Match function that is implemented on the repository level.
	match := i.repo.Match(absPath)
	if match == nil {
		return false
	}
	return match.Ignore()
}
