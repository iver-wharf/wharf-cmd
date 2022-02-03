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
		if err := r.scanner.Err(); err != nil {
			return LogLine{}, err
		}
		return LogLine{}, io.EOF
	}
	r.logID++
	return r.parseLogLine(r.scanner.Text()), nil
}

func (r *logLineReadCloser) parseLogLine(text string) LogLine {
	tim, msg := parseLogLine(text)
	return LogLine{
		StepID:    r.stepID,
		LogID:     r.logID,
		Line:      msg,
		Timestamp: tim,
	}
}

func (r *logLineReadCloser) Close() error {
	return r.closer.Close()
}

func (r *logLineReadCloser) ReadLastLogLine() (LogLine, error) {
	var any bool
	var lastLine string
	for r.scanner.Scan() {
		any = true
		lastLine = r.scanner.Text()
		r.logID++
	}
	if err := r.scanner.Err(); err != nil {
		return LogLine{}, err
	}
	if !any {
		return LogLine{}, io.EOF
	}
	return r.parseLogLine(lastLine), nil
}
