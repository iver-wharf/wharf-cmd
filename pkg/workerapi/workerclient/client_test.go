package workerclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	c, err := New("http://127.0.0.1", Options{InsecureSkipVerify: true})
	assert.NoError(t, err)
	assert.NotNil(t, c)
}
