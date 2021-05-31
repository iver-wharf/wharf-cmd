package utils

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveANSIcodes(t *testing.T) {
	type testCase struct {
		name        string
		data        []byte
		expectedStr string
	}

	tests := []testCase{
		{
			name:        "empty string",
			data:        []byte(""),
			expectedStr: "",
		},
		{
			name:        "no ANSI codes",
			data:        []byte("string without ANSI codes"),
			expectedStr: "string without ANSI codes",
		},
		{
			name:        "ANSI code at the beginning",
			data:        []byte("\x1b[Kremote: Enumerating objects: 52, done."),
			expectedStr: "remote: Enumerating objects: 52, done.",
		},
		{
			name:        "ANSI code at the end",
			data:        []byte("remote: Enumerating objects: 52, done.\x1b[K"),
			expectedStr: "remote: Enumerating objects: 52, done.",
		},
		{
			name:        "Only ANSI codes",
			data:        []byte("\x1b[K\x1b[K\x1b[K"),
			expectedStr: "",
		},
		{
			name:        "Only ANSI codes with CRLF between",
			data:        []byte("\x1b[K\x1b[K\r\x1b[K"),
			expectedStr: "\r",
		},
		{
			name:        "string with one ANSI code but with CR and new line signs",
			data:        []byte("remote: Enumerating objects: 52, done.\x1b[K\r\n"),
			expectedStr: "remote: Enumerating objects: 52, done.\r\n",
		},
		{
			name: "string with many ANSI codes but with CR and new line signs",
			data: []byte("remote: Enumerating objects: 52, done.\x1b[K\r\n" +
				"remote: Compressing objects: 100% (33/33), done.\x1b[K\r\n"),
			expectedStr: "remote: Enumerating objects: 52, done.\r\n" +
				"remote: Compressing objects: 100% (33/33), done.\r\n",
		},
		{
			name: "string with many ANSI codes but without new line signs",
			data: []byte("remote: Enumerating objects: 52, done.\x1b[K" +
				"remote: Compressing objects: 100% (33/33), done.\x1b[K"),
			expectedStr: "remote: Enumerating objects: 52, done." +
				"remote: Compressing objects: 100% (33/33), done.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := removeANSIcodes(tc.data)
			assert.Equal(t, tc.expectedStr, string(result))
		})
	}
}

func TestGetTextFromLastCR(t *testing.T) {
	type testCase struct {
		name        string
		data        []byte
		expectedStr string
	}

	tests := []testCase{
		{
			name:        "empty string",
			data:        []byte(""),
			expectedStr: "",
		},
		{
			name:        "string without CR",
			data:        []byte("string without CR"),
			expectedStr: "string without CR",
		},
		{
			name:        "string with one CR at the end",
			data:        []byte("remote: Enumerating objects: 52, done.\r"),
			expectedStr: "",
		},
		{
			name: "string with many CR signs",
			data: []byte("remote: Enumerating objects: 52, done.\r\n" +
				"remote: Counting objects:   1% (1/52)\r" +
				"remote: Counting objects: 100% (52/52)\r" +
				"remote: Counting objects: 100% (52/52), done.\r\n" +
				"remote: Compressing objects:   3% (1/33)\r" +
				"remote: Compressing objects: 100% (33/33)\r" +
				"remote: Compressing objects: 100% (33/33), done."),
			expectedStr: "remote: Compressing objects: 100% (33/33), done.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := fromLastCR(tc.data)
			assert.Equal(t, tc.expectedStr, string(result))
		})
	}
}

func TestSanitizeLogs(t *testing.T) {
	type testCase struct {
		name        string
		data        string
		expectedStr string
	}

	tests := []testCase{
		{
			name:        "empty string",
			data:        "",
			expectedStr: "",
		},
		{
			name:        "string without ANSI codes and new line signs",
			data:        "string without ANSI codes",
			expectedStr: "string without ANSI codes",
		},
		{
			name:        "string with one ANSI code and without new line",
			data:        "remote: Enumerating objects: 52, done.\x1b[K",
			expectedStr: "remote: Enumerating objects: 52, done.",
		},
		{
			name:        "string with one ANSI code but with CR and without new line sign",
			data:        "remote: Enumerating objects: 52, done.\x1b[K\r",
			expectedStr: "remote: Enumerating objects: 52, done.",
		},
		{
			name:        "string with one ANSI code but without CR and with new line sign",
			data:        "remote: Enumerating objects: 52, done.\x1b[K\n",
			expectedStr: "remote: Enumerating objects: 52, done.",
		},
		{
			name:        "string with one ANSI code but with CR and new line signs",
			data:        "remote: Enumerating objects: 52, done.\x1b[K\r\n",
			expectedStr: "remote: Enumerating objects: 52, done.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rc := ioutil.NopCloser(strings.NewReader(tc.data))
			sut := NewStreamScanner(rc, RemoveAnsiCodes | FromLastCR)

			for sut.Scan() {
				text := sut.Text()
				assert.Equal(t, tc.expectedStr, text)
			}
			rc.Close()
			
			err := sut.Err()
			assert.Nil(t, err)
		})
	}
}

func TestStreamContainerLogs(t *testing.T) {
	testData := "Cloning into '/gitRepo'...\r\n" +
		"remote: Enumerating objects: 52, done.\x1b[K\r\n" +
		"remote: Counting objects:   1% (1/52)\x1b[K\r" +
		"remote: Counting objects:   3% (2/52)\x1b[K\r" +
		"remote: Counting objects:  98% (51/52)\x1b[K\r" +
		"remote: Counting objects: 100% (52/52)\x1b[K\r" +
		"remote: Counting objects: 100% (52/52), done.\x1b[K\r\n" +
		"remote: Compressing objects:   3% (1/33)\x1b[K\r" +
		"remote: Compressing objects:   6% (2/33)\x1b[K\r" +
		"remote: Compressing objects:  93% (31/33)\x1b[K\r" +
		"remote: Compressing objects:  96% (32/33)\x1b[K\r" +
		"remote: Compressing objects: 100% (33/33)\x1b[K\r" +
		"remote: Compressing objects: 100% (33/33), done.\x1b[K\r\n"

	expectedData := []string{
		"Cloning into '/gitRepo'...",
		"remote: Enumerating objects: 52, done.",
		"remote: Counting objects: 100% (52/52), done.",
		"remote: Compressing objects: 100% (33/33), done.",
	}

	rc := ioutil.NopCloser(strings.NewReader(testData))
	sut := NewStreamScanner(rc, AllSanitizationMethods)

	for i := 0; sut.Scan(); i++ {
		require.Less(t, i, len(expectedData))
		assert.Equal(t, expectedData[i], sut.Text())
	}
	rc.Close()

	err := sut.Err()
	require.Nil(t, err)
}
