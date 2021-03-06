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

const (
	dockerFieldFile        = "file"
	dockerFieldTag         = "tag"
	dockerFieldDestination = "destination"
	dockerFieldName        = "name"
	dockerFieldGroup       = "group"
	dockerFieldContext     = "context"
	dockerFieldSecret      = "secret"
	dockerFieldRegistry    = "registry"
	dockerFieldAppendCert  = "append-cert"
	dockerFieldPush        = "push"
	dockerFieldArgs        = "args"
	dockerFieldSecretName  = "secretName"
	dockerFieldSecretArgs  = "secretArgs"
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

	instanceID string
	config     *config.DockerStepConfig
	podSpec    v1.PodSpec
}

// StepTypeName returns the name of this step type.
func (Docker) StepTypeName() string { return "docker" }

// PodSpec returns this step's Kubernetes Pod specification. Meant to be used
// by the wharf-cmd-worker when creating the actual pods.
func (s Docker) PodSpec() v1.PodSpec { return s.podSpec }

func (s Docker) init(stepName string, v visit.MapVisitor) (StepType, errutil.Slice) {
	s.Name = stepName
	s.Secret = "gitlab-registry"
	s.Push = true

	var errSlice errutil.Slice

	if !v.HasNode(dockerFieldDestination) {
		var repoName string
		var errs errutil.Slice
		errs.Add(
			v.VisitStringWithVarSub(dockerFieldRegistry, "REG_URL", &s.Registry),
			v.VisitStringWithVarSub(dockerFieldGroup, "REPO_GROUP", &s.Group),
			v.RequireStringFromVarSub("REPO_NAME", &repoName),
			v.VisitString(dockerFieldName, &s.Name), // Already defaults to step name
		)
		for _, err := range errs {
			errSlice.Add(fmt.Errorf(`eval '%s' default: %w`, dockerFieldDestination, err))
		}
		if repoName == s.Name {
			s.Destination = fmt.Sprintf("%s/%s/%s",
				s.Registry, s.Group, repoName)
		} else {
			s.Destination = fmt.Sprintf("%s/%s/%s/%s",
				s.Registry, s.Group, repoName, s.Name)
		}
	}

	if !v.HasNode(dockerFieldAppendCert) {
		var repoGroup string
		err := v.LookupStringFromVarSub("REPO_GROUP", &repoGroup)
		if err != nil {
			errSlice.Add(fmt.Errorf(`eval '%s' default: %w`, dockerFieldAppendCert, err))
		}
		if strings.HasPrefix(strings.ToLower(s.Group), "default") {
			s.AppendCert = true
		}
	}

	if v.HasNode(dockerFieldSecret) {
		errSlice.Add(v.VisitString(dockerFieldSecret, &s.Secret))
	} else {
		errSlice.Add(v.LookupStringFromVarSub("REG_SECRET", &s.Secret))
	}

	// Visiting
	errSlice.Add(
		v.VisitString(dockerFieldFile, &s.File),
		v.VisitString(dockerFieldTag, &s.Tag),
		v.VisitString(dockerFieldDestination, &s.Destination),
		v.VisitString(dockerFieldName, &s.Name),
		v.VisitString(dockerFieldContext, &s.Context),
		v.VisitBool(dockerFieldAppendCert, &s.AppendCert),
		v.VisitBool(dockerFieldPush, &s.Push),
		v.VisitString(dockerFieldSecretName, &s.SecretName),
	)
	errSlice.Add(v.VisitStringSlice(dockerFieldArgs, &s.Args)...)
	errSlice.Add(v.VisitStringSlice(dockerFieldSecretArgs, &s.SecretArgs)...)

	// Validation
	errSlice.Add(
		v.ValidateRequiredString(dockerFieldFile),
		v.ValidateRequiredString(dockerFieldTag),
	)

	for _, arg := range s.SecretArgs {
		argName, secretKey, hasCut := strings.Cut(arg, "=")
		if !hasCut {
			v.AddErrorFor(dockerFieldSecretArgs, &errSlice,
				fmt.Errorf("invalid secret format: missing '=', expected 'ARG=secret-key': %q", arg))
			continue
		}
		if len(argName) == 0 {
			v.AddErrorFor(dockerFieldSecretArgs, &errSlice,
				fmt.Errorf("invalid secret format: empty 'ARG', expected 'ARG=secret-key': %q", arg))
			continue
		}
		if len(secretKey) == 0 {
			v.AddErrorFor(dockerFieldSecretArgs, &errSlice,
				fmt.Errorf("invalid secret format: empty 'secret-key', expected 'ARG=secret-key': %q", arg))
		}
	}
	if len(s.SecretArgs) != 0 && s.SecretName == "" {
		v.AddErrorFor(dockerFieldSecretArgs, &errSlice,
			fmt.Errorf("found %s but is missing %s", dockerFieldSecretArgs, dockerFieldSecretName))
	} else if len(s.SecretArgs) == 0 && s.SecretName != "" {
		v.AddErrorFor(dockerFieldSecretName, &errSlice,
			fmt.Errorf("found %s but is missing %s", dockerFieldSecretName, dockerFieldSecretArgs))
	}

	podSpec, errs := s.applyStep(v)
	s.podSpec = podSpec
	errSlice.Add(errs...)

	return s, errSlice
}

