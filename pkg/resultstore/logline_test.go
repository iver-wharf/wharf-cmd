package resultstore

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogLineWriteCloser(t *testing.T) {
	var buf bytes.Buffer
	var w LogLineWriteCloser = logLineWriteCloser{
		writeCloser: nopWriteCloser{&buf},
		store:       &store{},
	}
	err := w.WriteLogLine(sampleTimeStr + " Foo bar")
	require.NoError(t, err)

	want := sampleTimeStr + " Foo bar\n"
	got := buf.String()
	assert.Equal(t, want, got)
}

func TestLogLineWriteCloser_Sanitizes(t *testing.T) {
	var buf bytes.Buffer
	var w LogLineWriteCloser = logLineWriteCloser{
		writeCloser: nopWriteCloser{&buf},
		store:       &store{},
	}
	err := w.WriteLogLine(sampleTimeStr + " Foo \nbar")
	require.NoError(t, err)

	want := sampleTimeStr + " Foo \\nbar\n"
	got := buf.String()
	assert.Equal(t, want, got)
}

func TestStore_ReadAllLogLines(t *testing.T) {
	buf := bytes.NewBufferString(fmt.Sprintf(`%[1]s Foo bar
%[1]s Moo doo
%[1]s Baz taz`, sampleTimeStr))
	fs := mockFS{
		openRead: func(name string) (io.ReadCloser, error) {
			return io.NopCloser(buf), nil
		},
	}
	s := NewStore(fs)
	const stepID uint64 = 1
	got, err := s.ReadAllLogLines(stepID)
	require.NoError(t, err)
	want := []LogLine{
		{StepID: stepID, LogID: 1, Line: "Foo bar", Timestamp: sampleTime},
		{StepID: stepID, LogID: 2, Line: "Moo doo", Timestamp: sampleTime},
		{StepID: stepID, LogID: 3, Line: "Baz taz", Timestamp: sampleTime},
	}
	assert.Equal(t, want, got)
}

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
	s := NewStore(mockFS{}).(*store)
	require.Empty(t, s.logSubs, "before sub")
	ch := s.SubAllLogLines(0)
	require.Len(t, s.logSubs, 1, "after sub")
	assert.True(t, s.logSubs[0] == ch, "after sub")
	s.UnsubAllLogLines(ch)
	assert.Empty(t, s.logSubs, "after unsub")
}

func TestStore_UnsubLogLinesMiddle(t *testing.T) {
	s := NewStore(mockFS{}).(*store)
	require.Empty(t, s.logSubs, "before sub")
	chs := []<-chan LogLine{
		s.SubAllLogLines(0),
		s.SubAllLogLines(0),
		s.SubAllLogLines(0),
		s.SubAllLogLines(0),
		s.SubAllLogLines(0),
	}
	require.Len(t, s.logSubs, 5, "after sub")
	s.UnsubAllLogLines(chs[2])
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
		openAppend: func(name string) (io.WriteCloser, error) {
			return nopWriteCloser{}, nil
		},
	})
	const buffer = 1
	const stepID uint64 = 1
	ch := s.SubAllLogLines(buffer)
	w, err := s.OpenLogFile(stepID)
	require.NoError(t, err)
	w.WriteLogLine(sampleTimeStr + " Hello there")
	w.Close()

	got := <-ch
	want := LogLine{
		StepID:    stepID,
		LogID:     1,
		Line:      "Hello there",
		Timestamp: sampleTime,
	}
	assert.Equal(t, want, got)
}
