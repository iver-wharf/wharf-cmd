package utils

import (
	"regexp"
	"strings"
)

type VarMatch struct {
	Name   string
	Syntax string
}

var varSyntaxPattern = regexp.MustCompile(`\${\s*(?P<paramName>%*[\w_\s]*%*)\s*}`)
var paramNamePattern = regexp.MustCompile(`\s*(?P<cleanParamName>\w*[\s_]*\w+)\s*`)
var escapedParamPattern = regexp.MustCompile(`%(?P<paramValue>\s*[\w_\s]*\s*)%`)

func ReplaceVariables(source string, params map[string]interface{}) string {
	result := source
	varMatches := GetVarMatches(source)

	var newValue string
	for _, match := range varMatches {
		ok := false

		if match.Name == "%" {
			newValue = "${}"
			ok = true
		} else if escapedParamPattern.MatchString(match.Name) {
			newValue = escapedParamPattern.ReplaceAllString(match.Name, "${$paramValue}")
			ok = true
		} else {
			newValue, ok = params[match.Name].(string)
		}

		if ok {
			result = strings.Replace(result, match.Syntax, newValue, -1)
		}
	}

	return result
}

func GetVarMatches(source string) []VarMatch {
	matches := varSyntaxPattern.FindAllStringSubmatch(source, -1)
	var params []VarMatch

	for _, match := range matches {
		paramName := match[1]

		if paramName == "" {
			continue
		}

		if paramName[0] != '%' {
			paramName = paramNamePattern.ReplaceAllString(paramName, "$cleanParamName")
		}

		params = append(params, VarMatch{
			Name:   paramName,
			Syntax: match[0],
		})
	}

	return params
}
