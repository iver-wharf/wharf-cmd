package lastbuild

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rogpeppe/go-internal/lockedfile"
)

// GuessNext returns the approximated next build ID to use. It calculates this
// by reading the build ID file, bumping it by +1, then releasing the file lock.
//
// This function does not mutate the next value.
//
// This means the guess from this function is not guaranteed to be the same
// value as returned by Next, so the return from this function should only be
// used to give approximations to the user.
func GuessNext() (uint, error) {
	file, err := editFile()
	if err != nil {
		return 0, err
	}
	defer file.Close()

	last, err := readLast(file)
	if err != nil {
		return 0, err
	}
	next := last + 1

	return next, nil
}

// Next returns the next build ID to use. It calculates this by locking the
// build ID file, reading its current value, bumping it by +1, then writing
// the new value, and then releasing the file lock.
func Next() (uint, error) {
	file, err := editFile()
	if err != nil {
		return 0, err
	}
	defer file.Close()

	last, err := readLast(file)
	if err != nil {
		return 0, err
	}
	next := last + 1

	file.Seek(0, io.SeekStart)
	err = writeNext(file, next)
	if err != nil {
		return 0, err
	}

	return next, nil
}

func editFile() (*lockedfile.File, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	file, err := lockedfile.Edit(path)
	if os.IsNotExist(err) {
		return createFile(path)
	}
	if err != nil {
		return nil, err
	}
	return file, nil
}

func createFile(path string) (*lockedfile.File, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0775); err != nil {
		return nil, err
	}
	return lockedfile.Create(path)
}

func readLast(reader io.Reader) (uint, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	s := strings.TrimSpace(scanner.Text())
	if s == "" {
		return 0, nil
	}
	i64, err := strconv.ParseUint(s, 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(i64), nil
}

func writeNext(writer io.Writer, next uint) error {
	_, err := fmt.Fprintln(writer, next)
	return err
}

// Path returns the path to the file that contains the last build ID.
func Path() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", nil
	}
	return filepath.Join(cacheDir, "iver-wharf", "wharf-cmd", "last-build-id.txt"), nil
}
