package utils

import (
	"bufio"
	"bytes"
	"io"

	"github.com/pborman/ansi"
)

type StreamScanner interface {
	Scan() bool
	Err() error
	Text() string
}

type streamScanner struct {
	sanitizationMethods SanitizationFlags
	scanner             *bufio.Scanner
}

func NewStreamScanner(stream io.ReadCloser, f SanitizationFlags) StreamScanner {
	scanner := bufio.NewScanner(stream)

	s := streamScanner{
		sanitizationMethods: f,
		scanner:             scanner,
	}

	scanner.Split(s.sanitizeLogLine)
	return s
}

func (s streamScanner) Scan() bool {
	return s.scanner.Scan()
}

func (s streamScanner) Err() error {
	return s.scanner.Err()
}

func (s streamScanner) Text() string {
	return s.scanner.Text()
}

func (s streamScanner) sanitizeLogLine(data []byte, atEOF bool) (advance int, token []byte, err error) {
	advance, token, err = bufio.ScanLines(data, atEOF)
	if len(token) > 0 && (err == nil || err == io.EOF) {
		if s.sanitizationMethods.HasFlag(FromLastCR) {
			token = fromLastCR(token)
		}

		if s.sanitizationMethods.HasFlag(RemoveAnsiCodes) {
			token = removeANSIcodes(token)
		}
	}
	return advance, token, err
}

func fromLastCR(data []byte) []byte {
	index := bytes.LastIndexByte(data, '\r')
	if index > 0 && index+1 <= len(data) {
		return data[index+1:]
	}
	return data
}

func removeANSIcodes(data []byte) []byte {
	stripped, err := ansi.Strip(data)
	if err != nil {
		log.Error().WithError(err).Message("Failed to strip logs from ANSI codes.")
		return data
	}

	return stripped
}
