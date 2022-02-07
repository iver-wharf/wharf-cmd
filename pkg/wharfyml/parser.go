package wharfyml

import "io"

type Definition struct {
	Stages []Stage
}

func parse(reader io.Reader) (Definition, []error) {
	return Definition{}, nil
}
