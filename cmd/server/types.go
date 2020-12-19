package server

import (
	"context"
	"github.com/google/uuid"
	"io"
)

type ASCIIImageService interface {
	GetASCIIImage(context.Context, uuid.UUID) ([]byte, error)
	NewASCIIImage(context.Context, io.ReadCloser) (*uuid.UUID, error)
	GetImageList(context.Context) ([]uuid.UUID, error)
}
