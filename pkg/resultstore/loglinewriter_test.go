package resultstore

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
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
	assert.Equal(t, uint64(2), w.lastLogID)
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

func TestStore_OpenLogWriterCollision(t *testing.T) {
	s := NewStore(mockFS{
		openRead: func(string) (io.ReadCloser, error) {
			return nil, fs.ErrNotExist
		},
		openAppend: func(name string) (io.WriteCloser, error) {
			return nopWriteCloser{}, nil
		},
	})
	const stepID uint64 = 1
	w1, err := s.OpenLogWriter(stepID)
	require.NoError(t, err, "open writer 1")

	_, err = s.OpenLogWriter(stepID)
	require.ErrorIs(t, err, ErrLogWriterAlreadyOpen, "open writer 2, expect collision")

	w1.Close()
	_, err = s.OpenLogWriter(stepID)
	require.NoError(t, err, "open writer 2, expect no collision")
}

func TestStore_OpenLogWriterGetListOfWriters(t *testing.T) {
	s := NewStore(mockFS{
		openRead: func(string) (io.ReadCloser, error) {
			return nil, fs.ErrNotExist
		},
		openAppend: func(name string) (io.WriteCloser, error) {
			return nopWriteCloser{}, nil
		},
	}).(*store)
	w1, err := s.OpenLogWriter(1)
	require.NoError(t, err, "open writer 1")

	_, err = s.OpenLogWriter(2)
	require.NoError(t, err, "open writer 2")

	_, err = s.OpenLogWriter(3)
	require.NoError(t, err, "open writer 3")

	w1.Close()

	_, ok := s.getLogWriter(1)
	assert.Equal(t, false, ok, "get writer 1 ok?")
	w2, ok := s.getLogWriter(2)
	assert.Equal(t, true, ok, "get writer 2 ok?")
	assert.NotNil(t, w2, "get writer 2")
	w3, ok := s.getLogWriter(3)
	assert.Equal(t, true, ok, "get writer 3 ok?")
	assert.NotNil(t, w3, "get writer 3")
}

func TestStore_OpenLogWriterUsesLastLogLineID(t *testing.T) {
	buf := bytes.NewBufferString(fmt.Sprintf(`%[1]s Foo bar 1
%[1]s Moo doo 2
%[1]s Faz 3
%[1]s Baz 4
%[1]s Boo 5
%[1]s Foz 6
%[1]s Roo 7
%[1]s Goo 8
`, sampleTimeStr))
	s := NewStore(mockFS{
		openRead: func(string) (io.ReadCloser, error) {
			return io.NopCloser(buf), nil
		},
		openAppend: func(name string) (io.WriteCloser, error) {
			return nopWriteCloser{}, nil
		},
	})
	const stepID uint64 = 1
	w, err := s.OpenLogWriter(stepID)
	require.NoError(t, err, "open writer")
	assert.Equal(t, uint64(8), w.(*logLineWriteCloser).lastLogID)

	err = w.WriteLogLine("Hello 9")
	require.NoError(t, err, "write line")
	assert.Equal(t, uint64(9), w.(*logLineWriteCloser).lastLogID)
}
