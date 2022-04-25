package resultstore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncSlice_Range(t *testing.T) {
	var s syncSlice[[]int, int]
	s.Append(34)
	s.Append(35)
	s.Append(20)
	s.Append(4)
	assert.Len(t, s.values, 4, "length")

	var indices []int
	var values []int
	s.Range(func(i, v int) bool {
		indices = append(indices, i)
		values = append(values, v)
		return true
	})

	assert.ElementsMatch(t, []int{0, 1, 2, 3}, indices, "indices whole")
	assert.ElementsMatch(t, []int{34, 35, 20, 4}, values, "values whole")

	indices = make([]int, 0)
	values = make([]int, 0)
	s.Range(func(i, v int) bool {
		indices = append(indices, i)
		values = append(values, v)
		return len(values) != 2
	})

	assert.ElementsMatch(t, []int{0, 1}, indices, "indices with break")
	assert.ElementsMatch(t, []int{34, 35}, values, "values with break")
}

func TestSyncSlice_Remove(t *testing.T) {
	var s syncSlice[[]int, int]
	s.Append(34)
	s.Append(35)
	s.Append(20)
	s.Append(4)
	assert.Len(t, s.values, 4, "length before remove")

	s.Remove(4)
	assert.Len(t, s.values, 3, "length after remove")
	assert.ElementsMatch(t, []int{34, 35, 20}, s.values, "elements after remove")
}

func TestSyncSlice_RemoveNotFound(t *testing.T) {
	var s syncSlice[[]int, int]
	s.Append(34)
	s.Append(35)
	s.Append(20)
	s.Append(4)
	assert.Len(t, s.values, 4, "length before failed remove")

	s.Remove(1000)
	assert.Len(t, s.values, 4, "length after failed remove")
	assert.ElementsMatch(t, []int{34, 35, 20, 4}, s.values, "elements after failed remove")
}
