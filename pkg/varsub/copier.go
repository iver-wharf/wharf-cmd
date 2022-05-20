package varsub

import (
	"bufio"
	"fmt"
	"io"

	"github.com/iver-wharf/wharf-cmd/internal/filecopy"
	"github.com/iver-wharf/wharf-cmd/internal/strutil"
)

// NewCopier creates a copier that applies variable substitution on each line.
func NewCopier(source Source) filecopy.Copier {
	return &copier{source: source}
}

type copier struct {
	source Source
}

func (c *copier) Copy(dst io.Writer, src io.Reader) error {
	scanner := bufio.NewScanner(src)
	var line int
	for scanner.Scan() {
		line++
		v, err := Substitute(scanner.Text(), c.source)
		if err != nil {
			return fmt.Errorf("line %d: %w", line, err)
		}
		if _, err := fmt.Fprintln(dst, strutil.Stringify(v)); err != nil {
			return err
		}
	}
	return scanner.Err()
}
