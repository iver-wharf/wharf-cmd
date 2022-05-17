package aggregator

type source[DstType any] interface {
	pushInto(dst chan<- DstType) error
}
