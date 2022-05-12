package steps

import (
	"fmt"
	"strings"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
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
}

// StepTypeName returns the name of this step type.
func (Docker) StepTypeName() string { return "docker" }

func (s *Docker) init(stepName string, v visit.MapVisitor) errutil.Slice {
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
	return errSlice
}
