package resultstore

import (
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/worker"
)

// LogLine is a single log line with metadata about its timestamp, ID, and what
// step it belongs to.
type LogLine struct {
	StepID    uint64
	LogID     uint64
	Line      string
	Timestamp time.Time
}

// StatusList is a list of status updates. This is the data structure that is
// serialized in the status update list file for a given step.
type StatusList struct {
	LastID        uint64         `json:"lastId"`
	StatusUpdates []StatusUpdate `json:"statusUpdates"`
}

// StatusUpdate is an update to a status of a build step.
type StatusUpdate struct {
	StepID    uint64    `json:"stepId"`
	UpdateID  uint64    `json:"updateId"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"` // TODO: use worker.Status here
}

// Store is the interface for storing build results and accessing them as they
// are created.
type Store interface {
	// OpenLogWriter opens a file handle abstraction for writing log lines. Logs
	// will be automatically parsed when written and published to any active
	// subscriptions.
	OpenLogWriter(stepID uint64) (LogLineWriteCloser, error)

	// SubAllLogLines creates a new channel that streams all log lines
	// from this restult store since the beginning, and keeps on streaming new
	// updates until unsubscribed.
	SubAllLogLines(buffer int) <-chan LogLine

	// UnsubAllLogLines unsubscribes a subscription of all status updates
	// created via SubAllLogLines.
	UnsubAllLogLines(ch <-chan LogLine) bool

	// AddStatusUpdate adds a status update to a step. If the latest status
	// update found for the step is the same as the new status, then this
	// status update is skipped. Any written status update is also published to
	// any active subscriptions.
	AddStatusUpdate(stepID uint64, timestamp time.Time, newStatus worker.Status) error

	// SubAllStatusUpdates creates a new channel that streams all status updates
	// from this restult store since the beginning, and keeps on streaming new
	// updates until unsubscribed.
	SubAllStatusUpdates(buffer int) (<-chan StatusUpdate, error)

	// UnsubAllStatusUpdates unsubscribes a subscription of all status updates
	// created via SubAllStatusUpdates.
	UnsubAllStatusUpdates(ch <-chan StatusUpdate) bool
}

// LogLineWriteCloser is the interface for writing log lines and ability to
// close the file handle.
type LogLineWriteCloser interface {
	io.Closer
	WriteLogLine(line string) error
}

// NewStore creates a new store using a given filesystem.
func NewStore(fs FS) Store {
	return &store{
		fs: fs,
	}
}

type store struct {
	fs FS

	statusSubMutex sync.RWMutex
	statusMutex    keyedMutex
	statusSubs     []chan StatusUpdate

	logSubMutex sync.RWMutex
	logSubs     []chan LogLine
}

func (s *store) listAllStepIDs() ([]uint64, error) {
	entries, err := s.fs.ListDirEntries("steps")
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id, err := strconv.ParseUint(e.Name(), 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}
