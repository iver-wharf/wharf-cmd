package slices

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReverseStrings(t *testing.T) {
	slice := []string{"a", "b", "c", "d", "e"}
	ReverseStrings(slice)
	want := []string{"e", "d", "c", "b", "a"}
	assert.Equal(t, want, slice)
}
