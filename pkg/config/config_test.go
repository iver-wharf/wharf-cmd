package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestValidateModifiesImagePullPolicyValue(t *testing.T) {
	testCases := []struct {
		name    string
		cfg     Config
		wantCfg Config
		wantOK  bool
	}{
		{
			name:    "both policies good",
			cfg:     newConfigWithPullPolicies("always", "never"),
			wantCfg: newConfigWithPullPolicies(string(v1.PullAlways), string(v1.PullNever)),
			wantOK:  true,
		},
		{
			name:    "first is bad - early exit",
			cfg:     newConfigWithPullPolicies("bad pull policy", "always"),
			wantCfg: newConfigWithPullPolicies("", "always"),
			wantOK:  false,
		},
		{
			name:    "second is bad",
			cfg:     newConfigWithPullPolicies("always", "bad pull policy"),
			wantCfg: newConfigWithPullPolicies(string(v1.PullAlways), ""),
			wantOK:  false,
		},
		{
			name:    "empty is bad",
			cfg:     newConfigWithPullPolicies("", ""),
			wantCfg: newConfigWithPullPolicies("", ""),
			wantOK:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.validate()
			if tc.wantOK {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
			assert.Equal(t, tc.wantCfg, tc.cfg)
		})
	}
}

func newConfigWithPullPolicies(initPullPolicy, pullPolicy string) Config {
	return Config{
		Provisioner: ProvisionerConfig{
			K8s: ProvisionerK8sConfig{
				Worker: ProvisionerK8sWorkerConfig{
					InitContainer: K8sContainerConfig{
						ImagePullPolicy: v1.PullPolicy(initPullPolicy),
					},
					Container: K8sContainerConfig{
						ImagePullPolicy: v1.PullPolicy(pullPolicy),
					},
				},
			},
		},
	}
}

func TestParsePullPolicy(t *testing.T) {
	testCases := []struct {
		name    string
		toParse string
		want    v1.PullPolicy
		wantOK  bool
	}{
		{
			name:    "parse exact match - Always",
			toParse: "Always",
			want:    v1.PullAlways,
			wantOK:  true,
		},
		{
			name:    "parse exact match - Never",
			toParse: "Never",
			want:    v1.PullNever,
			wantOK:  true,
		},
		{
			name:    "parse exact match - IfNotPresent",
			toParse: "IfNotPresent",
			want:    v1.PullIfNotPresent,
			wantOK:  true,
		},
		{
			name:    "parse mixed-case match - alWayS",
			toParse: "alWayS",
			want:    v1.PullAlways,
			wantOK:  true,
		},
		{
			name:    "parse mixed-case match - neVEr",
			toParse: "neVEr",
			want:    v1.PullNever,
			wantOK:  true,
		},
		{
			name:    "parse mixed-case match - ifNotPresEnt",
			toParse: "ifNotPresEnt",
			want:    v1.PullIfNotPresent,
			wantOK:  true,
		},
		{
			name:    "parse no match - badpullpolicy",
			toParse: "badpullpolicy",
			want:    v1.PullPolicy(""),
			wantOK:  false,
		},
		{
			name:    "parse no match - if not present",
			toParse: "if not present",
			want:    v1.PullPolicy(""),
			wantOK:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, gotOK := parseImagePolicy(v1.PullPolicy(tc.toParse))
			assert.Equal(t, tc.wantOK, gotOK)
			assert.Equal(t, tc.want, got)
		})
	}
}
