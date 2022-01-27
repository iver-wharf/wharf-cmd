package builder

type Status byte

const (
	StatusUnknown Status = iota
	StatusSuccess
	StatusFailed
	StatusCancelled
)

func (s Status) String() string {
	switch s {
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
