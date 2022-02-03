package resultstore

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLogLine(t *testing.T) {
	zeroTime := time.Time{}
	testCases := []struct {
		name     string
		input    string
		wantTime time.Time
		wantLine string
	}{
		{
			name:     "empty line",
			input:    "",
			wantTime: zeroTime,
			wantLine: "",
		},
		{
			name:     "missing time",
			input:    "hello world",
			wantTime: zeroTime,
			wantLine: "hello world",
		},
		{
			name:     "with time",
			input:    sampleTimeStr + " hello world",
			wantTime: sampleTime,
			wantLine: "hello world",
		},
		{
			name:     "invalid time",
			input:    "2021-99-09T55:13:65.1234Z hello world",
			wantTime: zeroTime,
			wantLine: "2021-99-09T55:13:65.1234Z hello world",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tim, line := parseLogLine(tc.input)
			assert.Equal(t, tc.wantLine, line, "log line")
			assert.Equal(t, tc.wantTime, tim, "log time")
		})
	}
}

func TestSanitizeLogLine(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "no changes",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "newlines",
			input: "hello\n\r world",
			want:  `hello\n\r world`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeLogLine(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestStore_SubUnsubLogLines(t *testing.T) {
	s := NewStore(mockFS{
		listDirEntries: func(name string) ([]fs.DirEntry, error) {
			return nil, nil
		},
	}).(*store)
	require.Empty(t, s.logSubs, "before sub")
	const buffer = 0
	ch := subLogLinesNoErr(t, s, buffer)
	require.Len(t, s.logSubs, 1, "after sub")
	assert.True(t, s.logSubs[0] == ch, "after sub")
	require.True(t, s.UnsubAllLogLines(ch), "unsub success")
	assert.Empty(t, s.logSubs, "after unsub")
}

func TestStore_UnsubLogLinesMiddle(t *testing.T) {
	s := NewStore(mockFS{
		listDirEntries: func(name string) ([]fs.DirEntry, error) {
			return nil, nil
		},
	}).(*store)
	require.Empty(t, s.logSubs, "before sub")
	const buffer = 0
	chs := []<-chan LogLine{
		subLogLinesNoErr(t, s, buffer),
		subLogLinesNoErr(t, s, buffer),
		subLogLinesNoErr(t, s, buffer),
		subLogLinesNoErr(t, s, buffer),
		subLogLinesNoErr(t, s, buffer),
	}
	require.Len(t, s.logSubs, 5, "after sub")
	require.True(t, s.UnsubAllLogLines(chs[2]), "unsub success")
	require.Len(t, s.logSubs, 4, "after unsub")
	want := []<-chan LogLine{
		chs[0], chs[1], chs[3], chs[4],
	}
	for i, ch := range want {
		assert.Truef(t, ch == s.logSubs[i], "index %d, %v != %v", i, ch, s.logSubs[i])
	}
}

func TestStore_PubSubLogLines(t *testing.T) {
	s := NewStore(mockFS{
		listDirEntries: func(name string) ([]fs.DirEntry, error) {
			return nil, nil
		},
		openRead: func(string) (io.ReadCloser, error) {
			return nil, fs.ErrNotExist
		},
		openAppend: func(name string) (io.WriteCloser, error) {
			return nopWriteCloser{}, nil
		},
	})
	const buffer = 1
	const stepID uint64 = 1
	ch := subLogLinesNoErr(t, s, buffer)
	require.NotNil(t, ch, "channel")
	w, err := s.OpenLogWriter(stepID)
	require.NoError(t, err)
	w.WriteLogLine(sampleTimeStr + " Hello there")
	w.Close()

	select {
	case got, ok := <-ch:
		require.True(t, ok, "received on channel")
		want := LogLine{
			StepID:    stepID,
			LogID:     1,
			Message:   "Hello there",
			Timestamp: sampleTime,
		}
		assert.Equal(t, want, got)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestStore_SubAllLogLinesSendsNonNewlyWrittenLogs(t *testing.T) {
	str := fmt.Sprintf(`%[1]s Foo bar 1
%[1]s Moo doo 2
%[1]s Faz 3
%[1]s Baz 4
%[1]s Boo 5
%[1]s Foz 6
%[1]s Roo 7
%[1]s Goo 8
`, sampleTimeStr)
	s := NewStore(mockFS{
		listDirEntries: func(name string) ([]fs.DirEntry, error) {
			if name != dirNameSteps {
				return nil, nil
			}
			return []fs.DirEntry{
				newMockDirEntryDir("1"),
			}, nil
		},
		openRead: func(string) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewBufferString(str)), nil
		},
		openAppend: func(name string) (io.WriteCloser, error) {
			return nopWriteCloser{}, nil
		},
	}).(*store)
	const stepID uint64 = 1
	const buffer = 10

	_, err := s.OpenLogWriter(stepID)
	require.NoError(t, err, "open writer")
	w, _ := s.getLogWriter(stepID)
	require.NotNil(t, w, "get writer")
	w.lastLogID = 5

	ch := subLogLinesNoErr(t, s, buffer)
	require.NotNil(t, ch, "channel")

	var logIDsRecieved []uint64
	for len(logIDsRecieved) < 5 {
		select {
		case got, ok := <-ch:
			require.True(t, ok, "received on channel")
			logIDsRecieved = append(logIDsRecieved, got.LogID)
		case <-time.After(time.Second):
			t.Fatalf("timeout, received %d logs", len(logIDsRecieved))
		}
	}
	select {
	case got := <-ch:
		t.Fatal("received too many logs, latest:", got)
	default:
		// OK
	}
	wantIDs := []uint64{1, 2, 3, 4, 5}
	assert.Equal(t, wantIDs, logIDsRecieved, "log IDs received")
}

func subLogLinesNoErr(t *testing.T, s Store, buffer int) <-chan LogLine {
	ch, err := s.SubAllLogLines(buffer)
	require.NoError(t, err, "sub logs: err")
	require.NotNil(t, ch, "sub logs: chan")
	return ch
}
