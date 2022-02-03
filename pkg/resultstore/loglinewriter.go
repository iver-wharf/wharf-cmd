package resultstore

import (
	"errors"
	"io"
	"sync/atomic"
)

// Errors specific to the LogLineWriteCloser
var (
	ErrLogWriterAlreadyOpen = errors.New("log write handle is already open for this file")
)

func (s *store) OpenLogWriter(stepID uint64) (LogLineWriteCloser, error) {
	_, alreadyOpen := s.logFilesOpened.LoadOrStore(stepID, true)
	if alreadyOpen {
		return nil, ErrLogWriterAlreadyOpen
	}
	// TODO: Read log file to see what logID should be set to
	file, err := s.fs.OpenAppend(s.resolveLogPath(stepID))
	if err != nil {
		return nil, err
	}
	return &logLineWriteCloser{
		stepID:      stepID,
		store:       s,
		writeCloser: file,
	}, nil
}

type logLineWriteCloser struct {
	stepID      uint64
	logID       uint64
	store       *store
	writeCloser io.WriteCloser
}

func (w *logLineWriteCloser) WriteLogLine(line string) error {
	sanitized := sanitizeLogLine(line)
	if _, err := w.writeCloser.Write([]byte(sanitized)); err != nil {
		return err
	}
	if _, err := w.writeCloser.Write(newLineBytes); err != nil {
		return err
	}
	tim, msg := parseLogLine(sanitized)
	w.store.pubLogLine(LogLine{
		StepID:    w.stepID,
		LogID:     atomic.AddUint64(&w.logID, 1),
		Line:      msg,
		Timestamp: tim,
	})
	return nil
}

func (w *logLineWriteCloser) Close() error {
	w.store.logFilesOpened.Delete(w.stepID)
	return w.writeCloser.Close()
}
