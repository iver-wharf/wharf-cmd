package resultstore

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"testing"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_ReadStatusUpdatesFile(t *testing.T) {
	buf := bytes.NewBufferString(`
{
	"statusUpdates": [
		{
			"stepId": 1,
			"updateId": 1,
			"timestamp": "2021-05-15T09:01:15.0000Z",
			"status": "Scheduling"
		},
		{
			"stepId": 1,
			"updateId": 2,
			"timestamp": "2021-05-15T09:01:15.0000Z",
			"status": "Running"
		},
		{
			"stepId": 1,
			"updateId": 3,
			"timestamp": "2021-05-15T09:01:15.0000Z",
			"status": "Failed"
		}
	]
}
`)
	wantTime := time.Date(2021, 5, 15, 9, 1, 15, 0, time.UTC)
	const stepID uint64 = 1
	want := StatusList{
		StatusUpdates: []StatusUpdate{
			{
				StepID:    stepID,
				UpdateID:  1,
				Timestamp: wantTime,
				Status:    "Scheduling",
			},
			{
				StepID:    stepID,
				UpdateID:  2,
				Timestamp: wantTime,
				Status:    "Running",
			},
			{
				StepID:    stepID,
				UpdateID:  3,
				Timestamp: wantTime,
				Status:    "Failed",
			},
		},
	}
	s := NewStore(mockFS{
		openRead: func(name string) (io.ReadCloser, error) {
			return io.NopCloser(buf), nil
		},
	}).(*store)
	got, err := s.readStatusUpdatesFile(stepID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestStore_WriteStatusUpdatesFile(t *testing.T) {
	const stepID uint64 = 1
	list := StatusList{
		LastID: 3,
		StatusUpdates: []StatusUpdate{
			{
				StepID:    stepID,
				UpdateID:  1,
				Timestamp: sampleTime,
				Status:    "Scheduling",
			},
			{
				StepID:    stepID,
				UpdateID:  2,
				Timestamp: sampleTime,
				Status:    "Running",
			},
			{
				StepID:    stepID,
				UpdateID:  3,
				Timestamp: sampleTime,
				Status:    "Failed",
			},
		},
	}
	var buf bytes.Buffer
	s := NewStore(mockFS{
		openWrite: func(name string) (io.WriteCloser, error) {
			return nopWriteCloser{&buf}, nil
		},
	}).(*store)
	err := s.writeStatusUpdatesFile(stepID, list)
	require.NoError(t, err)
	want := fmt.Sprintf(`
{
	"lastId": 3,
	"statusUpdates": [
		{
			"stepId": 1,
			"updateId": 1,
			"timestamp": "%[1]s",
			"status": "Scheduling"
		},
		{
			"stepId": 1,
			"updateId": 2,
			"timestamp": "%[1]s",
			"status": "Running"
		},
		{
			"stepId": 1,
			"updateId": 3,
			"timestamp": "%[1]s",
			"status": "Failed"
		}
	]
}`, sampleTimeStr)
	assert.JSONEq(t, want, buf.String())
}

func TestStore_AddStatusUpdateFirst(t *testing.T) {
	var buf bytes.Buffer
	s := NewStore(mockFS{
		openRead: func(name string) (io.ReadCloser, error) {
			return nil, fs.ErrNotExist
		},
		openWrite: func(name string) (io.WriteCloser, error) {
			return nopWriteCloser{&buf}, nil
		},
	})
	const stepID uint64 = 1
	err := s.AddStatusUpdate(stepID, sampleTime, worker.StatusCancelled)
	require.NoError(t, err)
	want := fmt.Sprintf(`
{
	"lastId": 1,
	"statusUpdates": [
		{
			"stepId": 1,
			"updateId": 1,
			"timestamp": "%s",
			"status": "Cancelled"
		}
	]
}`, sampleTimeStr)
	assert.JSONEq(t, want, buf.String())
}

func TestStore_AddStatusUpdateSecond(t *testing.T) {
	buf := bytes.NewBufferString(fmt.Sprintf(`{
	"lastId": 5,
	"statusUpdates": [
		{
			"stepId": 1,
			"updateId": 1,
			"timestamp": "%s",
			"status": "Scheduling"
		}
	]
}`, sampleTimeStr))
	s := NewStore(mockFS{
		openRead: func(name string) (io.ReadCloser, error) {
			return io.NopCloser(buf), nil
		},
		openWrite: func(name string) (io.WriteCloser, error) {
			return nopWriteCloser{buf}, nil
		},
	})
	const stepID uint64 = 1
	err := s.AddStatusUpdate(stepID, sampleTime, worker.StatusCancelled)
	require.NoError(t, err)
	want := fmt.Sprintf(`
{
	"lastId": 6,
	"statusUpdates": [
		{
			"stepId": 1,
			"updateId": 1,
			"timestamp": "%[1]s",
			"status": "Scheduling"
		},
		{
			"stepId": 1,
			"updateId": 6,
			"timestamp": "%[1]s",
			"status": "Cancelled"
		}
	]
}`, sampleTimeStr)
	assert.JSONEq(t, want, buf.String())
}

func TestStore_AddStatusUpdateSkipIfSameStatus(t *testing.T) {
	content := `{
	"statusUpdates": [
		{
			"stepId": 1,
			"updateId": 1,
			"timestamp": "2021-05-15T09:01:15Z",
			"status": "Cancelled"
		}
	]
}`
	s := NewStore(mockFS{
		openRead: func(name string) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader([]byte(content))), nil
		},
		openWrite: func(name string) (io.WriteCloser, error) {
			return nil, errors.New("should not write")
		},
	})
	err := s.AddStatusUpdate(1, time.Now(), worker.StatusCancelled)
	require.NoError(t, err)
}

