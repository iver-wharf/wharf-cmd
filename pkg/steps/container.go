package steps

import (
	"fmt"
	"strings"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	v1 "k8s.io/api/core/v1"
)

// Container represents a step type for running commands inside a Docker
// container.
type Container struct {
	// Required fields
	Image string
	Cmds  []string

	// Optional fields
	OS                    string
	Shell                 string
	SecretName            string
	ServiceAccount        string
	CertificatesMountPath string

	instanceID string
	podSpec    v1.PodSpec
}

// StepTypeName returns the name of this step type.
func (Container) StepTypeName() string { return "container" }

// PodSpec returns this step's Kubernetes Pod specification. Meant to be used
// by the wharf-cmd-worker when creating the actual pods.
func (s Container) PodSpec() v1.PodSpec { return s.podSpec }

func (s Container) init(_ string, v visit.MapVisitor) (StepType, errutil.Slice) {
	s.OS = "linux"
	s.Shell = "/bin/sh"
	s.ServiceAccount = "default"

	var errSlice errutil.Slice

	// Visiting
	errSlice.Add(
		v.VisitString("image", &s.Image),
		v.VisitString("os", &s.OS),
		v.VisitString("shell", &s.Shell),
		v.VisitString("secretName", &s.SecretName),
		v.VisitString("serviceAccount", &s.ServiceAccount),
		v.VisitString("certificatesMountPath", &s.CertificatesMountPath),
	)
	errSlice.Add(v.VisitStringSlice("cmds", &s.Cmds)...)

	// Validation
	errSlice.Add(
		v.ValidateRequiredString("image"),
		v.ValidateRequiredSlice("cmds"),
	)

	s.podSpec = s.applyStep()

	return s, errSlice
}

func (s Container) applyStep() v1.PodSpec {
	podSpec := newBasePodSpec()

	var cmds []string
	if s.OS == "windows" && s.Shell == "/bin/sh" {
		cmds = []string{"powershell.exe", "-C"}
	} else {
		cmds = []string{s.Shell, "-c"}
	}

	cont := v1.Container{
		Name:            commonContainerName,
		Image:           s.Image,
		ImagePullPolicy: v1.PullAlways,
		Command:         cmds,
		Args:            []string{strings.Join(s.Cmds, "\n")},
		WorkingDir:      commonRepoVolumeMount.MountPath,
		VolumeMounts: []v1.VolumeMount{
			commonRepoVolumeMount,
		},
	}

	if s.CertificatesMountPath != "" {
		podSpec.Volumes = append(podSpec.Volumes, v1.Volume{
			Name: "certificates",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "ca-certificates-config",
					},
				},
			},
		})
		cont.VolumeMounts = append(cont.VolumeMounts, v1.VolumeMount{
			Name:      "certificates",
			MountPath: s.CertificatesMountPath,
		})
	}

	if s.SecretName != "" {
		secretName := fmt.Sprintf("wharf-%s-project-%d-secretname-%s",
			s.instanceID,
			1, // TODO: Use project ID
			s.SecretName,
		)
		cont.EnvFrom = append(cont.EnvFrom, v1.EnvFromSource{
			SecretRef: &v1.SecretEnvSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: secretName,
				},
			},
		})
	}

	podSpec.ServiceAccountName = s.ServiceAccount
	podSpec.Containers = append(podSpec.Containers, cont)
	return podSpec
}
