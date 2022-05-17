package aggregator

// Source is an interface that is meant to be implemented for each different
// method that is required to retrieve data.
//
// Configuration may be provided through fields on the implementing struct.
//
// Retrieval of the data should be done from the PushInto function.
//
// The channel is NOT meant to be closed by the implementing struct. However,
// the function finishing should be indicative that there is nothing more to
// retrieve, and that the channel can therefore be closed without missing out on
// any data. This is to allow several sources to push into a single destination.
type Source[DstType any] interface {
	// PushInto retrieves data from a specific source, performing any necessary
	// conversion to the destination type.
	//
	// The destination channel is not closed upon return.
	PushInto(dst chan<- DstType) error
}
