package image

import "fmt"

type InternalProcessingError struct {
	e error
}

func (e InternalProcessingError) Error() string {
	return fmt.Sprintf("internal processing error: %v", e.e)
}

func NewInternalProcessingError(err error) InternalProcessingError {
	return InternalProcessingError{e: err}
}
