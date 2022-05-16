package worker

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/iver-wharf/wharf-cmd/pkg/config"
	"github.com/iver-wharf/wharf-cmd/pkg/steps"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-core/pkg/env"
	"gopkg.in/typ.v4/slices"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	commonContainerName   = "step"
	commonRepoVolumeMount = v1.VolumeMount{
		Name:      "repo",
		MountPath: "/mnt/repo",
	}

	//go:embed k8sscript-nuget-package.sh
	nugetPackageScript string
)

func (f k8sStepRunnerFactory) getStepPodSpec(ctx context.Context, step wharfyml.Step) (v1.Pod, error) {
	podSpecer, ok := step.Type.(steps.PodSpecer)
	if !ok {
		return v1.Pod{}, errors.New("step type cannot produce a Kubernetes Pod specification")
	}
	podSpecPtr := podSpecer.PodSpec()
	var podSpec v1.PodSpec
	if podSpecPtr != nil {
		// TODO: Return error if nil, as all steps should return valid pod spec.
		podSpec = *podSpecPtr
	}

	annotations := map[string]string{
		"wharf.iver.com/project-id": "456",
		"wharf.iver.com/stage-id":   "789",
		"wharf.iver.com/step-id":    "789",
		"wharf.iver.com/step-name":  step.Name,
	}
	if stage, ok := contextStageName(ctx); ok {
		annotations["wharf.iver.com/stage-name"] = stage
	}
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: getPodGenerateName(step),
			Annotations:  annotations,
			Labels: map[string]string{
				"app":                          "wharf-cmd-worker-step",
				"app.kubernetes.io/name":       "wharf-cmd-worker-step",
				"app.kubernetes.io/part-of":    "wharf",
				"app.kubernetes.io/managed-by": "wharf-cmd-worker",
				"app.kubernetes.io/created-by": "wharf-cmd-worker",

				"wharf.iver.com/instance":   f.Config.InstanceID,
				"wharf.iver.com/build-ref":  "123",
				"wharf.iver.com/project-id": "456",
				"wharf.iver.com/stage-id":   "789",
				"wharf.iver.com/step-id":    "789",
			},
			OwnerReferences: getOwnerReferences(),
		},
		Spec: podSpec,
	}

	if podSpecPtr == nil {
		if err := applyStep(f.Config.Worker.Steps, &pod, step); err != nil {
			return v1.Pod{}, err
		}
	}

	if len(pod.Spec.Containers) == 0 {
		return v1.Pod{}, errors.New("step type did not add an app container")
	}

	return pod, nil
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

func getOwnerReferences() []metav1.OwnerReference {
	var (
		enabled   bool
		name, uid string
	)
	if err := env.BindMultiple(map[any]string{
		&enabled: "WHARF_KUBERNETES_OWNER_ENABLE",
		&name:    "WHARF_KUBERNETES_OWNER_NAME",
		&uid:     "WHARF_KUBERNETES_OWNER_UID",
	}); err != nil {
		log.Warn().WithError(err).Message("Failed binding WHARF_KUBERNETES_OWNER_XXX environment variables.")
		enabled = false
	}

	log.Debug().
		WithBool("enabled", enabled).
		WithString("name", name).
		WithString("uid", uid).
		Message("Environment variables from owner.")

	var ownerReferences []metav1.OwnerReference
	if enabled {
		True := true
		ownerReferences = append(ownerReferences, metav1.OwnerReference{
			APIVersion:         "v1",
			Kind:               "Pod",
			Name:               name,
			UID:                types.UID(uid),
			BlockOwnerDeletion: &True,
			Controller:         &True,
		})
	}
	return ownerReferences
}

func getOnlyFilesToTransfer(step wharfyml.Step) ([]string, bool) {
	switch s := step.Type.(type) {
	case steps.Helm:
		return s.Files, true
	case steps.Kubectl:
		if s.File != "" {
			return append(s.Files, s.File), true
		}
		return s.Files, true
	default:
		return nil, false
	}
}

