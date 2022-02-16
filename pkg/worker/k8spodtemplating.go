package worker

import (
	"context"
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	commonRepoVolumeMount = v1.VolumeMount{
		Name:      "repo",
		MountPath: "/mnt/repo",
	}
)

func getPodSpec(ctx context.Context, step wharfyml.Step) (v1.Pod, error) {
	annotations := map[string]string{
		"wharf.iver.com/step": step.Name,
	}
	if stage, ok := contextStageName(ctx); ok {
		annotations["wharf.iver.com/stage"] = stage
	}
	pod := v1.Pod{
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
			Volumes: []v1.Volume{
				{
					Name: "repo",
					VolumeSource: v1.VolumeSource{
						EmptyDir: &v1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}

	if err := applyStep(&pod, step); err != nil {
		return v1.Pod{}, err
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

func applyStep(pod *v1.Pod, step wharfyml.Step) error {
	switch s := step.Type.(type) {
	case wharfyml.StepContainer:
		return applyStepContainer(pod, s)
	case wharfyml.StepDocker:
		return applyStepDocker(pod, s, step.Name)
	case wharfyml.StepHelmPackage:
		return applyStepHelmPackage(pod, s)
	case nil:
		return errors.New("nil step type")
	default:
		return fmt.Errorf("unknown step type: %q", s.StepTypeName())
	}
}

func applyStepContainer(pod *v1.Pod, step wharfyml.StepContainer) error {
	var cmds []string

	if step.OS == "windows" && step.Shell == "/bin/sh" {
		cmds = []string{"powershell.exe", "-C"}
	} else {
		cmds = []string{step.Shell, "-c"}
	}

	args := []string{strings.Join(step.Cmds, "\n")}
	cont := v1.Container{
		Name:            "step",
		Image:           step.Image,
		ImagePullPolicy: v1.PullAlways,
		Command:         cmds,
		Args:            args,
		WorkingDir:      commonRepoVolumeMount.MountPath,
		VolumeMounts: []v1.VolumeMount{
			commonRepoVolumeMount,
		},
	}

	pod.Spec.ServiceAccountName = step.ServiceAccount

	if step.CertificatesMountPath != "" {
		pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
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
			MountPath: step.CertificatesMountPath,
		})
	}

	if step.SecretName != "" {
		secretName := fmt.Sprintf("wharf-%s-project%d-secretname-%s",
			"local", // TODO: Use Wharf instance ID
			1,       // TODO: Use project ID
			step.SecretName,
		)
		optional := true
		cont.EnvFrom = append(cont.EnvFrom, v1.EnvFromSource{
			SecretRef: &v1.SecretEnvSource{
				LocalObjectReference: v1.LocalObjectReference{
					Name: secretName,
				},
				Optional: &optional,
			},
		})
	}

	pod.Spec.Containers = append(pod.Spec.Containers, cont)
	return nil
}

func applyStepDocker(pod *v1.Pod, step wharfyml.StepDocker, stepName string) error {
	repoDir := commonRepoVolumeMount.MountPath
	cont := v1.Container{
		Name:  "step",
		Image: "boolman/kaniko:busybox-2020-01-15",
		// default entrypoint for image is "/kaniko/executor"
		WorkingDir: repoDir,
		VolumeMounts: []v1.VolumeMount{
			commonRepoVolumeMount,
		},
	}

	// TODO: Load in certificates somehow

	// TODO: Mount Docker secrets from REG_SECRET built-in var

	// TODO: Add "--insecure" arg if REG_INSECURE

	args := []string{
		// Not using path/filepath package because we know don't want to
		// suddenly use Windows directory separator when running from Windows.
		"--dockerfile", path.Join(repoDir, step.File),
		"--context", path.Join(repoDir, step.Context),
		"--skip-tls-verify", // This is bad, but remains due to backward compatibility
	}

	for _, buildArg := range step.Args {
		args = append(args, "--build-arg", buildArg)
	}

	destination := getDockerDestination(step, stepName)
	anyTag := false
	for _, tag := range strings.Split(step.Tag, ",") {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		anyTag = true
		args = append(args, "--destination",
			fmt.Sprintf("%s:%s", destination, tag))
	}
	if !anyTag {
		return errors.New("tags field resolved to zero tags")
	}

	if !step.Push {
		args = append(args, "--no-push")
	}

	cont.Args = args

	argsQuoted := make([]string, len(args))
	copy(argsQuoted, args)
	for i, arg := range args {
		if strings.ContainsAny(arg, `"\' `) {
			argsQuoted[i] = fmt.Sprintf("%q", arg)
		}
	}
	log.Debug().WithStringf("args", strings.Join(argsQuoted, " ")).
		Message("Kaniko args.")

	pod.Spec.Containers = append(pod.Spec.Containers, cont)
	return nil
}

func getDockerDestination(step wharfyml.StepDocker, stepName string) string {
	if step.Destination != "" {
		return strings.ToLower(step.Destination)
	}
	const repoName = "project-name" // TODO: replace with REPO_NAME built-in var
	if step.Registry == "" {
		step.Registry = "docker.io" // TODO: replace with REG_URL
	}
	if step.Group == "" {
		step.Group = "iver-wharf" // TODO: replace with REPO_GROUP
	}
	if stepName == repoName {
		return strings.ToLower(fmt.Sprintf("%s/%s/%s",
			step.Registry, step.Group, repoName))
	}
	return strings.ToLower(fmt.Sprintf("%s/%s/%s/%s",
		step.Registry, step.Group, repoName, stepName))
}

func applyStepHelmPackage(pod *v1.Pod, step wharfyml.StepHelmPackage) error {
	destination := "https://harbor.local/chartrepo/my-group" // TODO: replace with CHART_REPO/REPO_GROUP
	if step.Destination != "" {
		destination = step.Destination
	}

	var versionFlag string
	if step.Version != "" {
		versionFlag = "--version=" + step.Version
	}

	cont := v1.Container{
		Name:  "step",
		Image: "wharfse/helm:v3.5.4",
		// default entrypoint for image is "/kaniko/executor"
		WorkingDir: commonRepoVolumeMount.MountPath,
		VolumeMounts: []v1.VolumeMount{
			commonRepoVolumeMount,
		},
		Command: []string{"/bin/sh", "-c"},
		Env: []v1.EnvVar{
			{Name: "CHART_PATH", Value: step.ChartPath},
			{Name: "CHART_REPO", Value: destination},
			{Name: "VERSION_FLAG", Value: versionFlag},
			{Name: "REG_USER", Value: "admin"},    // TODO: replace with REG_USER
			{Name: "REG_PASS", Value: "changeit"}, // TODO: replace with REG_PASS
		},
	}

	cont.Args = []string{`
echo "\$ helm package $CHART_PATH $VERSION_FLAG"
helm package "$CHART_PATH" "$VERSION_FLAG"

echo "\$ helm push *.tgz $CHART_REPO --insecure --username \$REG_USER --password \$REG_PASS"
helm push *.tgz "$CHART_REPO" --insecure --username "$REG_USER" --password "$REG_PASS"
`}

	pod.Spec.Containers = append(pod.Spec.Containers, cont)
	return nil
}
