package varsub

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/typ.v4/slices"
)

var (
	// ErrRecursiveLoop means a variable substitution has some self-referring
	// variables, either directly or indirectly.
	ErrRecursiveLoop = errors.New("recursive variable loop")
)

// VarMatch is a single variable match.
type VarMatch struct {
	// Name is the name of the variable inside the match.
	Name string
	// FullMatch is the entire match, including the variable syntax of ${}.
	FullMatch string
	// IsVar is true if this was a match on a variable, or false if it was
	// just a match on the delimiting string
	IsVar bool
}

var varSyntaxPattern = regexp.MustCompile(`\${\s*([^}]*)\s*}`)

// Substitute will replace all variables in the string using the variable
// substution source. Variables are looked up recursively.
func Substitute(value string, source Source) (any, error) {
	return substituteRec(value, source, nil)
}

func substituteRec(value string, source Source, usedParams []string) (any, error) {
	matches := Split(value)
	var sb strings.Builder
	for _, match := range matches {
		if !match.IsVar {
			sb.WriteString(match.FullMatch)
			continue
		}

		if unescaped, ok := unescapeFullMatch(match.FullMatch); ok {
			sb.WriteString(unescaped)
			continue
		}

		if slices.Contains(usedParams, match.Name) {
			return nil, ErrRecursiveLoop
		}

		v, ok := source.Lookup(match.Name)
		if !ok {
			sb.WriteString(match.FullMatch)
			continue
		}

		var matchVal = v.Value
		if str, ok := matchVal.(string); ok && strings.Contains(str, "${") {
			var err error
			matchVal, err = substituteRec(str, source, append(usedParams, match.Name))
			if err != nil {
				return nil, err
			}
		}
		if len(matches) == 1 {
			// keep the value as-is if it matches the whole value
			return matchVal, nil
		}
		matchValStr := stringify(matchVal)
		sb.WriteString(matchValStr)
	}
	return sb.String(), nil
}

func unescapeFullMatch(fullMatch string) (string, bool) {
	if fullMatch == "${%}" {
		return "${}", true
	}
	if !strings.HasPrefix(fullMatch, "${%") || !strings.HasSuffix(fullMatch, "%}") {
		return fullMatch, false
	}
	s := strings.TrimPrefix(strings.TrimSuffix(fullMatch, "%}"), "${%")
	return fmt.Sprintf("${%s}", s), true
}

func stringify(val any) string {
	switch val := val.(type) {
	case string:
		return val
	case nil: // fmt.Sprint returns "<nil>"; we don't want that
		return ""
	default:
		return fmt.Sprint(val)
	}
}

// Split up a string on its variable and non-variable matches.
func Split(value string) []VarMatch {
	matches := varSyntaxPattern.FindAllStringSubmatchIndex(value, -1)
	if len(matches) == 0 {
		return []VarMatch{{Name: "", FullMatch: value, IsVar: false}}
	}
	var vars []VarMatch
	var lastEnd int
	for _, m := range matches {
		if m[0] > lastEnd {
			vars = append(vars, VarMatch{
				FullMatch: value[lastEnd:m[0]],
			})
		}
		vars = append(vars, VarMatch{
			FullMatch: value[m[0]:m[1]],
			Name:      strings.TrimSpace(value[m[2]:m[3]]),
			IsVar:     true,
		})
		lastEnd = m[1]
	}
	if len(value) > lastEnd {
		vars = append(vars, VarMatch{
			FullMatch: value[lastEnd:],
		})
	}
	return vars
}
