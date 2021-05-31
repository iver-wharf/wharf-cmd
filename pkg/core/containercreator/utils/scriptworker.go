package utils

import "regexp"

const ParamPattern   = `\${([\w\_]+)}`

var paramPattern = regexp.MustCompile(ParamPattern)

func GetListOfParamsNames(source string) map[string]string {
	params := make(map[string]string)
	if !paramPattern.MatchString(source) {
		return params
	}

	matches := paramPattern.FindAllStringSubmatch(source, -1)

	names := make(map[string]bool)
	for _, match := range matches {
		if _, alreadySet := names[match[1]]; !alreadySet {
			names[match[1]] = true
			params[match[1]] = match[0]
		}
	}

	return params
}