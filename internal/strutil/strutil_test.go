package strutil

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringify(t *testing.T) {
	var tests = []struct {
		name string
		in   any
		want string
	}{
		{
			name: "nil",
			in:   nil,
			want: "",
		},
		{
			name: "float32 in decimal",
			in:   float32(0.0000013631),
			want: "0.0000013631",
		},
		{
			name: "float64 in decimal",
			in:   float64(0.0000013631),
			want: "0.0000013631",
		},
		{
			name: "boolean",
			in:   true,
			want: "true",
		},
		{
			name: "int64",
			in:   math.MaxInt64,
			want: "9223372036854775807",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Stringify(tc.in)
			assert.Equal(t, tc.want, got)
		})
	}
}
