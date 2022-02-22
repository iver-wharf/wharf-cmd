package varsub

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrRecursiveLoop = errors.New("recursive variable loop")
)

type VarMatch struct {
	Name       string
	FullMatch  string
	StartIndex int
	EndIndex   int
}

var varSyntaxPattern = regexp.MustCompile(`\${\s*(%*[\w_\s]*%*)\s*}`)
var paramNamePattern = regexp.MustCompile(`\s*(\w*[\s_]*\w+)\s*`)
var escapedParamPattern = regexp.MustCompile(`%(\s*[\w_\s]*\s*)%`)

func Substitute(source string, params map[string]interface{}) (interface{}, error) {
	return substituteRec(source, params, nil)
}

func substituteRec(source string, params map[string]interface{}, usedParams []string) (interface{}, error) {
	result := source
	matches := Matches(source)
	for _, match := range matches {
		var anyValue interface{}
		if match.Name == "%" {
			anyValue = "${}"
		} else if escapedParamPattern.MatchString(match.Name) {
			anyValue = escapedParamPattern.ReplaceAllString(match.Name, "${$1}")
		} else {
			if containsString(usedParams, match.Name) {
				return nil, ErrRecursiveLoop
			}
			v, ok := params[match.Name]
			if !ok {
				continue
			}
			anyValue = v
			if str, ok := anyValue.(string); ok && strings.Contains(str, "${") {
				var err error
				anyValue, err = substituteRec(str, params, append(usedParams, match.Name))
				if err != nil {
					return nil, err
				}
			}
		}
		if len(matches) == 1 && len(source) == len(match.FullMatch) {
			// keep the value as-is if it matches the whole source
			return anyValue, nil
		}
		strValue := stringify(anyValue)
		result = strings.Replace(result, match.FullMatch, strValue, 1)
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

func Matches(source string) []VarMatch {
	matches := varSyntaxPattern.FindAllStringSubmatchIndex(source, -1)
	var params []VarMatch

	for _, match := range matches {
		start, end := match[2], match[3]
		paramName := source[start:end]

		if paramName == "" {
			continue
		}

		if paramName[0] != '%' {
			paramName = paramNamePattern.ReplaceAllString(paramName, "$1")
		}

		params = append(params, VarMatch{
			Name:       paramName,
			FullMatch:  source[match[0]:match[1]],
			StartIndex: start,
			EndIndex:   end,
		})
	}

	return params
}
