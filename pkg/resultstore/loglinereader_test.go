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

func TestLogLineReadCloser(t *testing.T) {
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
