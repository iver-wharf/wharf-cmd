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
	if s.frozen {
		return nil, ErrFrozen
	}
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
	w.lastLogID = lastLogID
	file, err := s.fs.OpenAppend(s.resolveLogPath(stepID))
	if err != nil {
		return nil, err
	}
	w.writeCloser = file
	return w, nil
}

func (s *store) getLogWriter(stepID uint64) (*logLineWriteCloser, bool) {
	val, ok := s.logWritersOpened.Load(stepID)
	if !ok {
		return nil, false
	}
	return val, true
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
	lastLogID   uint64
	store       *store
	writeCloser io.WriteCloser
}

func (w *logLineWriteCloser) WriteLogLine(line string) (LogLine, error) {
	sanitized := sanitizeLogLine(line)
	if _, err := w.writeCloser.Write([]byte(sanitized)); err != nil {
		return LogLine{}, err
	}
	if _, err := w.writeCloser.Write(newLineBytes); err != nil {
		return LogLine{}, err
	}
	logID := atomic.AddUint64(&w.lastLogID, 1)
	return w.store.parseAndPubLogLine(w.stepID, logID, sanitized), nil
}

func (w *logLineWriteCloser) Close() error {
	w.store.logWritersOpened.Delete(w.stepID)
	return w.writeCloser.Close()
}
