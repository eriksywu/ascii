package image

import (
	"context"
	"fmt"
	async "github.com/eriksywu/go-async"
	"github.com/google/uuid"
	"github.com/qeesung/image2ascii/convert"
	"github.com/sirupsen/logrus"
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

func NewService(imageStore ImageStore) *Service {
	service := &Service{imageStore: imageStore}
	service.imageConverter = convert.NewImageConverter()
	service.asyncTasks = make(map[uuid.UUID]*async.Task)
	return service
}

func (i *Service) NewASCIIImageAsync(ctx context.Context, r io.ReadCloser) (*uuid.UUID, error) {
	id := uuid.New()
	// construct a new context that's not tied to the request context to decouple this async op from the request's cancelFunc
	// but copy over context data?
	asyncContext := context.Background()
	_ = i.createAndPushNewConversionTask(asyncContext , r, id)
	return &id, nil
}

func (i *Service) NewASCIIImageSync(ctx context.Context, r io.ReadCloser) (*uuid.UUID, error) {
	id := uuid.New()
	task := i.createAndPushNewConversionTask(ctx , r, id)
	_, err := task.Result()
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (i *Service) createAndPushNewConversionTask(ctx context.Context, r io.ReadCloser, id uuid.UUID) *async.Task {
	worker := func(_ context.Context) (async.T, error) {
		defer delete(i.asyncTasks, id)
		logger := getLogger(ctx)

		// check if context is closed at every step
		if isContextCancelled(ctx) {
			return nil, InternalProcessingError{fmt.Errorf("context cancelled")}

		}

		// step1. decode png
		logger.Infof("decoding image %s", id)
		defer r.Close()
		m, _, err := image.Decode(r)
		if err != nil {
			return nil, fmt.Errorf("error processing png image: %w", NewInternalProcessingError(err))
		}

		if isContextCancelled(ctx) {
			return nil, InternalProcessingError{fmt.Errorf("context cancelled")}
		}

		// step2: convert to ascii string
		logger.Infof("converting image %s to ascii", id)
		image := i.imageConverter.Image2ASCIIString(m, &convert.DefaultOptions)

		if isContextCancelled(ctx) {
			return nil, InternalProcessingError{fmt.Errorf("context cancelled")}

		}

		// step3: push to image store
		logger.Infof("storing image %s", id)
		err = i.imageStore.PushASCIIImage(image, id)
		if err != nil {
			return nil, fmt.Errorf("error storing ascii image: %w", NewInternalProcessingError(err))
		}

		// step4: remove self from the asyncTask map
		logger.Infof("processing successful")


		return id, nil
	}

	task := async.CreateTask(func() context.Context { return ctx }, async.WorkFn(worker))
	i.asyncTasks[id] = task
	return task
}

func (i *Service) GetASCIIImage(ctx context.Context, id uuid.UUID) (bool, []byte, error) {
	processingTask, k := i.asyncTasks[id]
	// if processing task is still running
	if k {
		status := processingTask.State()
		if status.IsRunning() {
			return false, nil, nil
		}
		if status == async.Cancelled || status == async.InternalError || status == async.Cancelling {
			return false, nil, InternalProcessingError{fmt.Errorf("image processing failed/cancelled")}
		}
	}
	image, err := i.imageStore.GetASCIIImage(id)
	if err != nil {
		return true, nil, err
	}
	return true, []byte(image), nil
}

func (i *Service) GetImageList(ctx context.Context) ([]uuid.UUID, error) {
	return i.imageStore.ListASCIIImages()
}

func isContextCancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		break
	}
	return false
}

func getLogger(ctx context.Context) *logrus.Entry {
	logger := ctx.Value("logger")
	if logger == nil || logger.(*logrus.Entry) == nil {
		return &logrus.Entry{}
	}
	return logger.(*logrus.Entry)
}


