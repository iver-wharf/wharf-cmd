package resultstore

import (
	"sync"

	"gopkg.in/typ.v3/pkg/sets"
)

// NewSyncSet creates a new empty sync set.
func NewSyncSet[T comparable]() SyncSet[T] {
	return NewSyncSetFromSet(make(sets.Set[T]))
}

// NewSyncSetFromSet creates a new sync set that is a clone of the provided set.
func NewSyncSetFromSet[T comparable](set sets.Set[T]) SyncSet[T] {
	return syncSet[T]{
		s:  set.Clone(),
		mu: &sync.RWMutex{},
	}
}

// SyncSet is an interface for thread-safe wrappers of a set.
type SyncSet[T comparable] interface {
	// String converts this set to its string representation.
	String() string
	// Has returns true if the value exists in the set.
	Has(value T) bool
	// Add will add an element to the set, and return true if it was added
	// or false if the value already existed in the set.
	Add(value T) bool
	// AddSet will add all element found in specified set to this set, and
	// return the number of values that was added.
	AddSet(set SyncSet[T]) int
	// Remove will remove an element from the set, and return true if it was removed
	// or false if no such value existed in the set.
	Remove(value T) bool
	// RemoveSet will remove all element found in specified set from this set, and
	// return the number of values that was removed.
	RemoveSet(set SyncSet[T]) int
	// Clone returns a copy of the set.
	Clone() SyncSet[T]
	// Slice returns a new slice of all values in the set.
	Slice() []T
	// Intersect performs an "intersection" on the sets and returns a new set.
	// An intersection is a set of all elements that appear in both sets. In
	// mathmatics it's denoted as:
	// 	A ∩ B
	// Example:
	// 	{1 2 3} ∩ {3 4 5} = {3}
	// This operation is commutative, meaning you will get the same result no matter
	// the order of the operands. In other words:
	// 	A.Intersect(B) == B.Intersect(A)
	Intersect(set SyncSet[T]) SyncSet[T]
	// Union performs a "union" on the sets and returns a new set.
	// A union is a set of all elements that appear in either set. In mathmatics
	// it's denoted as:
	// 	A ∪ B
	// Example:
	// 	{1 2 3} ∪ {3 4 5} = {1 2 3 4 5}
	// This operation is commutative, meaning you will get the same result no matter
	// the order of the operands. In other words:
	// 	A.Union(B) == B.Union(A)
	Union(set SyncSet[T]) SyncSet[T]
	// SetDiff performs a "set difference" on the sets and returns a new set.
	// A set difference resembles a subtraction, where the result is a set of all
	// elements that appears in the first set but not in the second. In mathmatics
	// it's denoted as:
	// 	A \ B
	// Example:
	// 	{1 2 3} \ {3 4 5} = {1 2}
	// This operation is noncommutative, meaning you will get different results
	// depending on the order of the operands. In other words:
	// 	A.SetDiff(B) != B.SetDiff(A)
	SetDiff(set SyncSet[T]) SyncSet[T]
	// SymDiff performs a "symmetric difference" on the sets and returns a new set.
	// A symmetric difference is the set of all elements that appear in either of
	// the sets, but not both. In mathmatics it's commonly denoted as either:
	// 	A △ B
	// or
	// 	A ⊖ B
	// Example:
	// 	{1 2 3} ⊖ {3 4 5} = {1 2 4 5}
	// This operation is commutative, meaning you will get the same result no matter
	// the order of the operands. In other words:
	// 	A.SymDiff(B) == B.SymDiff(A)
	SymDiff(set SyncSet[T]) SyncSet[T]

	lock()
	rLock()
	unlock()
	rUnlock()
	underlying() sets.Set[T]
}

// SyncSet is a wrapper for gopkg.in/typ.v3/pkg/sets.Set that is safe to use in
// a multithreaded environment.
type syncSet[T comparable] struct {
	s  sets.Set[T]
	mu *sync.RWMutex
}

func (s syncSet[T]) String() string {
	s.rLock()
	str := s.s.String()
	s.rUnlock()
	return str
}

func (s syncSet[T]) Has(value T) bool {
	s.rLock()
	ok := s.s.Has(value)
	s.rUnlock()
	return ok
}

func (s syncSet[T]) Add(value T) bool {
	s.lock()
	ok := s.s.Add(value)
	s.unlock()
	return ok
}

