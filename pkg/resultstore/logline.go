package resultstore

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"time"
)

var newLineBytes = []byte{'\n'}

type logLineWriteCloser struct {
	stepID      uint64
	logID       uint64
	store       *store
	writeCloser io.WriteCloser
}

func (s *store) OpenLogFile(stepID uint64) (LogLineWriteCloser, error) {
	file, err := s.fs.OpenAppend(s.resolveLogPath(stepID))
	if err != nil {
		return nil, err
	}
	return logLineWriteCloser{
		stepID:      stepID,
		store:       s,
		writeCloser: file,
	}, nil
}

func (w logLineWriteCloser) WriteLogLine(line string) error {
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

func (w logLineWriteCloser) Close() error {
	return w.writeCloser.Close()
}

func (s *store) ReadAllLogLines(stepID uint64) ([]LogLine, error) {
	file, err := s.fs.OpenRead(s.resolveLogPath(stepID))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var lines []LogLine
	var logID uint64
	for scanner.Scan() {
		tim, line := parseLogLine(scanner.Text())
		logID++
		lines = append(lines, LogLine{
			StepID:    stepID,
			LogID:     logID,
			Timestamp: tim,
			Line:      line,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func (s *store) SubAllLogLines(buffer int) <-chan LogLine {
	s.logSubMutex.Lock()
	defer s.logSubMutex.Unlock()
	ch := make(chan LogLine, buffer)
	s.logSubs = append(s.logSubs, ch)
	return ch
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
			return true
		}
	}
	return false
}

func (s *store) resolveLogPath(stepID uint64) string {
	return fmt.Sprintf("steps/%d/logs.log", stepID)
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
