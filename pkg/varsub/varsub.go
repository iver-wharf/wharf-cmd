package varsub

import (
	"fmt"
	"regexp"
	"strings"
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

func Substitute(source string, params map[string]interface{}) interface{} {
	result := source
	matches := Matches(source)
	for _, match := range matches {
		var newValue interface{}
		if match.Name == "%" {
			newValue = "${}"
		} else if escapedParamPattern.MatchString(match.Name) {
			newValue = escapedParamPattern.ReplaceAllString(match.Name, "${$1}")
		} else {
			v, ok := params[match.Name]
			if !ok {
				continue
			}
			newValue = v
		}
		if len(matches) == 1 && len(source) == len(match.FullMatch) {
			return newValue
		}
		result = strings.Replace(result, match.FullMatch, stringify(newValue), -1)
	}
	return result
}

func stringify(val interface{}) string {
	switch val := val.(type) {
	case string:
		return val
	case nil:
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
