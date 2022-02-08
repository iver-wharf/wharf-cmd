package wharfyml

import "fmt"

type StepDocker struct {
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
}

func (StepDocker) StepTypeName() string { return "docker" }

func (s StepDocker) Validate() (errSlice errorSlice) {
	if s.File == "" {
		errSlice.add(fmt.Errorf("%w: file", ErrStepTypeMissingRequired))
	}
	if s.Tag == "" {
		errSlice.add(fmt.Errorf("%w: tag", ErrStepTypeMissingRequired))
	}
	return
}

func (s *StepDocker) unmarshalNodes(nodes nodeMapUnmarshaller) (errSlice errorSlice) {
	errSlice.addNonNils(
		nodes.unmarshalString("file", &s.File),
		nodes.unmarshalString("tag", &s.Tag),
		nodes.unmarshalString("destination", &s.Destination),
		nodes.unmarshalString("name", &s.Name),
		nodes.unmarshalString("group", &s.Group),
		nodes.unmarshalString("context", &s.Context),
		nodes.unmarshalString("secret", &s.Secret),
		nodes.unmarshalString("registry", &s.Registry),
		nodes.unmarshalBool("append-cert", &s.AppendCert),
		nodes.unmarshalBool("push", &s.Push),
	)
	errSlice.add(nodes.unmarshalStringSlice("args", &s.Args)...)
	return
}

func (s *StepDocker) resetDefaults() errorSlice {
	s.Destination = ""  // TODO: default to "${registry}/${group}/${REPO_NAME}/${step_name}"
	s.Name = ""         // TODO: default to "${step_name}"
	s.Group = ""        // TODO: default to "${REPO_GROUP}"
	s.Registry = ""     // TODO: default to "${REG_URL}"
	s.AppendCert = true // TODO: default to true if REPO_GROUP starts with "default", case insensitive

	s.Push = true
	s.Secret = "gitlab-registry"
	return nil
}
