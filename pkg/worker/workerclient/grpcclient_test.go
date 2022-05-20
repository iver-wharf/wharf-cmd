package workerclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenBadAddress(t *testing.T) {
	c := newGRPCClient("badaddress", Options{InsecureSkipVerify: false})
	assert.Error(t, c.ensureOpen())
	assert.NoError(t, c.close())
}

func TestOpeningTwiceGivesSameConnection(t *testing.T) {
	c := newGRPCClient("127.0.0.1", Options{InsecureSkipVerify: true})
	assert.NoError(t, c.ensureOpen())
	conn := c.conn
	assert.NoError(t, c.ensureOpen())
	assert.Equal(t, c.conn, conn)
	assert.NoError(t, c.close())
}

func TestClosing(t *testing.T) {
	c := newGRPCClient("127.0.0.1", Options{InsecureSkipVerify: true})
	assert.NoError(t, c.ensureOpen())
	assert.NotNil(t, c.conn)
	assert.NoError(t, c.close())
	assert.Nil(t, c.conn)
}
