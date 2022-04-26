package varsub

import (
	"os"
	"strings"
)

const osEnvSourceName = "OS environment variables"

// NewOSEnvSource creates a new Source that uses your OS environment variables
// with a prefix as variables.
//
// To use all environment variables, you can specify an empty string as prefix.
func NewOSEnvSource(prefix string) Source {
	return osEnvSource{prefix}
}

type osEnvSource struct {
	prefix string
}

// Lookup tries to look up a value based on name and returns that value as
// well as true on success, or false if the variable was not found.
func (s osEnvSource) Lookup(name string) (Var, bool) {
	val, ok := os.LookupEnv(s.prefix + name)
	if !ok {
		return Var{}, false
	}
	return Var{
		Key:    name,
		Value:  val,
		Source: osEnvSourceName,
	}, true
}

// ListVars will return a slice of all variables that this varsub Source
// provides.
func (s osEnvSource) ListVars() []Var {
	var vars []Var
	for _, env := range os.Environ() {
		key, val, ok := strings.Cut(env, "=")
		if !ok {
			continue
		}
		trimmedKey := strings.TrimPrefix(key, s.prefix)
		if len(trimmedKey) == len(key) {
			// No prefix was trimmed. It didn't have the prefix
			continue
		}
		vars = append(vars, Var{
			Key:    trimmedKey,
			Value:  val,
			Source: osEnvSourceName,
		})
	}
	return vars
}
