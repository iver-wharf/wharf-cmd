package worker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizePodName(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "valid",
			input: "some-pod-name-123",
			want:  "some-pod-name-123",
		},
		{
			name:  "replaces invalid chars",
			input: "some/pod%name.com",
			want:  "some-pod-name-com",
		},
		{
			name:  "only dashes",
			input: "---%/--.-",
			want:  "",
		},
		{
			name:  "emoji",
			input: "party-🎉-popper",
			want:  "party---popper",
		},
		{
			name:  "enforce lowercase",
			input: "Some-Pod-Name",
			want:  "some-pod-name",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizePodName(tc.input)
			assert.Equal(t, tc.want, got)
		})
	}
}
