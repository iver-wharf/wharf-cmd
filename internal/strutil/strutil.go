package strutil

import (
	"fmt"
	"strconv"
)

// Stringify converts a value to a more expected result.
func Stringify(val any) string {
	switch val := val.(type) {
	case string:
		return val

	// fmt.Sprint returns "<nil>"; we don't want that
	case nil:
		return ""

	// Enforce decimal notation
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)

	default:
		return fmt.Sprint(val)
	}
}
