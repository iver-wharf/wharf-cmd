package main

import (
	_ "embed"
	"fmt"

	"github.com/iver-wharf/wharf-cmd/cmd"
	"github.com/iver-wharf/wharf-core/pkg/app"
)

//go:embed assets/version.yaml
var versionFile []byte

func getVersion() (app.Version, error) {
	var version app.Version
	if err := app.UnmarshalVersionYAML(versionFile, &version); err != nil {
		return app.Version{}, fmt.Errorf("load version file: %w", err)
	}
	return version, nil
}

func main() {
	version, err := getVersion()
	if err != nil {
		fmt.Println("Failed to load version:", err)
	}
	cmd.Execute(version)
}
