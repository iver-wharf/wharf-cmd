package gitutil

import (
	"github.com/denormal/go-gitignore"
)

// Ignorer is an interface for conditionally ignoring files or directory trees
// based on gitignore files.
type Ignorer interface {
	// Ignore returns true to ignore a file, and false to include the file.
	Ignore(path string) bool
}

// NewIgnorer creates a new ignorer that checks if a file should be ignored or
// not depending on the .gitignore files found inside the repository.
func NewIgnorer(repoRoot string) (Ignorer, error) {
	repo, err := gitignore.NewRepository(repoRoot)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

type ignorer struct {
	repo gitignore.GitIgnore
}

func (i *ignorer) Ignore(path string) bool {
	// NOTE: gitignore.GitIgnore has a .Ignore function, but that function
	// isn't implemented on the repository level. Therefore we need to
	// re-implement the gitignore.GitIgnore.Ignore() function here via the
	// .Match function that is implemented on the repository level.
	match := i.repo.Match(path)
	if match == nil {
		return false
	}
	return match.Ignore()
}
