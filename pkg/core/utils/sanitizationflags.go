package utils

type SanitizationFlags byte

const (
	FromLastCR SanitizationFlags = 1 << iota
	RemoveAnsiCodes
	AllSanitizationMethods = FromLastCR | RemoveAnsiCodes
)

func (m SanitizationFlags) HasFlag(f SanitizationFlags) bool {
	return m&f == f
}