func (s syncSet[T]) AddSet(set SyncSet[T]) int {
	set.rLock()
	s.lock()
	numAdded := s.s.AddSet(set.underlying())
	s.unlock()
	set.rUnlock()
	return numAdded
}

func (s syncSet[T]) Remove(value T) bool {
	s.lock()
	ok := s.s.Remove(value)
	s.unlock()
	return ok
}

func (s syncSet[T]) RemoveSet(set SyncSet[T]) int {
	set.rLock()
	s.lock()
	numRemoved := s.s.RemoveSet(set.underlying())
	s.unlock()
	set.rUnlock()
	return numRemoved
}

func (s syncSet[T]) Clone() SyncSet[T] {
	s.rLock()
	clone := NewSyncSetFromSet(s.s)
	s.rUnlock()
	return clone
}

func (s syncSet[T]) Slice() []T {
	s.rLock()
	slice := s.s.Slice()
	s.rUnlock()
	return slice
}

func (s syncSet[T]) Intersect(set SyncSet[T]) SyncSet[T] {
	s.rLock()
	set.rLock()
	intersection := s.s.Intersect(set.underlying())
	set.rUnlock()
	s.rUnlock()
	return syncSet[T]{
		s:  intersection,
		mu: &sync.RWMutex{},
	}
}

func (s syncSet[T]) Union(set SyncSet[T]) SyncSet[T] {
	s.rLock()
	set.rLock()
	union := s.s.Union(set.underlying())
	set.rUnlock()
	s.rUnlock()
	return syncSet[T]{
		s:  union,
		mu: &sync.RWMutex{},
	}
}

func (s syncSet[T]) SetDiff(set SyncSet[T]) SyncSet[T] {
	s.rLock()
	set.rLock()
	setDiff := s.s.SetDiff(set.underlying())
	set.rUnlock()
	s.rUnlock()
	return syncSet[T]{
		s:  setDiff,
		mu: &sync.RWMutex{},
	}
}

func (s syncSet[T]) SymDiff(set SyncSet[T]) SyncSet[T] {
	s.rLock()
	set.rLock()
	union := s.s.SymDiff(set.underlying())
	set.rUnlock()
	s.rUnlock()
	return &syncSet[T]{
		s:  union,
		mu: &sync.RWMutex{},
	}
}

func (s syncSet[T]) lock() {
	s.mu.Lock()
}

func (s syncSet[T]) rLock() {
	s.mu.RLock()
}

func (s syncSet[T]) unlock() {
	s.mu.Unlock()
}

func (s syncSet[T]) rUnlock() {
	s.mu.RUnlock()
}

func (s syncSet[T]) underlying() sets.Set[T] {
	return s.s
}

// CartesianProduct performs a "Cartesian product" on two sets and returns a new
// set. A Cartesian product of two sets is a set of all possible combinations
// between two sets. In mathmatics it's denoted as:
// 	A × B
// Example:
// 	{1 2 3} × {a b c} = {1a 1b 1c 2a 2b 2c 3a 3b 3c}
// This operation is noncommutative, meaning you will get different results
// depending on the order of the operands. In other words:
// 	A.CartesianProduct(B) != B.CartesianProduct(A)
// This noncommutative attribute of the Cartesian product operation is due to
// the pairs being in reverse order if you reverse the order of the operands.
// Example:
// 	{1 2 3} × {a b c} = {1a 1b 1c 2a 2b 2c 3a 3b 3c}
// 	{a b c} × {1 2 3} = {a1 a2 a3 b1 b2 b3 c1 c2 c3}
// 	{1a 1b 1c 2a 2b 2c 3a 3b 3c} != {a1 a2 a3 b1 b2 b3 c1 c2 c3}
func CartesianProduct[TA comparable, TB comparable](a SyncSet[TA], b SyncSet[TB]) SyncSet[Product[TA, TB]] {
	result := make(sets.Set[Product[TA, TB]])
	a.rLock()
	b.rLock()
	for valueA := range a.underlying() {
		for valueB := range b.underlying() {
			result.Add(Product[TA, TB]{valueA, valueB})
		}
	}
	b.rUnlock()
	a.rUnlock()
	return NewSyncSetFromSet(result)
}

// Product is the resulting type from a Cartesian product operation.
type Product[TA comparable, TB comparable] sets.Product[TA, TB]
