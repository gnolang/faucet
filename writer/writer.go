package writer

// ResponseWriter outlines the interface any
// response writer needs to implement
type ResponseWriter interface {
	// WriteResponse takes in the JSON response
	// which is either a single object, or a batch.
	// When technology goes ahead of its time and implements
	// generic methods, this interface should be changed
	// https://github.com/golang/go/issues/49085
	WriteResponse(response any)
}
