package resultstore

import (
	"bufio"
	"io"
)

func (s *store) OpenLogReader(stepID uint64) (LogLineReadCloser, error) {
	file, err := s.fs.OpenRead(s.resolveLogPath(stepID))
	if err != nil {
		return nil, err
	}
	return &logLineReadCloser{
		stepID:  stepID,
		store:   s,
		closer:  file,
		scanner: bufio.NewScanner(file),
	}, nil
}

type logLineReadCloser struct {
	stepID  uint64
	logID   uint64
	store   *store
	closer  io.Closer
	scanner *bufio.Scanner
}

func (r *logLineReadCloser) ReadLogLine() (LogLine, error) {
	if !r.scanner.Scan() {
		return LogLine{}, io.EOF
	}
	tim, msg := parseLogLine(r.scanner.Text())
	r.logID++
	return LogLine{
		StepID:    r.stepID,
		LogID:     r.logID,
		Line:      msg,
		Timestamp: tim,
	}, nil
}

func (r *logLineReadCloser) Close() error {
	return r.closer.Close()
}
