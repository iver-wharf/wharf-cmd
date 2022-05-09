package steps

import "github.com/iver-wharf/wharf-cmd/internal/errutil"

// StepType is an interface that is implemented by all step types.
type StepType interface {
	Name() string
	Validate() errutil.Slice
}
