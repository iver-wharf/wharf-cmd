package wharfyml

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

func (s StepDocker) visitStepTypeNode(p nodeMapParser) (StepType, Errors) {
	s.Meta = getStepTypeMeta(p)

	s.Destination = ""  // TODO: default to "${registry}/${group}/${REPO_NAME}/${step_name}"
	s.Name = ""         // TODO: default to "${step_name}"
	s.Group = ""        // TODO: default to "${REPO_GROUP}"
	s.Registry = ""     // TODO: default to "${REG_URL}"
	s.AppendCert = true // TODO: default to true if REPO_GROUP starts with "default", case insensitive

	s.Push = true
	s.Secret = "gitlab-registry"

	var errSlice Errors

	// Unmarshalling
	errSlice.addNonNils(
		p.unmarshalString("file", &s.File),
		p.unmarshalString("tag", &s.Tag),
		p.unmarshalString("destination", &s.Destination),
		p.unmarshalString("name", &s.Name),
		p.unmarshalString("group", &s.Group),
		p.unmarshalString("context", &s.Context),
		p.unmarshalString("secret", &s.Secret),
		p.unmarshalString("registry", &s.Registry),
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
