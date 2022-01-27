package builder

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/iver-wharf/wharf-cmd/pkg/core/wharfyml"
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
				strings.ToLower(step.Type.String()),
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
	switch step.Type {
	case wharfyml.Container:
		image, ok := step.Variables["image"]
		if !ok {
			return "", errors.New("missing required field: image")
		}
		imageStr, ok := image.(string)
		if !ok {
			return "", fmt.Errorf("invalid field type: image: want string, got: %T", image)
		}
		return imageStr, nil
	default:
		return "", fmt.Errorf("unsupported step type: %q", step.Type)
	}
}

func getPodCommandArgs(step wharfyml.Step) (cmds, args []string, err error) {
	switch step.Type {
	case wharfyml.Container:
		cmdsAny, ok := step.Variables["cmds"]
		if !ok {
			return nil, nil, errors.New("missing required field: cmds")
		}
		cmds, err := convStepFieldToStrings("cmds", cmdsAny)
		if err != nil {
			return nil, nil, err
		}
		args := strings.Join(cmds, "\n")
		shell := "/bin/sh"
		if shellAny, ok := step.Variables["shell"]; ok {
			shell, ok = shellAny.(string)
			if !ok {
				return nil, nil, fmt.Errorf("invalid field type: shell: want string, got %T", shellAny)
			}
		}
		return []string{shell, "-c"}, []string{args}, nil
	default:
		return nil, nil, fmt.Errorf("unsupported step type: %q", step.Type)
	}
}

func convStepFieldToStrings(fieldName string, value interface{}) ([]string, error) {
	anyArr, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid field type: %s: want string array, got: %T", fieldName, value)
	}
	strs := make([]string, 0, len(anyArr))
	for i, v := range anyArr {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("invalid field type: %s: index %d: want string, got: %T", fieldName, i, value)
		}
		strs = append(strs, str)
	}
	return strs, nil
}
