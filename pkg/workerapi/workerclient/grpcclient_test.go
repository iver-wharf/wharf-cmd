package workerclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenBadAddress(t *testing.T) {
	c := newGRPCClient("badaddress", ClientOptions{InsecureSkipVerify: false})
	assert.Error(t, c.open())
	assert.NoError(t, c.close())
}

func TestOpeningTwiceGivesErrAlreadyOpen(t *testing.T) {
	c := newGRPCClient("127.0.0.1", ClientOptions{InsecureSkipVerify: true})
	assert.NoError(t, c.open())
	assert.ErrorIs(t, c.open(), errAlreadyOpen)
	assert.NoError(t, c.close())
}

func TestClosing(t *testing.T) {
	c := newGRPCClient("127.0.0.1", ClientOptions{InsecureSkipVerify: true})
	assert.NoError(t, c.open())
	assert.NotNil(t, c.conn)
	assert.NoError(t, c.close())
	assert.Nil(t, c.conn)
}
