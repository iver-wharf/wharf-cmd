package resultstore

import (
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_AddFirstLogLine(t *testing.T) {
	fs := fstest.MapFS(fstest.MapFS{})
	s := NewStore(fs)
	stepID := uint64(1)
	s.AddLogLine(stepID, "2021-11-24T11:22:08.800Z Foo bar")
	gotLines, err := s.ReadAllLogLines(stepID)
	require.NoError(t, err)

	wantDate := time.Date(2021, 11, 24, 11, 22, 8, 800, time.UTC)
	wantLines := []LogLine{
		{StepID: stepID, Line: "Foo bar", Timestamp: wantDate},
	}
	assert.Equal(t, wantLines, gotLines)
}
