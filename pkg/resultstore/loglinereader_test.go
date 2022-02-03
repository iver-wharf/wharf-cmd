package resultstore

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogLineReadCloser_ReadLogLine(t *testing.T) {
	buf := bytes.NewBufferString(fmt.Sprintf(`%[1]s Foo bar
%[1]s Moo doo`, sampleTimeStr))
	const stepID uint64 = 5
	r := &logLineReadCloser{
		stepID:  stepID,
		closer:  nopCloser{},
		store:   &store{},
		scanner: bufio.NewScanner(buf),
	}
	logLine1, err := r.ReadLogLine()
	require.NoError(t, err, "read 1/2")
	logLine2, err := r.ReadLogLine()
	require.NoError(t, err, "read 2/2")

	logLineUnwanted, err := r.ReadLogLine()
	assert.ErrorIs(t, err, io.EOF, fmt.Sprintf("unexpected result: %v", logLineUnwanted))

	want := []LogLine{
		{StepID: stepID, LogID: 1, Line: "Foo bar", Timestamp: sampleTime},
		{StepID: stepID, LogID: 2, Line: "Moo doo", Timestamp: sampleTime},
	}
	got := []LogLine{logLine1, logLine2}
	assert.Equal(t, want, got)
	assert.Equal(t, uint64(2), r.logID)
}

func TestLogLineReadCloser_ReadLastLogLine(t *testing.T) {
	buf := bytes.NewBufferString(fmt.Sprintf(`%[1]s Foo bar 1
%[1]s Moo doo 2
%[1]s Faz 3
%[1]s Baz 4
%[1]s Boo 5
%[1]s Foz 6
%[1]s Roo 7
%[1]s Goo 8
`, sampleTimeStr))
	const stepID uint64 = 5
	r := &logLineReadCloser{
		stepID:  stepID,
		closer:  nopCloser{},
		store:   &store{},
		scanner: bufio.NewScanner(buf),
	}
	lastLine, err := r.ReadLastLogLine()
	require.NoError(t, err, "read last")
	want := LogLine{
		StepID:    stepID,
		LogID:     8,
		Line:      "Goo 8",
		Timestamp: sampleTime,
	}
	assert.Equal(t, want, lastLine)
}
