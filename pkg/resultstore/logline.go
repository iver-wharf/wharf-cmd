package resultstore

import (
	"bufio"
	"strings"
	"time"
)

var newLineBytes = []byte{'\n'}

func (s *store) AddLogLine(stepID uint64, line string) error {
	file, err := s.fs.OpenAppend(s.resolveLogPath(stepID))
	if err != nil {
		return err
	}
	if _, err := file.Write([]byte(sanitizeLogLine(line))); err != nil {
		return err
	}
	if _, err := file.Write(newLineBytes); err != nil {
		return err
	}
	return nil
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