func TestStore_SubUnsubStatusUpdates(t *testing.T) {
	s := NewStore(mockFS{}).(*store)
	require.Empty(t, s.statusSubs, "before sub")
	ch := s.SubAllStatusUpdates(0)
	require.Len(t, s.statusSubs, 1, "after sub")
	assert.True(t, s.statusSubs[0] == ch, "after sub")
	require.True(t, s.UnsubAllStatusUpdates(ch), "unsub success")
	assert.Empty(t, s.statusSubs, "after unsub")
}

func TestStore_UnsubStatusUpdatesMiddle(t *testing.T) {
	s := NewStore(mockFS{}).(*store)
	require.Empty(t, s.statusSubs, "before sub")
	const buffer = 0
	chs := []<-chan StatusUpdate{
		s.SubAllStatusUpdates(buffer),
		s.SubAllStatusUpdates(buffer),
		s.SubAllStatusUpdates(buffer),
		s.SubAllStatusUpdates(buffer),
		s.SubAllStatusUpdates(buffer),
	}
	require.Len(t, s.statusSubs, 5, "after sub")
	require.True(t, s.UnsubAllStatusUpdates(chs[2]), "unsub success")
	require.Len(t, s.statusSubs, 4, "after unsub")
	want := []<-chan StatusUpdate{
		chs[0], chs[1], chs[3], chs[4],
	}
	for i, ch := range want {
		assert.Truef(t, ch == s.statusSubs[i], "index %d, %v != %v", i, ch, s.statusSubs[i])
	}
}

func TestStore_PubSubStatusUpdates(t *testing.T) {
	s := NewStore(mockFS{
		openRead: func(name string) (io.ReadCloser, error) {
			return nil, fs.ErrNotExist
		},
		openWrite: func(name string) (io.WriteCloser, error) {
			return nopWriteCloser{}, nil
		},
	})
	const buffer = 1
	const stepID uint64 = 1
	ch := s.SubAllStatusUpdates(buffer)
	require.NotNil(t, ch, "channel")
	err := s.AddStatusUpdate(stepID, sampleTime, worker.StatusCancelled)
	require.NoError(t, err)

	select {
	case got, ok := <-ch:
		require.True(t, ok, "received on channel")
		want := StatusUpdate{
			StepID:    stepID,
			UpdateID:  1,
			Status:    worker.StatusCancelled.String(),
			Timestamp: sampleTime,
		}
		assert.Equal(t, want, got)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}
