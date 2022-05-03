package resultstore

import (
	"sync"

	"gopkg.in/typ.v4/sets"
)

func newSyncSet[T comparable]() syncSet[T] {
	return syncSet[T]{
		s:  make(sets.Set[T]),
		mu: &sync.RWMutex{},
	}
}

// syncSet is a wrapper for gopkg.in/typ.v4/sets.Set that is safe to use in
// a multithreaded environment.
type syncSet[T comparable] struct {
	s  sets.Set[T]
	mu *sync.RWMutex
}

func (s syncSet[T]) Add(value T) bool {
	s.mu.Lock()
	ok := s.s.Add(value)
	s.mu.Unlock()
	return ok
}

func (s syncSet[T]) Remove(value T) bool {
	s.mu.Lock()
	ok := s.s.Remove(value)
	s.mu.Unlock()
	return ok
}

func (s syncSet[T]) Slice() []T {
	s.mu.RLock()
	slice := s.s.Slice()
	s.mu.RUnlock()
	return slice
}
