package rpctests

import (
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/resultstore"
	"github.com/iver-wharf/wharf-cmd/pkg/worker"
)

type mockStore struct{}

// OpenLogWriter opens a file handle abstraction for writing log lines. Logs
// will be automatically parsed when written and published to any active
// subscriptions.
func (s *mockStore) OpenLogWriter(_ uint64) (resultstore.LogLineWriteCloser, error) {
	return nil, nil
}

// OpenLogReader opens a file handle abstraction for reading log lines. Logs
// will be automatically parsed when read. Will return fs.ErrNotExist if
// the log file does not exist yet.
func (s *mockStore) OpenLogReader(_ uint64) (resultstore.LogLineReadCloser, error) {
	return nil, nil
}

// SubAllLogLines creates a new channel that streams all log lines
// from this result store since the beginning, and keeps on streaming new
// updates until unsubscribed.
func (s *mockStore) SubAllLogLines(buffer int) (<-chan resultstore.LogLine, error) {
	ch := make(chan resultstore.LogLine, buffer)

	go func() {
		for i := 1; i <= 10; i++ {
			ch <- resultstore.LogLine{
				StepID:  uint64(i),
				LogID:   uint64(i),
				Message: "-",
			}
		}

		close(ch)
	}()

	return ch, nil
}

// UnsubAllLogLines unsubscribes a subscription of all status updates
// created via SubAllLogLines.
func (s *mockStore) UnsubAllLogLines(_ <-chan resultstore.LogLine) bool {
	return true
}

// AddStatusUpdate adds a status update to a step. If the latest status
// update found for the step is the same as the new status, then this
// status update is skipped. Any written status update is also published to
// any active subscriptions.
func (s *mockStore) AddStatusUpdate(_ uint64, _ time.Time, _ worker.Status) error {
	return nil
}

// SubAllStatusUpdates creates a new channel that streams all status updates
// from this result store since the beginning, and keeps on streaming new
// updates until unsubscribed.
func (s *mockStore) SubAllStatusUpdates(buffer int) (<-chan resultstore.StatusUpdate, error) {
	ch := make(chan resultstore.StatusUpdate, buffer)

	statuses := []worker.Status{
		worker.StatusUnknown,
		worker.StatusNone,
		worker.StatusScheduling,
		worker.StatusInitializing,
		worker.StatusRunning,
		worker.StatusSuccess,
		worker.StatusFailed,
		worker.StatusCancelled,
	}

	go func() {
		for i := 1; i <= len(statuses); i++ {
			ch <- resultstore.StatusUpdate{
				StepID:   uint64(i),
				UpdateID: uint64(i),
				Status:   statuses[(i-1)%len(statuses)],
			}
		}

		close(ch)
	}()

	return ch, nil
}

// UnsubAllStatusUpdates unsubscribes a subscription of all status updates
// created via SubAllStatusUpdates.
func (s *mockStore) UnsubAllStatusUpdates(_ <-chan resultstore.StatusUpdate) bool {
	return true
}
