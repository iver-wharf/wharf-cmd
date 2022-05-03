package resultstore

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/worker/workermodel"
	"gopkg.in/typ.v4/chans"
	"gopkg.in/typ.v4/sync2"
)

var (
	// ErrFrozen is returned when doing a write operation on a frozen store.
	ErrFrozen = errors.New("store frozen")
	// ErrClosed is returned when doing a read operation on a closed store.
	ErrClosed = errors.New("store closed")
)

var (
	dirNameSteps = "steps"
)

// LogLine is a single log line with metadata about its timestamp, ID, and what
// step it belongs to.
type LogLine struct {
	StepID    uint64
	LogID     uint64
	Message   string
	Timestamp time.Time
}

// GoString implements fmt.Stringer
func (log LogLine) String() string {
	return fmt.Sprintf("%s %s", log.Timestamp.Format(time.RFC3339Nano), log.Message)
}

// GoString implements fmt.GoStringer
func (log LogLine) GoString() string {
	return fmt.Sprintf("(stepID: %d, logID: %d) \"%s %s\"", log.StepID, log.LogID, log.Timestamp.Format(time.RFC3339Nano), log.Message)
}

// StatusList is a list of status updates. This is the data structure that is
// serialized in the status update list file for a given step.
type StatusList struct {
	LastID        uint64         `json:"lastId"`
	StatusUpdates []StatusUpdate `json:"statusUpdates"`
}

// StatusUpdate is an update to a status of a build step.
type StatusUpdate struct {
	StepID    uint64             `json:"-"`
	UpdateID  uint64             `json:"updateId"`
	Timestamp time.Time          `json:"timestamp"`
	Status    workermodel.Status `json:"status"`
}

// ArtifactEventList is a list of artifact events. This is the data structure
// that is serialized in the artifact event list file for a given step.
type ArtifactEventList struct {
	LastID         uint64          `json:"lastId"`
	ArtifactEvents []ArtifactEvent `json:"artifactEvents"`
}

// ArtifactEvent is metadata about an artifact created during building.
type ArtifactEvent struct {
	ArtifactID uint64 `json:"artifactId"`
	StepID     uint64 `json:"stepId"`
	Name       string `json:"name"`
}

// Store is the interface for storing build results and accessing them as they
// are created.
type Store interface {
	// OpenLogWriter opens a file handle abstraction for writing log lines. Logs
	// will be automatically parsed when written and published to any active
	// subscriptions.
	//
	// Will return ErrFrozen if the store is frozen.
	OpenLogWriter(stepID uint64) (LogLineWriteCloser, error)

	// OpenLogReader opens a file handle abstraction for reading log lines. Logs
	// will be automatically parsed when read.
	//
	// Will return fs.ErrNotExist if the log file does not exist yet.
	OpenLogReader(stepID uint64) (LogLineReadCloser, error)

	// SubAllLogLines creates a new channel that streams all log lines
	// from this result store since the beginning, and keeps on streaming new
	// updates until unsubscribed.
	SubAllLogLines(buffer int) (<-chan LogLine, error)

	// UnsubAllLogLines unsubscribes a subscription of all status updates
	// created via SubAllLogLines.
	UnsubAllLogLines(ch <-chan LogLine) error

	// AddStatusUpdate adds a status update to a step. If the latest status
	// update found for the step is the same as the new status, then this
	// status update is skipped. Any written status update is also published to
	// any active subscriptions.
	//
	// Will return ErrFrozen if the store is frozen.
	AddStatusUpdate(stepID uint64, timestamp time.Time, newStatus workermodel.Status) error

	// SubAllStatusUpdates creates a new channel that streams all status updates
	// from this result store since the beginning, and keeps on streaming new
	// updates until unsubscribed.
	SubAllStatusUpdates(buffer int) (<-chan StatusUpdate, error)

	// UnsubAllStatusUpdates unsubscribes a subscription of all status updates
	// created via SubAllStatusUpdates.
	UnsubAllStatusUpdates(ch <-chan StatusUpdate) error

	// AddArtifactEvent adds an artifact event to a step.
	// Any written artifact event is also published to any active subscriptions.
	//
	// Will return ErrFrozen if the store is frozen.
	AddArtifactEvent(stepID uint64, artifactMeta workermodel.ArtifactMeta) error

	// SubAllArtifactEvents creates a new channel that streams all artifact
	// events from this result store since the beginning, and keeps on
	// streaming new events until unsubscribed.
	SubAllArtifactEvents(buffer int) (<-chan ArtifactEvent, error)

	// UnsubAllArtifactEvents unsubscribes a subscription of all artifact
	// events created via SubAllStatusUpdates.
	UnsubAllArtifactEvents(ch <-chan ArtifactEvent) error

	// Freeze waits for all write operations to finish, closes any open writers
	// and causes future write operations to error. This cannot be undone.
	//
	// All new subscriptions will be closed after catching up.
	Freeze() error

	// Close, in addition to freezing the store by calling Freeze, closes any
	// open readers and causes future read operations to error. This cannot be
	// undone.
	Close() error
}

