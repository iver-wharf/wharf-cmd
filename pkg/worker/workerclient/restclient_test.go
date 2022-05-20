package workerclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/iver-wharf/wharf-core/v2/pkg/problem"
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
		resp    *http.Response
		err     error
		wantErr error
	}{
		{
			name:    "no problem response gives error",
			resp:    &http.Response{},
			err:     errNormal,
			wantErr: errNormal,
		},
		{
			name:    "no problem response and nil error gives nil",
			resp:    &http.Response{},
			err:     nil,
			wantErr: nil,
		},
		{
			name:    "nil response gives error",
			resp:    nil,
			err:     nil,
			wantErr: errors.New("response is nil"),
		},
		{
			name: "malformed problem response gives parse error",
			resp: &http.Response{
				Header: http.Header{
					"Content-Type": []string{problem.HTTPContentType},
				},
				Body: io.NopCloser(bytes.NewReader([]byte{})),
			},
			err:     nil,
			wantErr: errParse,
		},
		{
			name: "problem response gives problem response",
			resp: &http.Response{
				Header: http.Header{
					"Content-Type": []string{problem.HTTPContentType},
				},
				Body: io.NopCloser(bytes.NewReader(probBytes)),
			},
			err:     nil,
			wantErr: errProblem,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := assertResponseOK(tc.resp, tc.err)
			assert.Equal(t, fmt.Sprint(tc.wantErr), fmt.Sprint(gotErr))
		})
	}
}

func TestNewRestClient(t *testing.T) {
	testCases := []struct {
		name           string
		secure         bool
		wantSecure     bool
		wantNilRootCAs bool
	}{
		{
			name:       "secure_client",
			secure:     true,
			wantSecure: true,
		},
		{
			name:       "insecure_client",
			secure:     false,
			wantSecure: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := newRestClient(Options{
				InsecureSkipVerify: !tc.secure,
			})
			assert.NoError(t, err)
			assert.Equal(t, tc.wantSecure, !c.client.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify)
		})
	}
}
