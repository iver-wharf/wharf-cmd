package resultstore

import (
	"io"
	"sync"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/worker"
)

type LogLine struct {
	StepID    uint64
	LogID     uint64
	Line      string
	Timestamp time.Time
}

type StatusList struct {
	StatusUpdates []StatusUpdate `json:"statusUpdates"`
}

type StatusUpdate struct {
	StepID    uint64    `json:"stepId"`
	UpdateID  uint64    `json:"updateId"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

// Store is the interface for storing build results and accessing them as they
// are created.
type Store interface {
	OpenLogFile(stepID uint64) (LogLineWriteCloser, error)
	SubAllLogLines(buffer int) <-chan LogLine
	UnsubAllLogLines(ch <-chan LogLine) bool

	AddStatusUpdate(stepID uint64, timestamp time.Time, newStatus worker.Status) error
	SubAllStatusUpdates(buffer int) <-chan StatusUpdate
	UnsubAllStatusUpdates(ch <-chan StatusUpdate) bool
}

type LogLineWriteCloser interface {
	io.Closer
	WriteLogLine(line string) error
}

func NewStore(fs FS) Store {
	return &store{
		fs: fs,
	}
}

type store struct {
	fs FS

	lastStatusID   uint64
	statusSubMutex sync.RWMutex
	statusMutex    keyedMutex
	statusSubs     []chan StatusUpdate

	logSubMutex sync.RWMutex
	logSubs     []chan LogLine
}
