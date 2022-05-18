package flagtypes

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ensure they conform to the interfaces
var dryRun = DryRunNone
var _ pflag.Value = &dryRun

// DryRun is an enum flag for setting dry-run.
type DryRun string

const (
	// DryRunNone disables dry-run. The build will be performed as usual
	DryRunNone DryRun = "none"
	// DryRunClient only logs what would be run, without contacting Kubernetes
	DryRunClient DryRun = "client"
	// DryRunServer submits server-side dry-run requests to Kubernetes
	DryRunServer DryRun = "server"
)

// Set implements the pflag.Value and fmt.Stringer interfaces.
// This returns a human-readable representation of the loglevel.
func (d *DryRun) String() string {
	return fmt.Sprintf(`"%s"`, string(*d))
}

// Set implements the pflag.Value interface.
// This parses the loglevel string and updates the loglevel variable.
func (d *DryRun) Set(value string) error {
	dryRun, err := parseDryRun(value)
	if err != nil {
		return err
	}
	*d = dryRun
	return nil
}

func parseDryRun(value string) (DryRun, error) {
	switch strings.ToLower(value) {
	case "none":
		return DryRunNone, nil
	case "client":
		return DryRunClient, nil
	case "server":
		return DryRunServer, nil
	default:
		return "", errors.New(`must be one of "none", "client", or "server"`)
	}
}

// Type implements the pflag.Value interface.
// The value is only used in help text.
func (d *DryRun) Type() string {
	return "dry-run"
}

// CompleteDryRun returns completions for the DryRun type.
func CompleteDryRun(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return []string{
		string(DryRunNone) + "\tDisables dry-run. The build will be performed as usual",
		string(DryRunClient) + "\tOnly logs what would be run, without contacting Kubernetes",
		string(DryRunServer) + "\tSubmits server-side dry-run requests to Kubernetes",
	}, cobra.ShellCompDirectiveNoFileComp
}
