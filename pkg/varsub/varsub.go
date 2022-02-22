package varsub

import (
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

func Substitute(source string, params map[string]interface{}) string {
	result := source
	varMatches := Matches(source)

	var newValue string
	for _, match := range varMatches {
		ok := false

		if match.Name == "%" {
			newValue = "${}"
			ok = true
		} else if escapedParamPattern.MatchString(match.Name) {
			newValue = escapedParamPattern.ReplaceAllString(match.Name, "${$1}")
			ok = true
		} else {
			newValue, ok = params[match.Name].(string)
		}

		if ok {
			result = strings.Replace(result, match.FullMatch, newValue, -1)
		}
	}

	return result
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
