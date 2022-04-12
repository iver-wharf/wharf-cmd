package wharfcmd

import (
	_ "embed"
	"fmt"

	"github.com/iver-wharf/wharf-core/pkg/app"
)

//go:embed assets/version.yaml
var versionFile []byte

// GetVersion returns this app's version.
func GetVersion() (app.Version, error) {
	var version app.Version
	if err := app.UnmarshalVersionYAML(versionFile, &version); err != nil {
		return app.Version{}, fmt.Errorf("load version file: %w", err)
	}
	return version, nil
}
