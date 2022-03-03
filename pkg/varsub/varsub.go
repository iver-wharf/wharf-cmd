package varsub

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
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
}

var varSyntaxPattern = regexp.MustCompile(`\${\s*(%*[\w_\s]*%*)\s*}`)
var paramNamePattern = regexp.MustCompile(`\s*(\w*[\s_]*\w+)\s*`)
var escapedParamPattern = regexp.MustCompile(`%(\s*[\w_\s]*\s*)%`)

// Substitute will replace all variables in the string using the variable
// substution source. Variables are looked up recursively.
func Substitute(value string, source Source) (interface{}, error) {
	return substituteRec(value, source, nil)
}

func substituteRec(value string, source Source, usedParams []string) (interface{}, error) {
	result := value
	matches := Matches(value)
	for _, match := range matches {
		var matchVal interface{}
		if match.Name == "%" {
			matchVal = "${}"
		} else if escapedParamPattern.MatchString(match.Name) {
			matchVal = escapedParamPattern.ReplaceAllString(match.Name, "${$1}")
		} else {
			if containsString(usedParams, match.Name) {
				return nil, ErrRecursiveLoop
			}
			v, ok := source.Lookup(match.Name)
			if !ok {
				continue
			}
			matchVal = v
			if str, ok := matchVal.(string); ok && strings.Contains(str, "${") {
				var err error
				matchVal, err = substituteRec(str, source, append(usedParams, match.Name))
				if err != nil {
					return nil, err
				}
			}
		}
		if len(matches) == 1 && len(value) == len(match.FullMatch) {
			// keep the value as-is if it matches the whole value
			return matchVal, nil
		}
		matchValStr := stringify(matchVal)
		result = strings.Replace(result, match.FullMatch, matchValStr, 1)
	}
	return result, nil
}

func containsString(slice []string, element string) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

func stringify(val interface{}) string {
	switch val := val.(type) {
	case string:
		return val
	case nil: // fmt.Sprint returns "<nil>"; we don't want that
		return ""
	default:
		return fmt.Sprint(val)
	}
}

// Matches returns all variable substitution-prone matches from a string.
func Matches(value string) []VarMatch {
	matches := varSyntaxPattern.FindAllStringSubmatch(value, -1)
	var params []VarMatch

	for _, match := range matches {
		paramName := match[1]

		if paramName == "" {
			continue
		}

		if paramName[0] != '%' {
			paramName = paramNamePattern.ReplaceAllString(paramName, "$1")
		}

		params = append(params, VarMatch{
			Name:      paramName,
			FullMatch: match[0],
		})
	}

	return params
}
