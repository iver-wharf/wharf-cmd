package resultstore

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"sync/atomic"
)

// Errors specific to the LogLineWriteCloser
var (
	ErrLogWriterAlreadyOpen = errors.New("log write handle is already open for this file")
)

func (s *store) OpenLogWriter(stepID uint64) (LogLineWriteCloser, error) {
	w := &logLineWriteCloser{
		stepID: stepID,
		store:  s,
	}
	_, alreadyOpen := s.logWritersOpened.LoadOrStore(stepID, w)
	if alreadyOpen {
		return nil, ErrLogWriterAlreadyOpen
	}
	lastLogID, err := s.getLastLogLineID(stepID)
	if err != nil {
		return nil, fmt.Errorf("read log file to get last log ID: %w", err)
	}
	w.logID = lastLogID
	file, err := s.fs.OpenAppend(s.resolveLogPath(stepID))
	if err != nil {
		return nil, err
	}
	w.writeCloser = file
	return w, nil
}

func (s *store) listOpenLogWriters() []*logLineWriteCloser {
	var writers []*logLineWriteCloser
	s.logWritersOpened.Range(func(_, value interface{}) bool {
		writers = append(writers, value.(*logLineWriteCloser))
		return true
	})
	return writers
}

func (s *store) getLastLogLineID(stepID uint64) (uint64, error) {
	r, err := s.OpenLogReader(stepID)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	lastLine, err := r.ReadLastLogLine()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return 0, nil
		}
		return 0, err
	}
	return lastLine.LogID, nil
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
	w.store.logWritersOpened.Delete(w.stepID)
	return w.writeCloser.Close()
}
