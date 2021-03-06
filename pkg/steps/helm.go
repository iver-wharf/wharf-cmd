package steps

import (
	"fmt"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/config"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	v1 "k8s.io/api/core/v1"
)

// Helm represents a step type for installing a Helm chart into a Kubernetes
// cluster.
type Helm struct {
	// Required fields
	Chart     string
	Name      string
	Namespace string

	// Optional fields
	Repo         string
	Set          map[string]string
	Files        []string
	ChartVersion string
	HelmVersion  string
	Cluster      string
	Secret       string

	config  *config.HelmStepConfig
	podSpec v1.PodSpec
}

// StepTypeName returns the name of this step type.
func (Helm) StepTypeName() string { return "helm" }

// PodSpec returns this step's Kubernetes Pod specification. Meant to be used
// by the wharf-cmd-worker when creating the actual pods.
func (s Helm) PodSpec() v1.PodSpec { return s.podSpec }

func (s Helm) init(_ string, v visit.MapVisitor) (StepType, errutil.Slice) {
	s.Cluster = "kubectl-config"
	s.HelmVersion = "v2.14.1"

	var errSlice errutil.Slice

	if !v.HasNode("repo") {
		var chartRepo string
		var repoGroup string
		var errs errutil.Slice
		errs.Add(
			v.RequireStringFromVarSub("CHART_REPO", &chartRepo),
			v.RequireStringFromVarSub("REPO_GROUP", &repoGroup),
		)
		for _, err := range errs {
			errSlice.Add(fmt.Errorf(`eval "repo" default: %w`, err))
		}
		s.Repo = fmt.Sprintf("%s/%s", chartRepo, repoGroup)
	}

	if v.HasNode("secret") {
		errSlice.Add(v.VisitString("secret", &s.Secret))
	} else {
		errSlice.Add(v.LookupStringFromVarSub("HELM_REG_SECRET", &s.Secret))
	}

	// Visiting
	errSlice.Add(
		v.VisitString("chart", &s.Chart),
		v.VisitString("name", &s.Name),
		v.VisitString("namespace", &s.Namespace),
		v.VisitString("repo", &s.Repo),
		v.VisitString("chartVersion", &s.ChartVersion),
		v.VisitString("helmVersion", &s.HelmVersion),
		v.VisitString("cluster", &s.Cluster),
	)
	errSlice.Add(v.VisitStringStringMap("set", &s.Set)...)
	errSlice.Add(v.VisitStringSlice("files", &s.Files)...)
	if s.Repo == "stage" {
		s.Repo = "https://kubernetes-charts.storage.googleapis.com"
	}

	// Validation
	errSlice.Add(
		v.ValidateRequiredString("chart"),
		v.ValidateRequiredString("name"),
		v.ValidateRequiredString("namespace"),
	)

	s.podSpec = s.applyStep()

	return s, errSlice
}

func (s Helm) applyStep() v1.PodSpec {
	podSpec := newBasePodSpec()

	cont := v1.Container{
		Name:       commonContainerName,
		Image:      fmt.Sprintf("%s:%s", s.config.Image, s.HelmVersion),
		WorkingDir: commonRepoVolumeMount.MountPath,
		VolumeMounts: []v1.VolumeMount{
			commonRepoVolumeMount,
			{Name: "kubeconfig", MountPath: "/root/.kube"},
		},
	}

	cmd := []string{
		"helm",
		"upgrade",
		"--install",
		s.Name,
		s.Chart,
		"--repo", s.Repo,
		"--namespace", s.Namespace,
	}

	if s.ChartVersion != "" {
		cmd = append(cmd, "--version", s.ChartVersion)
	}

	for _, file := range s.Files {
		cmd = append(cmd, "--values", file)
	}

	cont.Command = cmd

	if s.Secret != "" {
		addHelmSecretVolume(s.Secret, &podSpec, &cont)
	}

	podSpec.Containers = append(podSpec.Containers, cont)
	podSpec.Volumes = append(podSpec.Volumes, v1.Volume{
		Name: "kubeconfig",
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: s.Cluster,
				},
			},
		},
	})

	return podSpec
}

func addHelmSecretVolume(secretName string, podSpec *v1.PodSpec, cont *v1.Container) {
	const volumeName = "helm-secrets"
	podSpec.Volumes = append(podSpec.Volumes, v1.Volume{
		Name: volumeName,
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: secretName,
				Items: []v1.KeyToPath{
					{
						Key:  "config.json",
						Path: "config.json",
					},
				},
			},
		},
	})
	cont.VolumeMounts = append(cont.VolumeMounts, v1.VolumeMount{
		Name:      volumeName,
		MountPath: "/etc/helm/registry",
	})
	// Override the config path. If the user uses a rootless image, then
	// the path wont be at the otherwise default location
	//   /root/.config/helm/registry/config.json
	cont.Env = append(cont.Env, v1.EnvVar{
		Name:  "HELM_REGISTRY_CONFIG",
		Value: "/etc/helm/registry/config.json",
	})
}
