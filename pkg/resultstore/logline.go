package resultstore

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/typ.v3/pkg/chans"
)

var (
	newLineBytes    = []byte{'\n'}
	logLineReplacer = strings.NewReplacer(
		"\n", `\n`,
		"\r", `\r`,
	)
	fileNameLogs = "logs.log"
)

func (s *store) SubAllLogLines(buffer int) (<-chan LogLine, error) {
	s.logSubMutex.Lock()
	defer s.logSubMutex.Unlock()
	readers, err := s.openAllLogReadersForCatchingUp()
	if err != nil {
		return nil, fmt.Errorf("open all log file handles: %w", err)
	}
	ch := s.logPubSub.SubBuf(buffer)
	go s.pubAllLogsToChanToCatchUp(readers, s.logPubSub.WithOnly(ch))
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

func (s *store) pubAllLogsToChanToCatchUp(readers []LogLineReadCloser, pubSub *chans.PubSub[LogLine]) {
	for _, r := range readers {
		go s.pubLogsToChanToCatchUp(r, pubSub)
	}
}

func (s *store) pubLogsToChanToCatchUp(r LogLineReadCloser, pubSub *chans.PubSub[LogLine]) error {
	defer r.Close()
	for {
		line, err := r.ReadLogLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		pubSub.PubSync(line)
	}
}

func (s *store) UnsubAllLogLines(logLineCh <-chan LogLine) error {
	return s.logPubSub.Unsub(logLineCh)
}

func (s *store) resolveLogPath(stepID uint64) string {
	return filepath.Join(dirNameSteps, fmt.Sprint(stepID), fileNameLogs)
}

func (s *store) parseAndPubLogLine(stepID uint64, logID uint64, line string) {
	tim, msg := parseLogLine(line)
	logLine := LogLine{
		StepID:    stepID,
		LogID:     logID,
		Message:   msg,
		Timestamp: tim,
	}
	// Locking to prevent new data being added during fetching existing data
	// part of when a new subscription is made.
	s.logSubMutex.RLock()
	s.logPubSub.PubSync(logLine)
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

func sanitizeLogLine(line string) string {
	return logLineReplacer.Replace(line)
}
