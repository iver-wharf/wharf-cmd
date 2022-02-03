package resultstore

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogLineWriteCloser(t *testing.T) {
	var buf bytes.Buffer
	w := &logLineWriteCloser{
		writeCloser: nopWriteCloser{&buf},
		store:       &store{},
	}
	err := w.WriteLogLine(sampleTimeStr + " Foo bar")
	require.NoError(t, err, "write 1/2")
	err = w.WriteLogLine(sampleTimeStr + " Moo doo")
	require.NoError(t, err, "write 2/2")

	want := sampleTimeStr + " Foo bar\n" + sampleTimeStr + " Moo doo\n"
	got := buf.String()
	assert.Equal(t, want, got)
	assert.Equal(t, uint64(2), w.logID)
}

func TestLogLineWriteCloser_Sanitizes(t *testing.T) {
	var buf bytes.Buffer
	var w LogLineWriteCloser = &logLineWriteCloser{
		writeCloser: nopWriteCloser{&buf},
		store:       &store{},
	}
	err := w.WriteLogLine(sampleTimeStr + " Foo \nbar")
	require.NoError(t, err)

	want := sampleTimeStr + " Foo \\nbar\n"
	got := buf.String()
	assert.Equal(t, want, got)
}
