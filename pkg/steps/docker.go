package steps

import (
	"fmt"
	"strings"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/varsub"
)

// StepDocker represents a step type for building and pushing Docker images.
type StepDocker struct {
	// Step type metadata
	Meta StepTypeMeta

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

// Name returns the name of this step type.
func (StepDocker) Name() string { return "docker" }

func (s StepDocker) visitStepTypeNode(stepName string, p nodeMapParser, source varsub.Source) (StepType, errutil.Slice) {
	s.Meta = getStepTypeMeta(p, stepName)

	s.Name = stepName
	s.Secret = "gitlab-registry"

	var errSlice errutil.Slice

	if !p.hasNode("destination") {
		var repoName string
		var errs errutil.Slice
		errs.Add(
			p.unmarshalStringWithVarSub("registry", "REG_URL", source, &s.Registry),
			p.unmarshalStringWithVarSub("group", "REPO_GROUP", source, &s.Registry),
			p.unmarshalStringFromVarSub("REPO_NAME", source, &repoName),
			p.unmarshalString("name", &s.Name), // Already defaults to step name
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

	if !p.hasNode("append-cert") {
		var repoGroup string
		err := p.unmarshalStringFromVarSub("REPO_GROUP", source, &repoGroup)
		if err != nil {
			errSlice.Add(fmt.Errorf(`eval "append-cert" default: %w`, err))
		}
		if strings.HasPrefix(strings.ToLower(s.Group), "default") {
			s.AppendCert = true
		}
	}

	// Unmarshalling
	errSlice.AddNonNils(
		p.unmarshalString("file", &s.File),
		p.unmarshalString("tag", &s.Tag),
		p.unmarshalString("destination", &s.Destination),
		p.unmarshalString("name", &s.Name),
		p.unmarshalString("context", &s.Context),
		p.unmarshalString("secret", &s.Secret),
		p.unmarshalBool("append-cert", &s.AppendCert),
		p.unmarshalBool("push", &s.Push),
		p.unmarshalString("secretName", &s.SecretName),
	)
	errSlice.Add(p.unmarshalStringSlice("args", &s.Args)...)
	errSlice.Add(p.unmarshalStringSlice("secretArgs", &s.SecretArgs)...)

	// Validation
	errSlice.AddNonNils(
		p.validateRequiredString("file"),
		p.validateRequiredString("tag"),
	)
	return s, errSlice
}
