package flagtypes

import (
	"errors"
	"strings"

	"github.com/iver-wharf/wharf-core/pkg/logger"
	"github.com/spf13/cobra"
)

// LogLevel is a pflag.Value-compatible logging level flag for wharf-core's
// logger.Level type.
type LogLevel logger.Level

// Level is a utility function to return the wharf-core logger.Level value.
func (l LogLevel) Level() logger.Level {
	return logger.Level(l)
}

// Set implements the pflag.Value and fmt.Stringer interfaces.
// This returns a human-readable representation of the loglevel.
func (l *LogLevel) String() string {
	switch l.Level() {
	case logger.LevelDebug:
		return `5, "debug"`
	case logger.LevelInfo:
		return `4, "info"`
	case logger.LevelWarn:
		return `3, "warn"`
	case logger.LevelError:
		return `2, "error"`
	case logger.LevelPanic:
		return `1, "panic"`
	default:
		return logger.Level(*l).String()
	}
}

// String implements the pflag.Value interface.
// This parses the loglevel string and updates the loglevel variable.
func (l *LogLevel) Set(val string) error {
	newLevel, err := parseLevel(val)
	if err != nil {
		return err
	}
	*l = LogLevel(newLevel)
	return nil
}

func parseLevel(lvl string) (logger.Level, error) {
	// Contains more than the completions, to be more user friendly
	switch strings.ToLower(lvl) {
	case "5", "d", "debug", "debugging":
		return logger.LevelDebug, nil
	case "4", "i", "info", "information":
		return logger.LevelInfo, nil
	case "3", "w", "warn", "warning", "warnings":
		return logger.LevelWarn, nil
	case "2", "e", "error", "errors":
		return logger.LevelError, nil
	case "1", "p", "panic", "panics":
		return logger.LevelPanic, nil
	default:
		// Errors shouldn't have mutliple lines, but as this is solely for
		// pflag.Value usage then this is an exception.
		return logger.LevelDebug, errors.New(`invalid logging level, possible values:
	5  d  debug  debugging
	4  i  info   information
	3  w  warn   warning      warnings
	2  e  error  errors
	1  p  panic  panics`)
	}
}

// Type implements the pflag.Value interface.
// The value is only used in help text.
func (l *LogLevel) Type() string {
	return "loglevel"
}

// CompleteLogLevel returns completions for the LogLevel type.
func CompleteLogLevel(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	// Contains less than actually possible, to not bloat the completions
	return []string{
		"5\tIncludes all logs",
		"d\tIncludes all logs",
		"debug\tIncludes all logs",
		"4\tIncludes INFO, WARN, ERROR, and PANIC logs (default)",
		"i\tIncludes INFO, WARN, ERROR, and PANIC logs (default)",
		"info\tIncludes INFO, WARN, ERROR, and PANIC logs (default)",
		"3\tIncludes WARN, ERROR, and PANIC logs",
		"w\tIncludes WARN, ERROR, and PANIC logs",
		"warn\tIncludes WARN, ERROR, and PANIC logs",
		"2\tIncludes ERROR, and PANIC logs",
		"e\tIncludes ERROR, and PANIC logs",
		"error\tIncludes ERROR, and PANIC logs",
		"1\tSilent, except for PANIC logs",
		"p\tSilent, except for PANIC logs",
		"panic\tSilent, except for PANIC logs",
	}, cobra.ShellCompDirectiveNoFileComp
}
