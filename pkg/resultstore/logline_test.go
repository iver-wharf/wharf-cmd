package resultstore

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_AddFirstLogLine(t *testing.T) {
	var buf bytes.Buffer
	fs := mockFS{
		openAppend: func(string) (io.WriteCloser, error) {
			return nopWriteCloser{&buf}, nil
		},
	}
	s := NewStore(fs)
	stepID := uint64(1)
	s.AddLogLine(stepID, "2021-11-24T11:22:08.800Z Foo bar")

	want := "2021-11-24T11:22:08.800Z Foo bar\n"
	got := buf.String()
	assert.Equal(t, want, got)
}

func TestStore_AddSecondLogLine(t *testing.T) {
	buf := bytes.NewBufferString("hello world\n")
	fs := mockFS{
		openAppend: func(string) (io.WriteCloser, error) {
			return nopWriteCloser{buf}, nil
		},
	}
	s := NewStore(fs)
	stepID := uint64(1)
	s.AddLogLine(stepID, "2021-11-24T11:22:08.800Z Foo bar")

	want := "hello world\n2021-11-24T11:22:08.800Z Foo bar\n"
	got := buf.String()
	assert.Equal(t, want, got)
}

func TestStore_ReadAllLogLines(t *testing.T) {
	buf := bytes.NewBufferString(`2021-11-24T11:22:08.800Z Foo bar
2021-11-24T11:22:08.800Z Moo doo
2021-11-24T11:22:08.800Z Baz taz`)
	fs := mockFS{
		openRead: func(name string) (io.ReadCloser, error) {
			return io.NopCloser(buf), nil
		},
	}
	s := NewStore(fs)
	stepID := uint64(1)
	got, err := s.ReadAllLogLines(stepID)
	require.NoError(t, err)
	wantTime := time.Date(2021, 11, 24, 11, 22, 8, 800000000, time.UTC)
	want := []LogLine{
		{StepID: stepID, LogID: 1, Line: "Foo bar", Timestamp: wantTime},
		{StepID: stepID, LogID: 2, Line: "Moo doo", Timestamp: wantTime},
		{StepID: stepID, LogID: 3, Line: "Baz taz", Timestamp: wantTime},
	}
	assert.Equal(t, want, got)
}

func TestParseLogLine(t *testing.T) {
	sampleTimeStr := "2021-05-09T12:13:14.1234Z"
	sampleTime := time.Date(2021, 5, 9, 12, 13, 14, 123400000, time.UTC)
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
