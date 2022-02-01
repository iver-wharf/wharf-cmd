package resultstore

import (
	"bytes"
	"io"
	"testing"
	"time"

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
	want := StatusList{
		StatusUpdates: []StatusUpdate{
			{
				UpdateID:  1,
				Timestamp: wantTime,
				Status:    "Scheduling",
			},
			{
				UpdateID:  2,
				Timestamp: wantTime,
				Status:    "Running",
			},
			{
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
	got, err := s.readStatusUpdatesFile(1)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestStore_WriteStatusUpdatesFile(t *testing.T) {
	sampleTime := time.Date(2021, 5, 15, 9, 1, 15, 0, time.UTC)
	list := StatusList{
		StatusUpdates: []StatusUpdate{
			{
				UpdateID:  1,
				Timestamp: sampleTime,
				Status:    "Scheduling",
			},
			{
				UpdateID:  2,
				Timestamp: sampleTime,
				Status:    "Running",
			},
			{
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
	err := s.writeStatusUpdatesFile(1, list)
	require.NoError(t, err)
	want := `
{
	"statusUpdates": [
		{
			"updateId": 1,
			"timestamp": "2021-05-15T09:01:15Z",
			"status": "Scheduling"
		},
		{
			"updateId": 2,
			"timestamp": "2021-05-15T09:01:15Z",
			"status": "Running"
		},
		{
			"updateId": 3,
			"timestamp": "2021-05-15T09:01:15Z",
			"status": "Failed"
		}
	]
}`
	assert.JSONEq(t, want, buf.String())
}
