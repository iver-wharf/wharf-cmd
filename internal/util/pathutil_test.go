package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUseShorthandHomeDir(t *testing.T) {
	home := "/home/root"
	path := "/home/root/.wharf-vars.yml"
	want := "~/.wharf-vars.yml"
	got := useShorthandHomePrefix(path, home)

	assert.Equal(t, want, got)
}
