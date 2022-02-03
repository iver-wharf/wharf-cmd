package resultstore

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

var newLineBytes = []byte{'\n'}

func (s *store) SubAllLogLines(buffer int) (<-chan LogLine, error) {
	s.logSubMutex.Lock()
	defer s.logSubMutex.Unlock()
	readers, err := s.openAllLogReadersForCatchingUp()
	if err != nil {
		return nil, fmt.Errorf("open all log file handles: %w", err)
	}
	ch := make(chan LogLine, buffer)
	s.logSubs = append(s.logSubs, ch)
	go s.pubAllLogsToChanToCatchUp(readers, ch)
	return ch, nil
}

func (s *store) openAllLogReadersForCatchingUp() ([]LogLineReadCloser, error) {
	stepIDs, err := s.listAllStepIDs()
	if err != nil {
		return nil, fmt.Errorf("list all steps: %w", err)
	}
	readers := make([]LogLineReadCloser, len(stepIDs))
	for i, stepID := range stepIDs {
		r, err := s.OpenLogReader(stepID)
		w, ok := s.getLogWriter(stepID)
		if ok {
			// Any additional logs that come in before this catching-up is done
			// will get published via writers.
			r.SetMaxLogID(w.lastLogID)
		}
		if err != nil {
			return nil, fmt.Errorf("read logs for step %d: %w", stepID, err)
		}
		readers[i] = r
	}
	return readers, nil
}

func (s *store) pubAllLogsToChanToCatchUp(readers []LogLineReadCloser, ch chan<- LogLine) {
	for _, r := range readers {
		go s.pubLogsToChanToCatchUp(r, ch)
	}
}

func (s *store) pubLogsToChanToCatchUp(r LogLineReadCloser, ch chan<- LogLine) error {
	defer r.Close()
	for {
		line, err := r.ReadLogLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		ch <- line
	}
}

func (s *store) UnsubAllLogLines(logLineCh <-chan LogLine) bool {
	s.logSubMutex.Lock()
	defer s.logSubMutex.Unlock()
	for i, ch := range s.logSubs {
		if ch == logLineCh {
			if i != len(s.logSubs)-1 {
				copy(s.logSubs[i:], s.logSubs[i+1:])
			}
			s.logSubs = s.logSubs[:len(s.logSubs)-1]
			close(ch)
			return true
		}
	}
	return false
}

func (s *store) resolveLogPath(stepID uint64) string {
	return filepath.Join("steps", fmt.Sprint(stepID), "logs.log")
}

func (s *store) pubLogLine(logLine LogLine) {
	s.logSubMutex.RLock()
	for _, ch := range s.logSubs {
		ch <- logLine
	}
	s.logSubMutex.RUnlock() // not deferring as it's performance critical
}

func parseLogLine(line string) (time.Time, string) {
	index := strings.IndexByte(line, ' ')
	if index == -1 {
		return time.Time{}, line
	}
	timeStr := line[:index]
	t, err := time.Parse(time.RFC3339Nano, timeStr)
	if err != nil {
		return time.Time{}, line
	}
	message := line[index+1:]
	return t, message
}

var logLineReplacer = strings.NewReplacer(
	"\n", `\n`,
	"\r", `\r`,
)

func sanitizeLogLine(line string) string {
	return logLineReplacer.Replace(line)
}
