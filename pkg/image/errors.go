package image

import (
	"fmt"
)

var ImageProcessingError = NewInternalProcessingError(fmt.Errorf("image processing error"))
var ImageStorageError = NewInternalProcessingError(fmt.Errorf("image storage error"))
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
	return fmt.Sprintf("resource not found: %v", e.e)
}

func NewResourceNotFoundError(err error) ResourceNotFoundError {
	return ResourceNotFoundError{e: err}
}

type InvalidInputError struct {
	e error
}

func (e InvalidInputError) Error() string {
	return fmt.Sprintf("invalid input error: %v", e.e)
}

func NewInvalidInputError(err error) InvalidInputError {
	return InvalidInputError{e: err}
}