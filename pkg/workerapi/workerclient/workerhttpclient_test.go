package workerclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/iver-wharf/wharf-core/pkg/problem"
	"github.com/stretchr/testify/assert"
)

var errNormal = errors.New("normal error")
var errParse = errors.New("failed to parse \"application/problem+json\" problem response: unexpected end of JSON input")
var errProblem = problem.Response{
	Type:   "/prob/something",
	Detail: "Some details",
	Status: 500,
}

func TestAssertResponseOK(t *testing.T) {
	probBytes, err := json.Marshal(&errProblem)
	assert.Nil(t, err)
	testCases := []struct {
		name    string
		wantErr error
		resp    *http.Response
		err     error
	}{
		{
			name:    "no problem response gives error",
			wantErr: errNormal,
			resp:    &http.Response{},
			err:     errNormal,
		},
		{
			name:    "no problem response and nil error gives nil",
			wantErr: nil,
			resp:    &http.Response{},
			err:     nil,
		},
		{
			name:    "all nil gives nil",
			wantErr: nil,
			resp:    nil,
			err:     nil,
		},
		{
			name:    "malformed problem response gives parse error",
			wantErr: errProblem,
			resp: &http.Response{
				Header: http.Header{
					"Content-Type": []string{problem.HTTPContentType},
				},
				Body: io.NopCloser(bytes.NewReader(probBytes)),
			},
			err: errNormal,
		},
		{
			name:    "problem response gives problem response",
			wantErr: errParse,
			resp: &http.Response{
				Header: http.Header{
					"Content-Type": []string{problem.HTTPContentType},
				},
				Body: io.NopCloser(bytes.NewReader([]byte{})),
			},
			err: errNormal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := assertResponseOK(tc.resp, tc.err)
			assert.Equal(t, fmt.Sprintf("%v", tc.wantErr), fmt.Sprintf("%v", gotErr))
		})
	}
}
