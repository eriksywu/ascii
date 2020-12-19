package image

import "github.com/google/uuid"

type ImageStore interface {
	PushASCIIImage(asciiImage string, id uuid.UUID)  error
	GetASCIIImage(id uuid.UUID) (string, error)
	ListASCIIImages() ([]uuid.UUID, error)
}
