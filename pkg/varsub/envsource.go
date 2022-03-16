package varsub

import "os"

// EnvSource is a Source implementation that will pull values from
// environment variables
type EnvSource struct{}

func (EnvSource) Lookup(name string) (interface{}, bool) {
	return os.LookupEnv(name)
}
