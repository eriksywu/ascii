package image

import (
	"context"
	"fmt"
	async "github.com/eriksywu/go-async"
	"github.com/google/uuid"
	"github.com/qeesung/image2ascii/convert"
	"image"
	_ "image/png" // register png decoder
	"io"
)

type Service struct {
	imageConverter *convert.ImageConverter
	imageStore     ImageStore

	//asyncTasks stores all currently running image processing tasks
	//Shameless plug: using my own go-async pkg here :)
	//TODO thread safety
	asyncTasks map[uuid.UUID]*async.Task
}

func (i *Service) NewASCIIImage(ctx context.Context, r io.ReadCloser) (*uuid.UUID, error) {
	id := uuid.New()
	defer r.Close()
	m, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("error processing png image: %w", NewInternalProcessingError(err))
	}
	image := i.imageConverter.Image2ASCIIString(m, &convert.DefaultOptions)
	err = i.imageStore.PushASCIIImage(image, id)
	if err != nil {
		return nil, fmt.Errorf("error storing ascii image: %w", NewInternalProcessingError(err))
	}
	return &id, nil
}

func (i *Service) createAndPushNewConversionTask(ctx context.Context, r io.ReadCloser, id uuid.UUID) *async.Task {
	worker := func(_ context.Context) (async.T, error) {
		defer r.Close()
		// check if context is closed at every step
		if i.isContextCancelled(ctx) {
			return nil, InternalProcessingError{fmt.Errorf("context timeout")}

		}
		// step1. decode png
		m, _, err := image.Decode(r)
		if err != nil {
			return nil, fmt.Errorf("error processing png image: %w", NewInternalProcessingError(err))
		}

		if i.isContextCancelled(ctx) {
			return nil, InternalProcessingError{fmt.Errorf("context timeout")}

		}

		// step2: convert to ascii string
		image := i.imageConverter.Image2ASCIIString(m, &convert.DefaultOptions)

		if i.isContextCancelled(ctx) {
			return nil, InternalProcessingError{fmt.Errorf("context timeout")}

		}
		// step3: push to image store
		err = i.imageStore.PushASCIIImage(image, id)
		if err != nil {
			return nil, fmt.Errorf("error storing ascii image: %w", NewInternalProcessingError(err))
		}

		// step4: remove self from the asyncTask map
		delete(i.asyncTasks, id)

		return id, nil
	}

	task := async.CreateTask(func() context.Context { return ctx }, async.WorkFn(worker))
	i.asyncTasks[id] = task
	return task
}

func (i *Service) isContextCancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		break
	}
	return false
}

func (i *Service) GetASCIIImage(ctx context.Context, id uuid.UUID) ([]byte, error) {
	panic("implement me")
}

func (i *Service) GetImageList(ctx context.Context) ([]uuid.UUID, error) {
	panic("implement me")
}
