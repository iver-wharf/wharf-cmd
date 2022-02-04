package worker

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getPodSpec(ctx context.Context, step wharfyml.Step) (v1.Pod, error) {
	image, err := getPodImage(step)
	if err != nil {
		return v1.Pod{}, err
	}
	cmds, args, err := getPodCommandArgs(step)
	if err != nil {
		return v1.Pod{}, err
	}
	annotations := map[string]string{
		"wharf.iver.com/step": step.Name,
	}
	if stage, ok := contextStageName(ctx); ok {
		annotations["wharf.iver.com/stage"] = stage
	}
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: getPodGenerateName(step),
			Annotations:  annotations,
			Labels: map[string]string{
				"wharf.iver.com/build": "true",
			},
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			InitContainers: []v1.Container{
				{
					Name:            "init",
					Image:           "alpine:3",
					ImagePullPolicy: v1.PullIfNotPresent,
					Command:         podInitWaitArgs,
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "repo",
							MountPath: "/mnt/repo",
						},
					},
				},
			},
			Containers: []v1.Container{
				{
					Name:            "step",
					Image:           image,
					ImagePullPolicy: v1.PullAlways,
					Command:         cmds,
					Args:            args,
					WorkingDir:      "/mnt/repo",
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "repo",
							MountPath: "/mnt/repo",
						},
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "repo",
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}, nil
}

func getPodGenerateName(step wharfyml.Step) string {
	name := fmt.Sprintf("wharf-build-%s-%s-",
		sanitizePodName(step.Type.StepTypeName()),
		sanitizePodName(step.Name))
	// Kubernetes API will respond with error if the GenerateName is too long.
	// We trim it here to less than the 253 char limit as 253 is an excessive
	// name length.
	const maxLen = 42 // jokes aside, 42 is actually a great maximum name length
	// For reference, this is what a 42-long name looks like:
	// wharf-build-container-some-long-step-name-
	if len(name) > maxLen {
		name = name[:maxLen-1] + "-"
	}
	return name
}

var regexInvalidDNSSubdomainChars = regexp.MustCompile(`[^a-z0-9-]`)

func sanitizePodName(name string) string {
	// Pods names must be valid DNS Subdomain names (IETF RFC-1123), meaning:
	// - max 253 chars long
	// - only lowercase alphanumeric or '-'
	// - must start and end with alphanumeric char
	// https://kubernetes.io/docs/concepts/workloads/pods/#working-with-pods
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-subdomain-names
	name = strings.ToLower(name)
	name = regexInvalidDNSSubdomainChars.ReplaceAllLiteralString(name, "-")
	name = strings.Trim(name, "-")
	return name
}

func getPodImage(step wharfyml.Step) (string, error) {
	switch s := step.Type.(type) {
	case wharfyml.StepContainer:
		return s.Image, nil
	default:
		return "", fmt.Errorf("unsupported step type: %q", step.Type)
	}
}

func getPodCommandArgs(step wharfyml.Step) (cmds, args []string, err error) {
	switch s := step.Type.(type) {
	case wharfyml.StepContainer:
		args := strings.Join(s.Cmds, "\n")
		return []string{s.Shell, "-c"}, []string{args}, nil
	default:
		return nil, nil, fmt.Errorf("unsupported step type: %q", step.Type)
	}
}
