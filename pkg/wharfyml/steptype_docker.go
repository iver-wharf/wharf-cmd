package wharfyml

import (
	"fmt"
	"strings"

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

// StepTypeName returns the name of this step type.
func (StepDocker) StepTypeName() string { return "docker" }

func (s StepDocker) visitStepTypeNode(stepName string, p nodeMapParser, source varsub.Source) (StepType, Errors) {
	s.Meta = getStepTypeMeta(p)

	s.Destination = ""
	s.Name = stepName

	s.Push = true
	s.Secret = "gitlab-registry"

	var errSlice Errors

	if _, ok := p.nodes["destination"]; !ok {
		if _, ok := p.nodes["registry"]; !ok {
			regURL, ok := source.Lookup("REG_URL")
			if !ok {
				errSlice.add(wrapPosErrorNode(fmt.Errorf(
					"%w: need REG_URL or 'registry' to construct 'destination'",
					ErrMissingBuiltinVar),
					p.parent))
			} else {
				newNode, err := newNodeWithValue(p.parent, regURL)
				if err != nil {
					errSlice.add(wrapPosErrorNode(fmt.Errorf(
						"read REG_URL to construct 'destination': %w", err),
						p.parent))
				} else {
					p.nodes["registry"] = newNode
				}
			}
		}

		if _, ok := p.nodes["group"]; !ok {
			repoGroup, ok := source.Lookup("REPO_GROUP")
			if !ok {
				errSlice.add(wrapPosErrorNode(fmt.Errorf(
					"%w: need REPO_GROUP or 'group' to construct 'destination'",
					ErrMissingBuiltinVar),
					p.parent))
			} else {
				newNode, err := newNodeWithValue(p.parent, repoGroup)
				if err != nil {
					errSlice.add(wrapPosErrorNode(fmt.Errorf(
						"read REPO_GROUP to construct 'destination': %w", err),
						p.parent))
				} else {
					p.nodes["group"] = newNode
				}
			}
		}

		repoNameVar, ok := source.Lookup("REPO_NAME")
		if !ok {
			errSlice.add(wrapPosErrorNode(fmt.Errorf(
				"%w: need REPO_NAME to construct 'destination'",
				ErrMissingBuiltinVar),
				p.parent))
		} else {
			newNode, err := newNodeWithValue(p.parent, repoNameVar)
			errSlice.add(wrapPosErrorNode(fmt.Errorf(
				"read REPO_NAME to construct 'destination': %w", err),
				p.parent))
			// __repoName isn't a real field, but we're setting it to abuse
			// p.unmarshalString()
			p.nodes["__repoName"] = newNode
		}

		var repoName string
		errSlice.addNonNils(
			p.unmarshalString("registry", &s.Registry),
			p.unmarshalString("group", &s.Group),
			p.unmarshalString("name", &s.Name),
			p.unmarshalString("__repoName", &repoName),
		)
		if len(errSlice) == 0 {
			s.Destination = fmt.Sprintf("%s/%s/%s/%s",
				s.Registry, s.Group, repoName, s.Name,
			)
			// default to "${registry}/${group}/${REPO_NAME}/${step_name}"
		}
	}

	if _, ok := p.nodes["group"]; !ok {
		repoGroup, ok := source.Lookup("REPO_GROUP")
		if ok {
			newNode, err := newNodeWithValue(p.parent, repoGroup)
			if err != nil {
				errSlice.add(wrapPosErrorNode(fmt.Errorf(
					"read REPO_GROUP to construct 'destination': %w", err),
					p.parent))
			} else {
				p.nodes["group"] = newNode
				err := p.unmarshalString("group", &s.Group)
				if err != nil {
					errSlice.add(err)
				} else if strings.HasPrefix(strings.ToLower(s.Group), "default") {
					s.AppendCert = true
				}
			}
		}
	}

	// Unmarshalling
	errSlice.addNonNils(
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
	errSlice.add(p.unmarshalStringSlice("args", &s.Args)...)
	errSlice.add(p.unmarshalStringSlice("secretArgs", &s.SecretArgs)...)

	// Validation
	errSlice.addNonNils(
		p.validateRequiredString("file"),
		p.validateRequiredString("tag"),
	)
	return s, errSlice
}
