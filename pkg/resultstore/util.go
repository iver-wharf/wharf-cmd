package resultstore

import "gopkg.in/typ.v3/pkg/chans"

func sendAllToChan[C chans.Sender[E], S ~[]E, E any](ch C, values S) {
	for _, v := range values {
		ch <- v
	}
}
