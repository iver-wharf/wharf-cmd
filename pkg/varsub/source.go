package varsub

// Source is a variable substitution source.
type Source interface {
	// Lookup tries to look up a value based on name and returns that value as
	// well as true on success, or false if the variable was not found.
	Lookup(name string) (any, bool)
}

// SourceSlice is a slice of variable substution sources that act as a source
// itself by returning the first successful lookup.
type SourceSlice []Source

// Lookup tries to look up a value based on name and returns that value as
// well as true on success, or false if the variable was not found.
func (ss SourceSlice) Lookup(name string) (any, bool) {
	for _, s := range ss {
		val, ok := s.Lookup(name)
		if ok {
			return val, true
		}
	}
	return nil, false
}

// SourceMap is a variable substitution source based on a map where it uses the
// underlying map as the variable source.
type SourceMap map[string]any

// Lookup tries to look up a value based on name and returns that value as
// well as true on success, or false if the variable was not found.
func (s SourceMap) Lookup(name string) (any, bool) {
	val, ok := s[name]
	return val, ok
}
