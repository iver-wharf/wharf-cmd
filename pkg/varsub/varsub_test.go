package varsub

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			got := Substitute(tc.source, map[string]interface{}{
				"lorem": "ipsum",
			})
			assert.Equal(t, tc.want, got)
		})
	}
}
