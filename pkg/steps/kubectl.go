package steps

import (
	"fmt"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/config"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	v1 "k8s.io/api/core/v1"
)

// Kubectl represents a step type for running kubectl commands on some
// Kubernetes manifest files.
type Kubectl struct {
	// Required fields
	File  string
	Files []string

	// Optional fields
	Namespace string
	Action    string
	Force     bool
	Cluster   string

	config  *config.KubectlStepConfig
	podSpec v1.PodSpec
}

// StepTypeName returns the name of this step type.
func (Kubectl) StepTypeName() string { return "kubectl" }

// PodSpec returns this step's Kubernetes Pod specification. Meant to be used
// by the wharf-cmd-worker when creating the actual pods.
func (s Kubectl) PodSpec() v1.PodSpec { return s.podSpec }

func (s Kubectl) init(_ string, v visit.MapVisitor) (StepType, errutil.Slice) {
	s.Cluster = "kubectl-config"
	s.Action = "apply"

	var errSlice errutil.Slice

	// Visiting
	errSlice.Add(
		v.VisitString("file", &s.File),
		v.VisitString("namespace", &s.Namespace),
		v.VisitString("action", &s.Action),
		v.VisitBool("force", &s.Force),
		v.VisitString("cluster", &s.Cluster),
	)
	errSlice.Add(v.VisitStringSlice("files", &s.Files)...)

	// Validation
	if len(s.Files) == 0 {
		// Only either file or files is required
		errSlice.Add(v.ValidateRequiredString("file"))
	}

	s.podSpec = s.applyStep()

	return s, errSlice
}

func (s Kubectl) applyStep() v1.PodSpec {
	podSpec := newBasePodSpec()

	cont := v1.Container{
		Name:       commonContainerName,
		Image:      fmt.Sprintf("%s:%s", s.config.Image, s.config.ImageTag),
		WorkingDir: commonRepoVolumeMount.MountPath,
		VolumeMounts: []v1.VolumeMount{
			commonRepoVolumeMount,
			{Name: "kubeconfig", MountPath: "/root/.kube"},
		},
	}

	cmd := []string{
		"kubectl",
		s.Action,
	}

	if s.Namespace != "" {
		cmd = append(cmd, "--namespace", s.Namespace)
	}

	files := s.Files
	if s.File != "" {
		files = append(files, s.File)
	}

	for _, file := range files {
		cmd = append(cmd, "--filename", file)
	}

	if s.Force {
		cmd = append(cmd, "--force")
	}

	cont.Command = cmd
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