func (s Docker) applyStep(v visit.MapVisitor) (v1.PodSpec, errutil.Slice) {
	var errSlice errutil.Slice
	podSpec := newBasePodSpec()

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
		podSpec.Volumes = append(podSpec.Volumes, v1.Volume{
			Name: "cert",
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
		cont.VolumeMounts = append(cont.VolumeMounts, v1.VolumeMount{
			Name:      "cert",
			ReadOnly:  true,
			MountPath: "/mnt/cert",
		})
	}

	if s.Secret != "" {
		const volumeName = "docker-secrets"
		podSpec.Volumes = append(podSpec.Volumes, v1.Volume{
			Name: volumeName,
			VolumeSource: v1.VolumeSource{
				Projected: &v1.ProjectedVolumeSource{
					Sources: []v1.VolumeProjection{
						{
							Secret: &v1.SecretProjection{
								LocalObjectReference: v1.LocalObjectReference{
									Name: s.Secret,
								},
								Items: []v1.KeyToPath{
									{
										Key:  ".dockerconfigjson",
										Path: "config.json",
									},
								},
							},
						},
					},
				},
			},
		})
		cont.VolumeMounts = append(cont.VolumeMounts, v1.VolumeMount{
			Name:      volumeName,
			MountPath: "/kaniko/.docker",
		})
	}

	args := []string{
		// Not using path/filepath package because we know don't want to
		// suddenly use Windows directory separator when running from Windows.
		"--dockerfile", path.Join(repoDir, s.File),
		"--context", path.Join(repoDir, s.Context),
		// TODO:
		//   --skip-tls-verify is bad, but remains due to backward compatibility.
		//   Would be nice to make optional/phase it out.
		"--skip-tls-verify",
	}

	var registryInsecure bool
	errSlice.Add(v.LookupBoolFromVarSub("REG_INSECURE", &registryInsecure))
	if registryInsecure {
		args = append(args, "--insecure")
	}

	for _, buildArg := range s.Args {
		args = append(args, "--build-arg", buildArg)
	}

	anyTag := false
	for _, tag := range strings.Split(s.Tag, ",") {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		anyTag = true
		args = append(args, "--destination",
			fmt.Sprintf("%s:%s", s.Destination, tag))
	}
	if !anyTag {
		v.AddErrorFor(dockerFieldTag, &errSlice, errors.New("field resolved to zero tags"))
	}

	if !s.Push {
		args = append(args, "--no-push")
	}

	if s.SecretName != "" {
		// In Docker & Kaniko, adding only `--build-arg MY_ARG` will make it
		// pull the value from an environment variable instead of from a literal.
		// This is used to not specify the secret values in the pod manifest.
		var projectID uint
		if err := v.RequireUintFromVarSub("PROJECT_ID", &projectID); err != nil {
			v.AddErrorFor(dockerFieldSecretName, &errSlice, err)
		}
		secretName := fmt.Sprintf("wharf-%s-project-%d-secretname-%s",
			s.instanceID,
			projectID,
			s.SecretName,
		)
		for _, arg := range s.SecretArgs {
			argName, secretKey, hasCut := strings.Cut(arg, "=")
			if !hasCut || len(argName) == 0 || len(secretKey) == 0 {
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
						Key: secretKey,
					},
				},
			})
		}
	}

	cont.Args = args
	podSpec.Containers = append(podSpec.Containers, cont)
	return podSpec, errSlice
}
