package wharfyml

import (
	"errors"
	"fmt"
	"strings"

	"github.com/iver-wharf/wharf-cmd/internal/errutil"
	"github.com/iver-wharf/wharf-cmd/pkg/wharfyml/visit"
	"gopkg.in/yaml.v3"
)

// Errors related to parsing the run conditions.
var (
	ErrInvalidRunCondition = errors.New("invalid run condition")
)

// StageRunsIf is an enum of different run behaviors for a stage, dependent on
// previous stages.
type StageRunsIf string

const (
	// StageRunsIfSuccess runs the stage if all previous stages were successful.
	StageRunsIfSuccess = "success"
	// StageRunsIfFail runs the stage if one or more of the previous stages were
	// unsuccessful.
	StageRunsIfFail = "fail"
	// StageRunsIfAlways always runs the stage, regardless of the success state
	// of previous stages.
	StageRunsIfAlways = "always"
)

func visitStageRunsIfNode(node *yaml.Node) (StageRunsIf, errutil.Slice) {
	runsIfStr, err := visit.String(node)
	if err != nil {
		return "", []error{err}
	}
	runsIf := StageRunsIf(strings.ToLower(runsIfStr))
	switch runsIf {
	case StageRunsIfAlways, StageRunsIfFail, StageRunsIfSuccess:
		return runsIf, nil
	default:
		err := fmt.Errorf("%w: '%s' must be one of 'always', 'success', or 'fail'", ErrInvalidRunCondition, runsIf)
		posErr := errutil.NewPosFromNode(err, node)
		return "", []error{posErr}
	}
}
