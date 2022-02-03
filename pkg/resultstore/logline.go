package resultstore

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

var newLineBytes = []byte{'\n'}

func (s *store) SubAllLogLines(buffer int) <-chan LogLine {
	s.logSubMutex.Lock()
	defer s.logSubMutex.Unlock()
	// TODO: Feed all existing logs into new channel
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