// LogLineWriteCloser is the interface for writing log lines and ability to
// close the file handle.
type LogLineWriteCloser interface {
	io.Closer
	// WriteLogLine will write the log line to the file and publish a parsed
	// LogLine to any active subscriptions. An error is returned if it failed
	// to write, such as if the file system has run out of disk space or if the
	// file was removed.
	WriteLogLine(line string) error
}

// LogLineReadCloser is the interface for reading log lines and ability to
// close the file handle.
type LogLineReadCloser interface {
	io.Closer
	// SetMaxLogID sets the last log ID this reader will read before artificially
	// returning io.EOF in ReadLogLine and ReadLastLogLine. Setting this to zero,
	// which is the default, will disable this functionality.
	SetMaxLogID(logID uint64)
	// ReadLogLine will read the next log line in the file, or return io.EOF
	// when the reader has reached the end of the file.
	ReadLogLine() (LogLine, error)
	// ReadLastLogLine will read the entire file and return the last log line
	// in the file. Wil return io.EOF if the file is empty.
	ReadLastLogLine() (LogLine, error)
}

// NewStore creates a new store using a given filesystem.
func NewStore(fs FS) Store {
	return &store{
		fs:               fs,
		logReadersOpened: newSyncSet[*logLineReadCloser](),
	}
}

type store struct {
	fs FS

	statusPubSub   chans.PubSub[StatusUpdate]
	statusSubMutex sync.RWMutex
	statusMutex    sync2.KeyedMutex[uint64]

	logSubMutex      sync.RWMutex
	logPubSub        chans.PubSub[LogLine]
	logWritersOpened sync2.Map[uint64, *logLineWriteCloser]
	logReadersOpened syncSet[*logLineReadCloser]

	artifactPubSub   chans.PubSub[ArtifactEvent]
	artifactSubMutex sync.RWMutex
	artifactMutex    sync2.KeyedMutex[uint64]

	frozen bool
	closed bool
}

func (s *store) Freeze() error {
	if s.frozen {
		return nil
	}

	s.frozen = true
	var closeWriterErr error
	s.logWritersOpened.Range(func(_ uint64, writer *logLineWriteCloser) bool {
		if err := writer.Close(); err != nil && closeWriterErr == nil {
			closeWriterErr = err
			return false
		}
		return true
	})
	if closeWriterErr != nil {
		return closeWriterErr
	}

	stepIDs, err := s.listAllStepIDs()
	if err != nil {
		return err
	}

	for _, stepID := range stepIDs {
		s.artifactMutex.LockKey(stepID)
		s.statusMutex.LockKey(stepID)
	}

	s.artifactPubSub.UnsubAll()
	s.logPubSub.UnsubAll()
	s.statusPubSub.UnsubAll()

	for _, stepID := range stepIDs {
		s.artifactMutex.UnlockKey(stepID)
		s.statusMutex.UnlockKey(stepID)
	}
	return nil
}

func (s *store) Close() error {
	if s.closed {
		return nil
	}

	err := s.Freeze()
	if err != nil {
		return err
	}
	s.closed = true

	var closeReaderErr error

	readers := s.logReadersOpened.Slice()
	for _, reader := range readers {
		if err := reader.Close(); err != nil && closeReaderErr == nil {
			closeReaderErr = err
		}
	}
	return err
}

func (s *store) listAllStepIDs() ([]uint64, error) {
	entries, err := s.fs.ListDirEntries(dirNameSteps)
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
