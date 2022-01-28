package worker

type Status byte

const (
	StatusUnknown Status = iota
	StatusNone
	StatusSuccess
	StatusFailed
	StatusCancelled
	StatusSkipped
)

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
