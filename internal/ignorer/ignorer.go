package ignorer

// Ignorer is an interface for conditionally ignoring files or directory trees
// when creating a tarball.
type Ignorer interface {
	// Ignore returns true to ignore a file, and false to include the file.
	Ignore(absPath, relPath string) bool
}
