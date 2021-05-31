package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizationFlagsHasFlag(t *testing.T) {
	type testCase struct {
		name  string
		flags SanitizationFlags
		flag  SanitizationFlags
		has   bool
	}

	tests := []testCase{
		{
			name:  "Both methods set, Last CR under test",
			flags: AllSanitizationMethods,
			flag:  FromLastCR,
			has:   true,
		},
		{
			name:  "Both methods set, Remove Ansi Codes under test",
			flags: FromLastCR | RemoveAnsiCodes,
			flag:  RemoveAnsiCodes,
			has:   true,
		},
		{
			name:  "flag value 0x1111, Remove Ansi Codes under test",
			flags: 0xF,
			flag:  RemoveAnsiCodes,
			has:   true,
		},
		{
			name:  "flag value 0x0011, Remove Ansi Codes under test",
			flags: 0x3,
			flag:  RemoveAnsiCodes,
			has:   true,
		},
		{
			name:  "flag value 0x1111, invalid value under test",
			flags: 0xF,
			flag:  0x8,
			has:   true,
		},
		{
			name:  "Last CR method set and Last CR under test",
			flags: FromLastCR,
			flag:  FromLastCR,
			has:   true,
		},
		{
			name:  "Remove ansi codes method set and ansi codes under test",
			flags: RemoveAnsiCodes,
			flag:  RemoveAnsiCodes,
			has:   true,
		},
		{
			name:  "Last CR method set and remove ansi codes under test",
			flags: FromLastCR,
			flag:  RemoveAnsiCodes,
			has:   false,
		},
		{
			name:  "Remove ansi codes method set and Last CR under test",
			flags: RemoveAnsiCodes,
			flag:  FromLastCR,
			has:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			has := tc.flags.HasFlag(tc.flag)
			assert.Equal(t, tc.has, has)
		})
	}
}
