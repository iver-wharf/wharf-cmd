package pathutil

import (
	"os"
	"strings"
)

// ShorthandHome returns the same path but replaces the first part of the
// path with the home shorthand ("~") if the path is inside the home directory.
//
// Example:
//
//   Input:  "/home/jane/Downloads"
//   Output: "~/Downloads"
func ShorthandHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return useShorthandHomePrefix(path, home)
}

func useShorthandHomePrefix(path, home string) string {
	if !strings.HasPrefix(path, home) {
		return path
	}
	return "~" + strings.TrimPrefix(path, home)
}