func applyStep(c config.StepsConfig, pod *v1.Pod, step wharfyml.Step) error {
	switch s := step.Type.(type) {
	case steps.Helm:
		return applyStepHelm(c.Helm, pod, s)
	case steps.Kubectl:
		return applyStepKubectl(c.Kubectl, pod, s)
	case steps.NuGetPackage:
		return applyStepNuGetPackage(pod, s)
	case nil:
		return errors.New("nil step type")
	default:
		return fmt.Errorf("unknown step type: %q", s.StepTypeName())
	}
}

func applyStepHelm(config config.HelmStepConfig, pod *v1.Pod, step steps.Helm) error {
	cont := v1.Container{
		Name:       commonContainerName,
		Image:      fmt.Sprintf("%s:%s", config.Image, step.HelmVersion),
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
		step.Name,
		step.Chart,
		"--repo", step.Repo,
		"--namespace", step.Namespace,
	}

	if step.ChartVersion != "" {
		cmd = append(cmd, "--version", step.ChartVersion)
	}

	for _, file := range step.Files {
		cmd = append(cmd, "--values", file)
	}

	// TODO: Add chart repo credentials from REG_USER & REG_PASS if set
	// TODO: Also make sure to censor them, so their values don't get logged.

	log.Debug().WithString("args", quoteArgsForLogging(cmd)).
		Message("Helm args.")

	cont.Command = cmd
	pod.Spec.Containers = append(pod.Spec.Containers, cont)
	pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
		Name: "kubeconfig",
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: step.Cluster,
				},
			},
		},
	})
	return nil
}

func applyStepKubectl(config config.KubectlStepConfig, pod *v1.Pod, step steps.Kubectl) error {
	cont := v1.Container{
		Name:       commonContainerName,
		Image:      fmt.Sprintf("%s:%s", config.Image, config.ImageTag),
		WorkingDir: commonRepoVolumeMount.MountPath,
		VolumeMounts: []v1.VolumeMount{
			commonRepoVolumeMount,
			{Name: "kubeconfig", MountPath: "/root/.kube"},
		},
	}

	cmd := []string{
		"kubectl",
		step.Action,
	}

	if step.Namespace != "" {
		cmd = append(cmd, "--namespace", step.Namespace)
	}

	files := step.Files
	if step.File != "" {
		files = append(files, step.File)
	}

	for _, file := range files {
		cmd = append(cmd, "--filename", file)
	}

	if step.Force {
		cmd = append(cmd, "--force")
	}

	log.Debug().WithString("args", quoteArgsForLogging(cmd)).
		Message("Kubectl args.")

	cont.Command = cmd
	pod.Spec.Containers = append(pod.Spec.Containers, cont)
	pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
		Name: "kubeconfig",
		VolumeSource: v1.VolumeSource{
			ConfigMap: &v1.ConfigMapVolumeSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: step.Cluster,
				},
			},
		},
	})
	return nil
}

func applyStepNuGetPackage(pod *v1.Pod, step steps.NuGetPackage) error {
	cont := v1.Container{
		Name:       commonContainerName,
		Image:      "mcr.microsoft.com/dotnet/sdk:3.1-alpine",
		WorkingDir: commonRepoVolumeMount.MountPath,
		VolumeMounts: []v1.VolumeMount{
			commonRepoVolumeMount,
		},
		Env: []v1.EnvVar{
			{
				Name: "NUGET_TOKEN",
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "wharf-nuget-api-token",
						},
						Key: "token",
					},
				},
			},
			{Name: "NUGET_REPO", Value: step.Repo},
			{Name: "NUGET_PROJECT_PATH", Value: step.ProjectPath},
			{Name: "NUGET_VERSION", Value: step.Version},
			{Name: "NUGET_SKIP_DUP", Value: boolString(step.SkipDuplicate)},
		},
		Command: []string{"/bin/bash", "-c"},
		Args:    []string{nugetPackageScript},
	}

	pod.Spec.Containers = append(pod.Spec.Containers, cont)
	return nil
}

func quoteArgsForLogging(args []string) string {
	argsQuoted := slices.Clone(args)
	for i, arg := range args {
		if strings.ContainsAny(arg, `"\' `) {
			argsQuoted[i] = fmt.Sprintf("%q", arg)
		}
	}
	return strings.Join(argsQuoted, " ")
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
