package resultstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"testing"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/worker/workermodel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_ReadStatusUpdatesFile(t *testing.T) {
	buf := bytes.NewBufferString(`
{
	"statusUpdates": [
		{
			"updateId": 1,
			"timestamp": "2021-05-15T09:01:15.0000Z",
			"status": "Scheduling"
		},
		{
			"updateId": 2,
			"timestamp": "2021-05-15T09:01:15.0000Z",
			"status": "Running"
		},
		{
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
				Status:    workermodel.StatusScheduling,
			},
			{
				StepID:    stepID,
				UpdateID:  2,
				Timestamp: wantTime,
				Status:    workermodel.StatusRunning,
			},
			{
				StepID:    stepID,
				UpdateID:  3,
				Timestamp: wantTime,
				Status:    workermodel.StatusFailed,
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
				Status:    workermodel.StatusScheduling,
			},
			{
				StepID:    stepID,
				UpdateID:  2,
				Timestamp: sampleTime,
				Status:    workermodel.StatusRunning,
			},
			{
				StepID:    stepID,
				UpdateID:  3,
				Timestamp: sampleTime,
				Status:    workermodel.StatusFailed,
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
			"updateId": 1,
			"timestamp": "%[1]s",
			"status": "Scheduling"
		},
		{
			"updateId": 2,
			"timestamp": "%[1]s",
			"status": "Running"
		},
		{
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
	err := s.AddStatusUpdate(stepID, sampleTime, workermodel.StatusCancelled)
	require.NoError(t, err)
	want := fmt.Sprintf(`
{
	"lastId": 1,
	"statusUpdates": [
		{
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
	err := s.AddStatusUpdate(stepID, sampleTime, workermodel.StatusCancelled)
	require.NoError(t, err)
	want := fmt.Sprintf(`
{
	"lastId": 6,
	"statusUpdates": [
		{
			"updateId": 1,
			"timestamp": "%[1]s",
			"status": "Scheduling"
		},
		{
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
	err := s.AddStatusUpdate(1, time.Now(), workermodel.StatusCancelled)
	require.NoError(t, err)
}

func TestStore_SubStatusUpdatesSendsAllOldStatuses(t *testing.T) {
	updates1 := []StatusUpdate{
		{StepID: 1, UpdateID: 1, Status: workermodel.StatusCancelled},
	}
	updates2 := []StatusUpdate{
		{StepID: 2, UpdateID: 1, Status: workermodel.StatusRunning},
		{StepID: 2, UpdateID: 2, Status: workermodel.StatusSuccess},
	}
	oldLists := map[string]StatusList{
		filepath.Join(dirNameSteps, "1", fileNameStatusUpdates): {
			StatusUpdates: updates1,
		},
		filepath.Join(dirNameSteps, "2", fileNameStatusUpdates): {
			StatusUpdates: updates2,
		},
	}
	s := NewStore(mockFS{
		listDirEntries: func(name string) ([]fs.DirEntry, error) {
			if name != dirNameSteps {
				return nil, errors.New("wrong dir")
			}
			return []fs.DirEntry{
				newMockDirEntryDir("1"),
				newMockDirEntryDir("2"),
			}, nil
		},
		openRead: func(name string) (io.ReadCloser, error) {
			list, ok := oldLists[name]
			if !ok {
				return nil, fs.ErrNotExist
			}
			b, err := json.Marshal(list)
			if err != nil {
				return nil, err
			}
			return io.NopCloser(bytes.NewReader(b)), nil
		},
	}).(*store)

	buffer := len(updates1) + len(updates2)
	ch := subStatusUpdatesNoErr(t, s, buffer)
	var got []StatusUpdate
	for i := 0; i < 3; i++ {
		select {
		case gotUpdate := <-ch:
			got = append(got, gotUpdate)
		case <-time.After(time.Second):
			t.Fatalf("timed out, did not get enough results, only got: %d", len(got))
		}
	}
	want := append(updates1, updates2...)
	assert.ElementsMatch(t, want, got)
}

func TestStore_SubUnsubStatusUpdates(t *testing.T) {
	s := NewStore(mockFS{
		listDirEntries: func(string) ([]fs.DirEntry, error) {
			return nil, nil
		},
	}).(*store)
	require.Empty(t, s.statusSubs, "before sub")
	const buffer = 0
	ch := subStatusUpdatesNoErr(t, s, buffer)
	require.Len(t, s.statusSubs, 1, "after sub")
	assert.True(t, s.statusSubs[0] == ch, "after sub")
	require.True(t, s.UnsubAllStatusUpdates(ch), "unsub success")
	assert.Empty(t, s.statusSubs, "after unsub")
}

func TestStore_UnsubStatusUpdatesMiddle(t *testing.T) {
	s := NewStore(mockFS{
		listDirEntries: func(string) ([]fs.DirEntry, error) {
			return nil, nil
		},
	}).(*store)
	require.Empty(t, s.statusSubs, "before sub")
	const buffer = 0
	chs := []<-chan StatusUpdate{
		subStatusUpdatesNoErr(t, s, buffer),
		subStatusUpdatesNoErr(t, s, buffer),
		subStatusUpdatesNoErr(t, s, buffer),
		subStatusUpdatesNoErr(t, s, buffer),
		subStatusUpdatesNoErr(t, s, buffer),
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
		listDirEntries: func(string) ([]fs.DirEntry, error) {
			return nil, nil
		},
	})
	const buffer = 1
	const stepID uint64 = 1
	ch := subStatusUpdatesNoErr(t, s, buffer)
	err := s.AddStatusUpdate(stepID, sampleTime, workermodel.StatusCancelled)
	require.NoError(t, err)

	select {
	case got, ok := <-ch:
		require.True(t, ok, "received on channel")
		want := StatusUpdate{
			StepID:    stepID,
			UpdateID:  1,
			Status:    workermodel.StatusCancelled,
			Timestamp: sampleTime,
		}
		assert.Equal(t, want, got)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func subStatusUpdatesNoErr(t *testing.T, s Store, buffer int) <-chan StatusUpdate {
	ch, err := s.SubAllStatusUpdates(buffer)
	require.NoError(t, err, "sub status updates: err")
	require.NotNil(t, ch, "sub status updates: chan")
	return ch
}
