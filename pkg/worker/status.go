package worker

// Status is an enum of the different statuses for a Wharf build, stage, or step.
type Status byte

const (
	// StatusUnknown means no status has been set. This is an errornous status.
	StatusUnknown Status = iota
	// StatusNone means no execution has been performed. Such as when running a
	// Wharf build stage with no steps.
	StatusNone
	// StatusSuccess means the build succeeded.
	StatusSuccess
	// StatusFailed means the build failed. More details of how it failed can be
	// found in the StepResult.Error field.
	StatusFailed
	// StatusCancelled means the build, stage, or step was cancelled.
	StatusCancelled
)

// String implements the fmt.Stringer interface.
func (s Status) String() string {
	switch s {
	case StatusNone:
		return "None"
	case StatusSuccess:
		return "Success"
	case StatusFailed:
		return "Failed"
	case StatusCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}
