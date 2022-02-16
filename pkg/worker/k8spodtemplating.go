package worker

import (
	"context"
	"fmt"
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
			GenerateName: fmt.Sprintf("wharf-build-%s-%s-",
				strings.ToLower(step.Type.StepTypeName()),
				strings.ToLower(step.Name)),
			Annotations: annotations,
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
