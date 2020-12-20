package image

import "github.com/google/uuid"

type ImageStore interface {
	PushASCIIImage(asciiImage string, id uuid.UUID) error
	GetASCIIImage(id uuid.UUID) (bool, string, error)
	ListASCIIImages() ([]uuid.UUID, error)
}
