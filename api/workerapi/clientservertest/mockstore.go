package main

import (
	"time"

	"github.com/iver-wharf/wharf-cmd/pkg/resultstore"
	"github.com/iver-wharf/wharf-cmd/pkg/worker"
)

type mockStore struct{}

// OpenLogWriter opens a file handle abstraction for writing log lines. Logs
// will be automatically parsed when written and published to any active
// subscriptions.
func (s *mockStore) OpenLogWriter(stepID uint64) (resultstore.LogLineWriteCloser, error) {
	return nil, nil
}

// OpenLogReader opens a file handle abstraction for reading log lines. Logs
// will be automatically parsed when read. Will return fs.ErrNotExist if
// the log file does not exist yet.
func (s *mockStore) OpenLogReader(stepID uint64) (resultstore.LogLineReadCloser, error) {
	return nil, nil
}

// SubAllLogLines creates a new channel that streams all log lines
// from this result store since the beginning, and keeps on streaming new
// updates until unsubscribed.
func (s *mockStore) SubAllLogLines(buffer int) (<-chan resultstore.LogLine, error) {
	log.Info().WithInt("buffer", buffer).Message("SubAllLogLines - mockStore")
	ch := make(chan resultstore.LogLine, buffer)

	go func() {
		for i := 1; i <= 1000; i++ {
			ch <- resultstore.LogLine{
				StepID:    uint64(i / 100),
				LogID:     uint64(i),
				Message:   "-",
				Timestamp: time.Now(),
			}
		}

		close(ch)
	}()

	return ch, nil
}

// UnsubAllLogLines unsubscribes a subscription of all status updates
// created via SubAllLogLines.
func (s *mockStore) UnsubAllLogLines(ch <-chan resultstore.LogLine) bool {
	log.Info().Message("UnsubAllLogLines - mockStore")
	return true
}

// AddStatusUpdate adds a status update to a step. If the latest status
// update found for the step is the same as the new status, then this
// status update is skipped. Any written status update is also published to
// any active subscriptions.
func (s *mockStore) AddStatusUpdate(stepID uint64, timestamp time.Time, newStatus worker.Status) error {
	return nil
}

// SubAllStatusUpdates creates a new channel that streams all status updates
// from this result store since the beginning, and keeps on streaming new
// updates until unsubscribed.
func (s *mockStore) SubAllStatusUpdates(buffer int) (<-chan resultstore.StatusUpdate, error) {
	log.Info().WithInt("buffer", buffer).Message("SubAllStatusUpdates - mockStore")
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
		statusIdx := 0
		updateID := 1
		for i := 0; i < len(statuses)*4; i++ {
			statusIdx++
			updateID++
			ch <- resultstore.StatusUpdate{
				StepID:    uint64(i / 100),
				UpdateID:  uint64(updateID),
				Status:    statuses[statusIdx%len(statuses)],
				Timestamp: time.Now(),
			}
		}

		close(ch)
	}()

	return ch, nil
}

// UnsubAllStatusUpdates unsubscribes a subscription of all status updates
// created via SubAllStatusUpdates.
func (s *mockStore) UnsubAllStatusUpdates(ch <-chan resultstore.StatusUpdate) bool {
	log.Info().Message("UnsubAllStatusUpdates - mockStore")
	return true
}
