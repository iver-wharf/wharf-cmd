package varsub

import (
	"bufio"
	"io"
)

// NewReader wraps an io.Reader that performs variable substitution.
func NewReader(source Source, r io.Reader) io.Reader {
	return &reader{
		source:  source,
		scanner: bufio.NewScanner(r),
	}
}

type reader struct {
	source   Source
	scanner  *bufio.Scanner
	prevScan []byte
}

func (r *reader) Read(p []byte) (int, error) {
	if len(r.prevScan) > 0 {
		n := copy(p, r.prevScan)
		r.prevScan = r.prevScan[n:]
		return n, nil
	}
	if !r.scanner.Scan() {
		return 0, io.EOF
	}
	if err := r.scanner.Err(); err != nil {
		return 0, err
	}
	v, err := Substitute(r.scanner.Text(), r.source)
	if err != nil {
		return 0, err
	}
	// Will add LF at EOF even if file didn't end with LF (as bufio.Scanner
	// stops at either LF or EOF), but that's fine.
	r.prevScan = append([]byte(stringify(v)), '\n')
	n := copy(p, r.prevScan)
	r.prevScan = r.prevScan[n:]
	return n, nil
}
