package resultstore

import (
	"errors"
	"io/fs"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	sampleTimeStr = "2021-05-09T12:13:14.1234Z"
	sampleTime    = time.Date(2021, 5, 9, 12, 13, 14, 123400000, time.UTC)
)

func TestStore_ListAllStepIDs(t *testing.T) {
	testCases := []struct {
		name  string
		dirs  []string
		files []string
		want  []uint64
	}{
		{
			name: "empty dir",
			want: []uint64{},
		},
		{
			name:  "only files",
			files: []string{"1", "2", "3"},
			want:  []uint64{},
		},
		{
			name: "only non-number dirs",
			dirs: []string{"a", "b", "c"},
			want: []uint64{},
		},
		{
			name:  "mix of invalid dirs and files",
			files: []string{"a", "1", "b", "2", "c"},
			dirs:  []string{"a", "b", "c"},
			want:  []uint64{},
		},
		{
			name: "valid dirs",
			dirs: []string{"1", "2", "3"},
			want: []uint64{1, 2, 3},
		},
		{
			name: "mix of valid and invalid dirs",
			dirs: []string{"a", "1", "b", "2", "c", "3"},
			want: []uint64{1, 2, 3},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewStore(mockFS{
				listDirEntries: func(name string) ([]fs.DirEntry, error) {
					if name != dirNameSteps {
						return nil, errors.New("wrong dir")
					}
					var dirEntries []fs.DirEntry
					for _, f := range tc.files {
						dirEntries = append(dirEntries, newMockDirEntryFile(f))
					}
					for _, d := range tc.dirs {
						dirEntries = append(dirEntries, newMockDirEntryDir(d))
					}
					return dirEntries, nil
				},
			}).(*store)
			got, err := s.listAllStepIDs()
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestStore_Close(t *testing.T) {
	store := NewStore(mockFS{
		listDirEntries: func(name string) ([]fs.DirEntry, error) {
			return nil, nil
		},
	})

	err := store.Close()
	assert.NoError(t, err, "close store")

	err = store.Close()
	assert.NoError(t, err, "close store again")
}
