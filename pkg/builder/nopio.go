package builder

type nopWriter struct{}

func (nopWriter) Write(bytes []byte) (int, error) {
	return len(bytes), nil
}
