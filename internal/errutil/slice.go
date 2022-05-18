package errutil

// Slice is a slice of errors.
type Slice []error

// Add appends another error to this slice of errors. Nil errors are ignored.
func (s *Slice) Add(errs ...error) {
	for _, err := range errs {
		if err == nil {
			continue
		}
		*s = append(*s, err)
	}
}
