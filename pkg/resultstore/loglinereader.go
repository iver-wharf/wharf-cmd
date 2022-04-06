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
	stepID    uint64
	nextLogID uint64
	store     *store
	closer    io.Closer
	scanner   *bufio.Scanner
	maxLogID  uint64
}

func (r *logLineReadCloser) ReadLogLine() (LogLine, error) {
	if !r.scan() {
		if err := r.scanner.Err(); err != nil {
			return LogLine{}, err
		}
		return LogLine{}, io.EOF
	}
	return r.parseLogLine(r.scanner.Text()), nil
}

func (r *logLineReadCloser) parseLogLine(text string) LogLine {
	tim, msg := parseLogLine(text)
	return LogLine{
		StepID:    r.stepID,
		LogID:     r.nextLogID,
		Message:   msg,
		Timestamp: tim,
	}
}

func (r *logLineReadCloser) Close() error {
	return r.closer.Close()
}

func (r *logLineReadCloser) ReadLastLogLine() (LogLine, error) {
	noLineFound := true
	var lastLine string
	for r.scan() {
		noLineFound = false
		lastLine = r.scanner.Text()
	}
	if err := r.scanner.Err(); err != nil {
		return LogLine{}, err
	}
	if noLineFound {
		return LogLine{}, io.EOF
	}
	return r.parseLogLine(lastLine), nil
}

func (r *logLineReadCloser) scan() bool {
	if r.maxLogID != 0 && r.nextLogID >= r.maxLogID {
		return false
	}
	if !r.scanner.Scan() {
		return false
	}
	r.nextLogID++
	return true
}

func (r *logLineReadCloser) SetMaxLogID(logID uint64) {
	r.maxLogID = logID
}
