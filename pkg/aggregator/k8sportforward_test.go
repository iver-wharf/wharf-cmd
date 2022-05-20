package aggregator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPortForwardURL(t *testing.T) {
	const (
		namespace = "my-ns"
		podName   = "my-pod"
	)
	var tests = []struct {
		name   string
		apiURL string
		want   string
	}{
		{
			name:   "ip",
			apiURL: "https://172.50.123.3:6443",
			want:   "https://172.50.123.3:6443/api/v1/namespaces/my-ns/pods/my-pod/portforward",
		},
		{
			name:   "ip with user",
			apiURL: "https://user:password@172.50.123.3:6443",
			want:   "https://user:password@172.50.123.3:6443/api/v1/namespaces/my-ns/pods/my-pod/portforward",
		},
		{
			name:   "rancher url",
			apiURL: "https://rancher.example.com/k8s/clusters/c-m-13mz8a32",
			want:   "https://rancher.example.com/k8s/clusters/c-m-13mz8a32/api/v1/namespaces/my-ns/pods/my-pod/portforward",
		},
		{
			name:   "preserves query params",
			apiURL: "https://172.50.123.3:6443?foo=bar",
			want:   "https://172.50.123.3:6443/api/v1/namespaces/my-ns/pods/my-pod/portforward?foo=bar",
		},
		{
			name:   "preserves protocol",
			apiURL: "foobar://172.50.123.3:6443",
			want:   "foobar://172.50.123.3:6443/api/v1/namespaces/my-ns/pods/my-pod/portforward",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := newPortForwardURL(tc.apiURL, namespace, podName)
			require.NoError(t, err)
			gotStr := got.String()
			assert.Equal(t, tc.want, gotStr)
		})
	}
}
