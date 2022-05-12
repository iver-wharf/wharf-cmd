package steps

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/config"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	v1 "k8s.io/api/core/v1"
)

// Docker represents a step type for building and pushing Docker images.
type Docker struct {
	// Required fields
	File string
	Tag  string

	// Optional fields
	Destination string
	Name        string
	Group       string
	Context     string
	Secret      string
	Registry    string
	AppendCert  bool
	Push        bool
	Args        []string
	SecretName  string
	SecretArgs  []string

	config  *config.DockerStepConfig
	podSpec *v1.PodSpec
}

// StepTypeName returns the name of this step type.
func (Docker) StepTypeName() string { return "docker" }

func (s Docker) PodSpec() *v1.PodSpec { return s.podSpec }

func (s Docker) init(stepName string, v visit.MapVisitor) (StepType, errutil.Slice) {
	s.Name = stepName
	s.Secret = "gitlab-registry"

	var errSlice errutil.Slice

	if !v.HasNode("destination") {
		var repoName string
		var errs errutil.Slice
		errs.Add(
			v.VisitStringWithVarSub("registry", "REG_URL", &s.Registry),
			v.VisitStringWithVarSub("group", "REPO_GROUP", &s.Registry),
			v.VisitStringFromVarSub("REPO_NAME", &repoName),
			v.VisitString("name", &s.Name), // Already defaults to step name
		)
		for _, err := range errs {
			errSlice.Add(fmt.Errorf(`eval "destination" default: %w`, err))
		}
		if repoName == s.Name {
			s.Destination = fmt.Sprintf("%s/%s/%s",
				s.Registry, s.Group, repoName)
		} else {
			s.Destination = fmt.Sprintf("%s/%s/%s/%s",
				s.Registry, s.Group, repoName, s.Name)
		}
	}

	if !v.HasNode("append-cert") {
		var repoGroup string
		err := v.VisitStringFromVarSub("REPO_GROUP", &repoGroup)
		if err != nil {
			errSlice.Add(fmt.Errorf(`eval "append-cert" default: %w`, err))
		}
		if strings.HasPrefix(strings.ToLower(s.Group), "default") {
			s.AppendCert = true
		}
	}

	// Visitling
	errSlice.Add(
		v.VisitString("file", &s.File),
		v.VisitString("tag", &s.Tag),
		v.VisitString("destination", &s.Destination),
		v.VisitString("name", &s.Name),
		v.VisitString("context", &s.Context),
		v.VisitString("secret", &s.Secret),
		v.VisitBool("append-cert", &s.AppendCert),
		v.VisitBool("push", &s.Push),
		v.VisitString("secretName", &s.SecretName),
	)
	errSlice.Add(v.VisitStringSlice("args", &s.Args)...)
	errSlice.Add(v.VisitStringSlice("secretArgs", &s.SecretArgs)...)

	// Validation
	errSlice.Add(
		v.ValidateRequiredString("file"),
		v.ValidateRequiredString("tag"),
	)

	podSpec, errs := s.applyStepDocker(stepName, v)
	s.podSpec = podSpec
	errSlice.Add(errs...)

	return s, errSlice
}

func (s Docker) applyStepDocker(stepName string, v visit.MapVisitor) (*v1.PodSpec, errutil.Slice) {
	var errSlice errutil.Slice
	podSpec := basePodSpec

	repoDir := commonRepoVolumeMount.MountPath
	cont := v1.Container{
		Name:  commonContainerName,
		Image: fmt.Sprintf("%s:%s", s.config.Image, s.config.ImageTag),
		// default entrypoint for image is "/kaniko/executor"
		WorkingDir: repoDir,
		VolumeMounts: []v1.VolumeMount{
			commonRepoVolumeMount,
		},
	}

	if s.AppendCert {
		cont.VolumeMounts = append(cont.VolumeMounts,
			v1.VolumeMount{
				Name:      "cert",
				ReadOnly:  true,
				MountPath: "/mnt/cert",
			})
		podSpec.Volumes = append(podSpec.Volumes, v1.Volume{
			Name: "cert",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	}

	// TODO: Mount Docker secrets from REG_SECRET built-in var

	// TODO: Add "--insecure" arg if REG_INSECURE

	args := []string{
		// Not using path/filepath package because we know don't want to
		// suddenly use Windows directory separator when running from Windows.
		"--dockerfile", path.Join(repoDir, s.File),
		"--context", path.Join(repoDir, s.Context),
		"--skip-tls-verify", // This is bad, but remains due to backward compatibility
	}

	for _, buildArg := range s.Args {
		args = append(args, "--build-arg", buildArg)
	}

	destination := s.getDockerDestination(stepName)
	anyTag := false
	for _, tag := range strings.Split(s.Tag, ",") {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		anyTag = true
		args = append(args, "--destination",
			fmt.Sprintf("%s:%s", destination, tag))
	}
	if !anyTag {
		errSlice.Add(errors.New("tags field resolved to zero tags"))
	}

	if !s.Push {
		args = append(args, "--no-push")
	}

	if s.SecretName != "" {
		// In Docker & Kaniko, adding only `--build-arg MY_ARG` will make it
		// pull the value from an environment variable instead of from a literal.
		// This is used to not specify the secret values in the pod manifest.

		secretName := fmt.Sprintf("wharf-%s-project-%d-secretname-%s",
			"local", // TODO: Use Wharf instance ID
			1,       // TODO: Use project ID
			s.SecretName,
		)
		optional := true
		for _, arg := range s.SecretArgs {
			argName, secretKey, hasCut := strings.Cut(arg, "=")
			if !hasCut {
				errSlice.Add(errors.New("invalid secret format: missing '=', expected 'ARG=secret-key'"))
				continue
			}
			if len(argName) == 0 {
				errSlice.Add(errors.New("invalid secret format: empty 'ARG', expected 'ARG=secret-key'"))
				continue
			}
			if len(secretKey) == 0 {
				errSlice.Add(errors.New("invalid secret format: empty 'secret-key', expected 'ARG=secret-key'"))
				continue
			}
			args = append(args, "--build-arg", argName)
			cont.Env = append(cont.Env, v1.EnvVar{
				Name: argName,
				ValueFrom: &v1.EnvVarSource{
					SecretKeyRef: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: secretName,
						},
						Optional: &optional,
					},
				},
			})
		}
	} else if len(s.SecretArgs) != 0 {
		errSlice.Add(errors.New("found secretArgs but is missing secretName"))
	}

	cont.Args = args
	podSpec.Containers = append(podSpec.Containers, cont)
	return &podSpec, errSlice
}

func (s Docker) getDockerDestination(stepName string) string {
	if s.Destination != "" {
		return strings.ToLower(s.Destination)
	}
	const repoName = "project-name" // TODO: replace with REPO_NAME built-in var
	if s.Registry == "" {
		s.Registry = "docker.io" // TODO: replace with REG_URL
	}
	if s.Group == "" {
		s.Group = "iver-wharf" // TODO: replace with REPO_GROUP
	}
	if stepName == repoName {
		return strings.ToLower(fmt.Sprintf("%s/%s/%s",
			s.Registry, s.Group, repoName))
	}
	return strings.ToLower(fmt.Sprintf("%s/%s/%s/%s",
		s.Registry, s.Group, repoName, stepName))
}
