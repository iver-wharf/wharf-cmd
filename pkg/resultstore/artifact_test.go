package resultstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"path/filepath"
	"testing"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/worker/workermodel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_ReadArtifactEventsFile(t *testing.T) {
	buf := bytes.NewBufferString(`
{
	"artifactEvents": [
		{
			"artifactId": 1,
			"name": "artifact-1"
		},
		{
			"artifactId": 2,
			"name": "artifact-2"
		},
		{
			"artifactId": 3,
			"name": "artifact-3"
		}
	]
}
`)
	const stepID uint64 = 1
	want := ArtifactEventList{
		ArtifactEvents: []ArtifactEvent{
			{
				ArtifactID: 1,
				StepID:     stepID,
				Name:       "artifact-1",
			},
			{
				ArtifactID: 2,
				StepID:     stepID,
				Name:       "artifact-2",
			},
			{
				ArtifactID: 3,
				StepID:     stepID,
				Name:       "artifact-3",
			},
		},
	}
	s := NewStore(mockFS{
		openRead: func(name string) (io.ReadCloser, error) {
			return io.NopCloser(buf), nil
		},
	}).(*store)
	got, err := s.readArtifactEventsFile(stepID)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestStore_WriteArtifactEventsFile(t *testing.T) {
	const stepID uint64 = 1
	list := ArtifactEventList{
		LastID: 3,
		ArtifactEvents: []ArtifactEvent{
			{
				ArtifactID: 1,
				StepID:     stepID,
				Name:       "artifact-1",
			},
			{
				ArtifactID: 2,
				StepID:     stepID,
				Name:       "artifact-2",
			},
			{
				ArtifactID: 3,
				StepID:     stepID,
				Name:       "artifact-3",
			},
		},
	}
	var buf bytes.Buffer
	s := NewStore(mockFS{
		openWrite: func(name string) (io.WriteCloser, error) {
			return nopWriteCloser{&buf}, nil
		},
	}).(*store)
	err := s.writeArtifactEventsFile(stepID, list)
	require.NoError(t, err)
	want := `
{
	"lastId": 3,
	"artifactEvents": [
		{
			"artifactId": 1,
			"stepId": 1,
			"name": "artifact-1"
		},
		{
			"artifactId": 2,
			"stepId": 1,
			"name": "artifact-2"
		},
		{
			"artifactId": 3,
			"stepId": 1,
			"name": "artifact-3"
		}
	]
}`
	assert.JSONEq(t, want, buf.String())
}

func TestStore_AddArtifactEventFirst(t *testing.T) {
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
	err := s.AddArtifactEvent(stepID, workermodel.ArtifactMeta{Name: "artifact-1"})
	require.NoError(t, err)
	want := `
{
	"lastId": 1,
	"artifactEvents": [
		{
			"artifactId": 1,
			"stepId": 1,
			"name": "artifact-1"
		}
	]
}`
	assert.JSONEq(t, want, buf.String())
}

func TestStore_AddArtifactEventSecond(t *testing.T) {
	buf := bytes.NewBufferString(`{
	"lastId": 5,
	"artifactEvents": [
		{
			"artifactId": 1,
			"name": "artifact-1"
		}
	]
}`)
	s := NewStore(mockFS{
		openRead: func(name string) (io.ReadCloser, error) {
			return io.NopCloser(buf), nil
		},
		openWrite: func(name string) (io.WriteCloser, error) {
			return nopWriteCloser{buf}, nil
		},
	})
	const stepID uint64 = 1
	err := s.AddArtifactEvent(stepID, workermodel.ArtifactMeta{Name: "artifact-2"})
	require.NoError(t, err)
	want := `
{
	"lastId": 6,
	"artifactEvents": [
		{
			"artifactId": 1,
			"stepId": 1,
			"name": "artifact-1"
		},
		{
			"artifactId": 6,
			"stepId": 1,
			"name": "artifact-2"
		}
	]
}`
	assert.JSONEq(t, want, buf.String())
}

func TestStore_SubArtifactEventsSendsAllOldEvents(t *testing.T) {
	events1 := []ArtifactEvent{
		{StepID: 1, ArtifactID: 1, Name: "artifact-1"},
	}
	events2 := []ArtifactEvent{
		{StepID: 2, ArtifactID: 2, Name: "artifact-2"},
		{StepID: 2, ArtifactID: 3, Name: "artifact-3"},
	}
	oldLists := map[string]ArtifactEventList{
		filepath.Join(dirNameSteps, "1", fileNameArtifactEvents): {
			ArtifactEvents: events1,
		},
		filepath.Join(dirNameSteps, "2", fileNameArtifactEvents): {
			ArtifactEvents: events2,
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

	buffer := len(events1) + len(events2)
	ch := subArtifactEventsNoErr(t, s, buffer)
	var got []ArtifactEvent
	for i := 0; i < 3; i++ {
		select {
		case gotEvent := <-ch:
			got = append(got, gotEvent)
		case <-time.After(time.Second):
			t.Fatalf("timed out, did not get enough results, only got: %d", len(got))
		}
	}
	want := append(events1, events2...)
	assert.ElementsMatch(t, want, got)
}

func TestStore_PubSubArtifactEvents(t *testing.T) {
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
	ch := subArtifactEventsNoErr(t, s, buffer)
	err := s.AddArtifactEvent(stepID, workermodel.ArtifactMeta{Name: "artifact-1"})
	require.NoError(t, err)

	select {
	case got, ok := <-ch:
		require.True(t, ok, "received on channel")
		want := ArtifactEvent{
			ArtifactID: 1,
			StepID:     stepID,
			Name:       "artifact-1",
		}
		assert.Equal(t, want, got)
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func subArtifactEventsNoErr(t *testing.T, s Store, buffer int) <-chan ArtifactEvent {
	ch, err := s.SubAllArtifactEvents(buffer)
	require.NoError(t, err, "sub artifact events: err")
	require.NotNil(t, ch, "sub artifact events: chan")
	return ch
}
