package resultstore

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"
)

var newLineBytes = []byte{'\n'}

type logLineWriteCloser struct {
	writeCloser io.WriteCloser
}

func (s *store) OpenLogFile(stepID uint64) (LogLineWriteCloser, error) {
	file, err := s.fs.OpenAppend(s.resolveLogPath(stepID))
	if err != nil {
		return nil, err
	}
	return logLineWriteCloser{
		writeCloser: file,
	}, nil
}

func (w logLineWriteCloser) WriteLogLine(line string) error {
	if _, err := w.writeCloser.Write([]byte(sanitizeLogLine(line))); err != nil {
		return err
	}
	if _, err := w.writeCloser.Write(newLineBytes); err != nil {
		return err
	}
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

func (s *store) resolveLogPath(stepID uint64) string {
	return fmt.Sprintf("steps/%d/logs.log", stepID)
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
