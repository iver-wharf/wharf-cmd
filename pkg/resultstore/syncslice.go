package resultstore

import "sync"

type syncSlice[S ~[]E, E comparable] struct {
	values S

	mu sync.RWMutex
}

// Append appends a value to the slice in a thread-safe manner.
func (s *syncSlice[S, E]) Append(value E) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = append(s.values, value)
}

func (s *syncSlice[S, E]) Remove(value E) {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := s.Index(value)
	if idx == -1 {
		return
	}
	s.values[idx] = s.values[len(s.values)-1]
	var defaultValue E
	s.values[len(s.values)-1] = defaultValue
	s.values = s.values[:len(s.values)-1]
}

// Index returns the index of a value, or -1 if it was not found.
func (s *syncSlice[S, E]) Index(value E) int {
	for i, v := range s.values {
		if v == value {
			return i
		}
	}
	return -1
}

// Range calls f sequentially for each value present in the slice.
// If f returns false, range stops the iteration.
//
// Attempting to insert/remove elements in the passed in function will create a
// deadlock.
func (s *syncSlice[S, E]) Range(f func(index int, value E) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i, v := range s.values {
		if !f(i, v) {
			return
		}
	}
}
