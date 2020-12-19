package server

import (
	"github.com/google/uuid"
	"io"
)

type ASCIIImageService interface {
	GetASCIIImage(id uuid.UUID) ([]byte, error)
	NewASCIIImage(r io.ByteReader) (uuid.UUID, error)
}
