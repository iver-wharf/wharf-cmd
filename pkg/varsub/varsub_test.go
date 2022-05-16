package varsub

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatches(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  []VarMatch
	}{
		{
			name:  "text without variable",
			value: "text without variable",
			want:  nil,
		},
		{
			name:  "simple variable",
			value: "${lorem}",
			want:  []VarMatch{{Name: "lorem", FullMatch: "${lorem}"}},
		},
		{
			name:  "invalid simple variable",
			value: "${lorem ipsum}",
			want:  []VarMatch{{Name: "lorem ipsum", FullMatch: "${lorem ipsum}"}},
		},
		{
			name:  "simple text with variable",
			value: "Foo ${lorem} bar",
			want:  []VarMatch{{Name: "lorem", FullMatch: "${lorem}"}},
		},
		{
			name:  "simple text with variable and white spaces",
			value: "Foo ${\n \tlorem\r} bar",
			want:  []VarMatch{{Name: "lorem", FullMatch: "${\n \tlorem\r}"}},
		},
		{
			name:  "simple text with escaped variable",
			value: "Foo ${%lorem%} bar",
			want:  []VarMatch{{Name: "%lorem%", FullMatch: "${%lorem%}"}},
		},
		{
			name:  "simple text with escaped empty string",
			value: "Foo ${%%} bar",
			want:  []VarMatch{{Name: "%%", FullMatch: "${%%}"}},
		},
		{
			name:  "simple text with escaped empty string by singular percent",
			value: "Foo ${%} bar",
			want:  []VarMatch{{Name: "%", FullMatch: "${%}"}},
		},
		{
			name:  "simple text with escaped white signs",
			value: "Foo ${%\n \r%} bar",
			want:  []VarMatch{{Name: "%\n \r%", FullMatch: "${%\n \r%}"}},
		},
		{
			name:  "simple text with escaped white signs 2",
			value: "Foo ${\t%\n \r% } bar",
			want:  []VarMatch{{Name: "%\n \r%", FullMatch: "${\t%\n \r% }"}},
		},
		{
			name:  "simple text with invalid escaped text",
			value: "Foo ${%lorem} bar",
			want:  []VarMatch{{Name: "%lorem", FullMatch: "${%lorem}"}},
		},
		{
			name:  "simple text with invalid variable",
			value: "Foo ${} bar",
			want:  nil,
		},
		{
			name:  "three variables",
			value: "${lorem} ${ipsum} ${dolor}",
			want: []VarMatch{
				{Name: "lorem", FullMatch: "${lorem}"},
				{Name: "ipsum", FullMatch: "${ipsum}"},
				{Name: "dolor", FullMatch: "${dolor}"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Matches(tc.value)
			if len(tc.want) == 0 {
				assert.Len(t, got, 0)
				return
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSubstitute(t *testing.T) {
	source := SourceMap{
		"lorem":   Val{Value: "ipsum"},
		"foo-bar": Val{Value: "smilie"},
	}
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "simple variable",
			value: "${lorem}",
			want:  "ipsum",
		},
		{
			name:  "undefined simple variable",
			value: "${lorem ipsum}",
			want:  "${lorem ipsum}",
		},
		{
			name:  "simple text with variable",
			value: "Foo ${lorem} bar",
			want:  "Foo ipsum bar",
		},
		{
			name:  "simple text with kebab variable",
			value: "Foo ${foo-bar} bar",
			want:  "Foo smilie bar",
		},
		{
			name:  "simple text with variable and white spaces",
			value: "Foo ${\n \tlorem\r} bar",
			want:  "Foo ipsum bar",
		},
		{
			name:  "simple text with escaped variable",
			value: "Foo ${%lorem%} bar",
			want:  "Foo ${lorem} bar",
		},
		{
			name:  "simple text with escaped empty string",
			value: "Foo ${%%} bar",
			want:  "Foo ${} bar",
		},
		{
			name:  "simple text with escaped empty string by singular percent",
			value: "Foo ${%} bar",
			want:  "Foo ${} bar",
		},
		{
			name:  "simple text with escaped empty white signs",
			value: "Foo ${%\n \r%} bar",
			want:  "Foo ${\n \r} bar",
		},
		{
			name:  "simple text with escaped empty white signs 2",
			value: "Foo ${ %\n \r%\n} bar",
			want:  "Foo ${ %\n \r%\n} bar",
		},
		{
			name:  "unescaped variables don't mess with substitution of matching vars after",
			value: "Foo ${lorem} ${%lorem%} ${lorem}",
			want:  "Foo ipsum ${lorem} ipsum",
		},
		{
			name:  "simple text with invalid escaped text",
			value: "Foo ${%lorem} bar",
			want:  "Foo ${%lorem} bar",
		},
		{
			name:  "simple text with invalid variable",
			value: "Foo ${} bar",
			want:  "Foo ${} bar",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Substitute(tc.value, source)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSubstitute_nonStrings(t *testing.T) {
	tests := []struct {
		name   string
		source SourceMap
		value  string
		want   any
	}{
		{
			name:   "full/bool",
			source: SourceMap{"lorem": Val{Value: true}},
			value:  "${lorem}",
			want:   true,
		},
		{
			name:   "full/int",
			source: SourceMap{"lorem": Val{Value: 123}},
			value:  "${lorem}",
			want:   123,
		},
		{
			name:   "full/float",
			source: SourceMap{"lorem": Val{Value: 123.0}},
			value:  "${lorem}",
			want:   123.0,
		},
		{
			name:   "full/nil",
			source: SourceMap{"lorem": Val{Value: nil}},
			value:  "${lorem}",
			want:   nil,
		},
		{
			name:   "embed/bool",
			source: SourceMap{"lorem": Val{Value: true}},
			value:  "foo ${lorem} bar",
			want:   "foo true bar",
		},
		{
			name:   "embed/int",
			source: SourceMap{"lorem": Val{Value: 123}},
			value:  "foo ${lorem} bar",
			want:   "foo 123 bar",
		},
		{
			name:   "embed/float",
			source: SourceMap{"lorem": Val{Value: 123.0}},
			value:  "foo ${lorem} bar",
			want:   "foo 123 bar",
		},
		{
			name:   "embed/nil",
			source: SourceMap{"lorem": Val{Value: nil}},
			value:  "foo ${lorem} bar",
			want:   "foo  bar",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Substitute(tc.value, tc.source)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSubstitute_recursive(t *testing.T) {
	tests := []struct {
		name   string
		source SourceMap
		value  string
		want   any
	}{
		{
			name: "string",
			source: SourceMap{
				"one": Val{Value: "11${two}11"},
				"two": Val{Value: 2222},
			},
			value: "00${one}00",
			want:  "001122221100",
		},
		{
			name: "typed",
			source: SourceMap{
				"one": Val{Value: "${two}"},
				"two": Val{Value: 222},
			},
			value: "${one}",
			want:  222,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Substitute(tc.value, tc.source)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSubstitute_errIfRecursiveLoop(t *testing.T) {
	source := SourceMap{
		"lorem": Val{Value: "ipsum: ${lorem}"},
	}
	result, err := Substitute("root: ${lorem}", source)
	assert.ErrorIsf(t, err, ErrRecursiveLoop, "unexpected result: %q", result)
}
