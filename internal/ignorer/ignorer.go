package ignorer

// Ignorer is an interface for conditionally ignoring files or directory trees
// when creating a tarball.
type Ignorer interface {
	// Ignore returns true to ignore a file, and false to include the file.
	Ignore(absPath, relPath string) bool
}

// NewAny returns an Ignorer implementation that returns true if any of the
// provided ignorers return true.
func NewAny(ignorers ...Ignorer) Ignorer {
	return ignoreIfAny(ignorers)
}

type ignoreIfAny []Ignorer

func (m ignoreIfAny) Ignore(absPath, relPath string) bool {
	for _, i := range m {
		if i.Ignore(absPath, relPath) {
			return true
		}
	}
	return false
}
