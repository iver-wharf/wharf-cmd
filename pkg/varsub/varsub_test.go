package varsub

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetListOfParamsNames(t *testing.T) {
	type testCase struct {
		name           string
		source         string
		expectedResult []VarMatch
	}

	tests := []testCase{
		{
			name:           "text without variable",
			source:         "text without variable",
			expectedResult: nil,
		},
		{
			name:           "simple variable",
			source:         "${lorem}",
			expectedResult: []VarMatch{{Name: "lorem", Syntax: "${lorem}"}},
		},
		{
			name:           "invalid simple variable",
			source:         "${lorem ipsum}",
			expectedResult: []VarMatch{{Name: "lorem ipsum", Syntax: "${lorem ipsum}"}},
		},
		{
			name:           "simple text with variable",
			source:         "Foo ${lorem} bar",
			expectedResult: []VarMatch{{Name: "lorem", Syntax: "${lorem}"}},
		},
		{
			name:           "simple text with variable and white spaces",
			source:         "Foo ${\n \tlorem\r} bar",
			expectedResult: []VarMatch{{Name: "lorem", Syntax: "${\n \tlorem\r}"}},
		},
		{
			name:           "simple text with escaped variable",
			source:         "Foo ${%lorem%} bar",
			expectedResult: []VarMatch{{Name: "%lorem%", Syntax: "${%lorem%}"}},
		},
		{
			name:           "simple text with escaped empty string",
			source:         "Foo ${%%} bar",
			expectedResult: []VarMatch{{Name: "%%", Syntax: "${%%}"}},
		},
		{
			name:           "simple text with escaped empty string by singular percent",
			source:         "Foo ${%} bar",
			expectedResult: []VarMatch{{Name: "%", Syntax: "${%}"}},
		},
		{
			name:           "simple text with escaped white signs",
			source:         "Foo ${%\n \r%} bar",
			expectedResult: []VarMatch{{Name: "%\n \r%", Syntax: "${%\n \r%}"}},
		},
		{
			name:           "simple text with escaped white signs 2",
			source:         "Foo ${\t%\n \r% } bar",
			expectedResult: []VarMatch{{Name: "%\n \r%", Syntax: "${\t%\n \r% }"}},
		},
		{
			name:           "simple text with invalid escaped text",
			source:         "Foo ${%lorem} bar",
			expectedResult: []VarMatch{{Name: "%lorem", Syntax: "${%lorem}"}},
		},
		{
			name:           "simple text with invalid variable",
			source:         "Foo ${} bar",
			expectedResult: nil,
		},
		{
			name:   "three variables",
			source: "${lorem} ${ipsum} ${dolor}",
			expectedResult: []VarMatch{
				{Name: "lorem", Syntax: "${lorem}"},
				{Name: "ipsum", Syntax: "${ipsum}"},
				{Name: "dolor", Syntax: "${dolor}"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := GetVarMatches(tc.source)
			eq := reflect.DeepEqual(tc.expectedResult, result)
			assert.True(t, eq, fmt.Sprintf("should be: %v, got: %v", tc.expectedResult, result))
		})
	}
}

func TestReplaceVariables(t *testing.T) {
	type testCase struct {
		name           string
		source         string
		expectedResult string
	}

	tests := []testCase{
		{
			name:           "simple variable",
			source:         "${lorem}",
			expectedResult: "ipsum",
		},
		{
			name:           "invalid simple variable",
			source:         "${lorem ipsum}",
			expectedResult: "${lorem ipsum}",
		},
		{
			name:           "simple text with variable",
			source:         "Foo ${lorem} bar",
			expectedResult: "Foo ipsum bar",
		},
		{
			name:           "simple text with variable and white spaces",
			source:         "Foo ${\n \tlorem\r} bar",
			expectedResult: "Foo ipsum bar",
		},
		{
			name:           "simple text with escaped variable",
			source:         "Foo ${%lorem%} bar",
			expectedResult: "Foo ${lorem} bar",
		},
		{
			name:           "simple text with escaped empty string",
			source:         "Foo ${%%} bar",
			expectedResult: "Foo ${} bar",
		},
		{
			name:           "simple text with escaped empty string by singular percent",
			source:         "Foo ${%} bar",
			expectedResult: "Foo ${} bar",
		},
		{
			name:           "simple text with escaped empty white signs",
			source:         "Foo ${%\n \r%} bar",
			expectedResult: "Foo ${\n \r} bar",
		},
		{
			name:           "simple text with escaped empty white signs 2",
			source:         "Foo ${ %\n \r%\n} bar",
			expectedResult: "Foo ${\n \r} bar",
		},
		{
			name:           "simple text with invalid escaped text",
			source:         "Foo ${%lorem} bar",
			expectedResult: "Foo ${%lorem} bar",
		},
		{
			name:           "simple text with invalid variable",
			source:         "Foo ${} bar",
			expectedResult: "Foo ${} bar",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ReplaceVariables(tc.source, map[string]interface{}{"lorem": "ipsum"})
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
