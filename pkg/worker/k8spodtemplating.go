package worker

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/iver-wharf/wharf-cmd/pkg/steps"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml"
	"github.com/iver-wharf/wharf-core/v2/pkg/env"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (f k8sStepRunnerFactory) getStepPodSpec(ctx context.Context, step wharfyml.Step) (v1.Pod, error) {
	podSpecer, ok := step.Type.(steps.PodSpecer)
	if !ok {
		return v1.Pod{}, errors.New("step type cannot produce a Kubernetes Pod specification")
	}

	annotations := map[string]string{
		"wharf.iver.com/project-id": "456", // TODO: Use real numbers
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
				"wharf.iver.com/build-ref":  "123", // TODO: Use real numbers
				"wharf.iver.com/project-id": "456",
				"wharf.iver.com/stage-id":   "789",
				"wharf.iver.com/step-id":    "789",
			},
			OwnerReferences: getOwnerReferences(),
		},
		Spec: podSpecer.PodSpec(),
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
	var enabled bool
	if err := env.Bind(&enabled, "WHARF_KUBERNETES_OWNER_ENABLE"); err != nil {
		log.Warn().WithError(err).Message("Failed binding WHARF_KUBERNETES_OWNER_ENABLE environment variables.")
		return nil
	}

	if !enabled {
		log.Debug().
			Message("Skipping Kubernetes OwnerReference because WHARF_KUBERNETES_OWNER_ENABLE was not set to 'true'.")
		return nil
	}

	var name, uid string
	if err := env.BindMultiple(map[*string]string{
		&name: "WHARF_KUBERNETES_OWNER_NAME",
		&uid:  "WHARF_KUBERNETES_OWNER_UID",
	}); err != nil {
		log.Warn().WithError(err).Message("Failed binding WHARF_KUBERNETES_OWNER_XXX environment variables.")
		return nil
	}

	log.Info().
		WithString("name", name).
		WithString("uid", uid).
		Message("Enabling Kubernetes OwnerReference.")

	True := true
	return []metav1.OwnerReference{
		{
			APIVersion:         "v1",
			Kind:               "Pod",
			Name:               name,
			UID:                types.UID(uid),
			BlockOwnerDeletion: &True,
			Controller:         &True,
		},
	}
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
