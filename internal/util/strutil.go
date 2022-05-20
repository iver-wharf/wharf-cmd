package util

import (
	"fmt"
	"strconv"
)

// Stringify converts a value to a string. This functions differently than
// the stdlib fmt.Sprint but having some custom settings for some types:
//
// - Nils are formatted as empty string, instead of "<nil>"
//
// - Floats are always formatted in decimal notation
//
// - Other types still use fmt.Sprint
func Stringify(val any) string {
	switch val := val.(type) {
	case string:
		return val

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
