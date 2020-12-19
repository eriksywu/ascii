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

type ResourceNotFoundError struct {
	e error
}

func (e ResourceNotFoundError) Error() string {
	return fmt.Sprintf("internal processing error: %v", e.e)
}

func NewResourceNotFoundError(err error) ResourceNotFoundError {
	return ResourceNotFoundError{e: err}
}
