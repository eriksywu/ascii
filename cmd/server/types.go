package server

import (
	"context"
	"github.com/google/uuid"
	"io"
)

type ASCIIImageService interface {
	GetASCIIImage(context.Context, uuid.UUID) (bool, []byte, error)
	NewASCIIImageAsync(context.Context, io.ReadCloser) (*uuid.UUID, error)
	NewASCIIImageSync(context.Context, io.ReadCloser) (*uuid.UUID, error)
	GetImageList(context.Context) ([]uuid.UUID, error)
}
