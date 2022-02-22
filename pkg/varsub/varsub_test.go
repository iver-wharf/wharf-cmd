package varsub

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatches(t *testing.T) {
	type testMatch struct {
		Name      string
		FullMatch string
	}
	tests := []struct {
		name   string
		source string
		want   []testMatch
	}{
		{
			name:   "text without variable",
			source: "text without variable",
			want:   nil,
		},
		{
			name:   "simple variable",
			source: "${lorem}",
			want:   []testMatch{{Name: "lorem", FullMatch: "${lorem}"}},
		},
		{
			name:   "invalid simple variable",
			source: "${lorem ipsum}",
			want:   []testMatch{{Name: "lorem ipsum", FullMatch: "${lorem ipsum}"}},
		},
		{
			name:   "simple text with variable",
			source: "Foo ${lorem} bar",
			want:   []testMatch{{Name: "lorem", FullMatch: "${lorem}"}},
		},
		{
			name:   "simple text with variable and white spaces",
			source: "Foo ${\n \tlorem\r} bar",
			want:   []testMatch{{Name: "lorem", FullMatch: "${\n \tlorem\r}"}},
		},
		{
			name:   "simple text with escaped variable",
			source: "Foo ${%lorem%} bar",
			want:   []testMatch{{Name: "%lorem%", FullMatch: "${%lorem%}"}},
		},
		{
			name:   "simple text with escaped empty string",
			source: "Foo ${%%} bar",
			want:   []testMatch{{Name: "%%", FullMatch: "${%%}"}},
		},
		{
			name:   "simple text with escaped empty string by singular percent",
			source: "Foo ${%} bar",
			want:   []testMatch{{Name: "%", FullMatch: "${%}"}},
		},
		{
			name:   "simple text with escaped white signs",
			source: "Foo ${%\n \r%} bar",
			want:   []testMatch{{Name: "%\n \r%", FullMatch: "${%\n \r%}"}},
		},
		{
			name:   "simple text with escaped white signs 2",
			source: "Foo ${\t%\n \r% } bar",
			want:   []testMatch{{Name: "%\n \r%", FullMatch: "${\t%\n \r% }"}},
		},
		{
			name:   "simple text with invalid escaped text",
			source: "Foo ${%lorem} bar",
			want:   []testMatch{{Name: "%lorem", FullMatch: "${%lorem}"}},
		},
		{
			name:   "simple text with invalid variable",
			source: "Foo ${} bar",
			want:   nil,
		},
		{
			name:   "three variables",
			source: "${lorem} ${ipsum} ${dolor}",
			want: []testMatch{
				{Name: "lorem", FullMatch: "${lorem}"},
				{Name: "ipsum", FullMatch: "${ipsum}"},
				{Name: "dolor", FullMatch: "${dolor}"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			matches := Matches(tc.source)
			if len(tc.want) == 0 {
				assert.Len(t, matches, 0)
				return
			}
			got := make([]testMatch, len(matches))
			for i, m := range matches {
				got[i] = testMatch{Name: m.Name, FullMatch: m.FullMatch}
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSubstitute(t *testing.T) {
	vars := map[string]interface{}{
		"lorem": "ipsum",
	}
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{
			name:   "simple variable",
			source: "${lorem}",
			want:   "ipsum",
		},
		{
			name:   "invalid simple variable",
			source: "${lorem ipsum}",
			want:   "${lorem ipsum}",
		},
		{
			name:   "simple text with variable",
			source: "Foo ${lorem} bar",
			want:   "Foo ipsum bar",
		},
		{
			name:   "simple text with variable and white spaces",
			source: "Foo ${\n \tlorem\r} bar",
			want:   "Foo ipsum bar",
		},
		{
			name:   "simple text with escaped variable",
			source: "Foo ${%lorem%} bar",
			want:   "Foo ${lorem} bar",
		},
		{
			name:   "simple text with escaped empty string",
			source: "Foo ${%%} bar",
			want:   "Foo ${} bar",
		},
		{
			name:   "simple text with escaped empty string by singular percent",
			source: "Foo ${%} bar",
			want:   "Foo ${} bar",
		},
		{
			name:   "simple text with escaped empty white signs",
			source: "Foo ${%\n \r%} bar",
			want:   "Foo ${\n \r} bar",
		},
		{
			name:   "simple text with escaped empty white signs 2",
			source: "Foo ${ %\n \r%\n} bar",
			want:   "Foo ${\n \r} bar",
		},
		{
			name:   "simple text with invalid escaped text",
			source: "Foo ${%lorem} bar",
			want:   "Foo ${%lorem} bar",
		},
		{
			name:   "simple text with invalid variable",
			source: "Foo ${} bar",
			want:   "Foo ${} bar",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Substitute(tc.source, vars)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSubstitute_nonStrings(t *testing.T) {
	tests := []struct {
		name   string
		vars   map[string]interface{}
		source string
		want   interface{}
	}{
		{
			name:   "full/bool",
			vars:   map[string]interface{}{"lorem": true},
			source: "${lorem}",
			want:   true,
		},
		{
			name:   "full/int",
			vars:   map[string]interface{}{"lorem": 123},
			source: "${lorem}",
			want:   123,
		},
		{
			name:   "full/float",
			vars:   map[string]interface{}{"lorem": 123.0},
			source: "${lorem}",
			want:   123.0,
		},
		{
			name:   "full/nil",
			vars:   map[string]interface{}{"lorem": nil},
			source: "${lorem}",
			want:   nil,
		},
		{
			name:   "embed/bool",
			vars:   map[string]interface{}{"lorem": true},
			source: "foo ${lorem} bar",
			want:   "foo true bar",
		},
		{
			name:   "embed/int",
			vars:   map[string]interface{}{"lorem": 123},
			source: "foo ${lorem} bar",
			want:   "foo 123 bar",
		},
		{
			name:   "embed/float",
			vars:   map[string]interface{}{"lorem": 123.0},
			source: "foo ${lorem} bar",
			want:   "foo 123 bar",
		},
		{
			name:   "embed/nil",
			vars:   map[string]interface{}{"lorem": nil},
			source: "foo ${lorem} bar",
			want:   "foo  bar",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Substitute(tc.source, tc.vars)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSubstitute_recursive(t *testing.T) {
	tests := []struct {
		name   string
		vars   map[string]interface{}
		source string
		want   interface{}
	}{
		{
			name: "string",
			vars: map[string]interface{}{
				"one": "11${two}11",
				"two": 2222,
			},
			source: "00${one}00",
			want:   "001122221100",
		},
		{
			name: "typed",
			vars: map[string]interface{}{
				"one": "${two}",
				"two": 2222,
			},
			source: "${one}",
			want:   2222,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Substitute(tc.source, tc.vars)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSubstitute_errIfRecursiveLoop(t *testing.T) {
	vars := map[string]interface{}{
		"lorem": "ipsum: ${lorem}",
	}
	result, err := Substitute("root: ${lorem}", vars)
	assert.ErrorIsf(t, err, ErrRecursiveLoop, "unexpected result: %q", result)
}
