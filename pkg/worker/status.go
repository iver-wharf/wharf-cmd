package worker

import (
	"encoding/json"
	"strings"
)

// Status is an enum of the different statuses for a Wharf build, stage, or step.
type Status byte

const (
	// StatusUnknown means no status has been set. This is an errornous status.
	StatusUnknown Status = iota
	// StatusNone means no execution has been performed. Such as when running a
	// Wharf build stage with no steps.
	StatusNone
	// StatusScheduling means the step is not yet running.
	StatusScheduling
	// StatusRunning means the step is now running.
	StatusRunning
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
	case StatusScheduling:
		return "Scheduling"
	case StatusRunning:
		return "Running"
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

// ParseStatus parses a string as a status, or return StatusUnknown if it cannot
// find a matching status value. This is the inverse of the Status.String()
// method.
func ParseStatus(s string) Status {
	switch strings.ToLower(s) {
	case "none":
		return StatusNone
	case "scheduling":
		return StatusScheduling
	case "running":
		return StatusRunning
	case "success":
		return StatusSuccess
	case "failed":
		return StatusFailed
	case "cancelled":
		return StatusCancelled
	default:
		return StatusUnknown
	}
}

// UnmarshalJSON implements json.Unmarshaler
func (s *Status) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	*s = ParseStatus(str)
	return nil
}

// MarshalJSON implements json.Marshaler
func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}
