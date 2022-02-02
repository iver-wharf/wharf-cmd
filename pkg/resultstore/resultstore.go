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
	UpdateID  uint64    `json:"updateId"`
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"`
}

type ArtifactListMeta struct {
	Artifacts []ArtifactMeta `json:"artifacts"`
}

type ArtifactMeta struct {
	ArtifactID uint64 `json:"artifactId"`
	Name       string `json:"name"`
	Path       string `json:"path"`
}

type Store interface {
	OpenLogFile(stepID uint64) (LogLineWriteCloser, error)

	AddStatusUpdate(stepID uint64, timestamp time.Time, newStatus worker.Status) error
	AddArtifact(stepID uint64, artifactName string) (io.WriteCloser, error)

	SubAllLogLines(buffer int) <-chan LogLine
	UnsubAllLogLines(ch <-chan LogLine) bool

	// TODO: Remove this:
	ReadAllLogLines(stepID uint64) ([]LogLine, error)

	// TODO: Add streaming handles. Callback or channels? Need smart buffering
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
	fs             FS
	lastStatusID   uint64
	lastArtifactID uint64

	statusMutex keyedMutex

	logSubMutex sync.RWMutex
	logSubs     []chan LogLine
}
