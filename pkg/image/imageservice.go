package image

import (
	"bytes"
	"context"
	"fmt"
	"github.com/eriksywu/ascii/pkg/logging"
	async "github.com/eriksywu/go-async"
	"github.com/google/uuid"
	"github.com/qeesung/image2ascii/convert"
	"github.com/sirupsen/logrus"
	"image"
	_ "image/png" // register png decoder
	"io"
	"io/ioutil"
)


type Service struct {
	imageConverter *convert.ImageConverter
	imageStore     ImageStore

	//asyncTasks stores all currently running image processing tasks
	//Shameless plug: using my own go-async pkg here
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
	// but copy over context-based logger
	asyncContext := context.WithValue(context.Background(), "logger", getLogger(ctx))
	// input ReadCloser will be auto-closed at the end of the httphandlefunc. We need to copy it.
	rCopyBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	newRW :=  ioutil.NopCloser(bytes.NewBuffer(rCopyBytes))
	err = i.createAndPushNewConversionTask(asyncContext, newRW, id).RunAsync()
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (i *Service) NewASCIIImageSync(ctx context.Context, r io.ReadCloser) (*uuid.UUID, error) {
	id := uuid.New()
	task := i.createAndPushNewConversionTask(ctx, r, id)
	result, err := task.Result()
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, result.Error
	}
	return &id, nil
}

func (i *Service) createAndPushNewConversionTask(ctx context.Context, r io.ReadCloser, id uuid.UUID) *async.Task {
	worker := func(_ context.Context) (async.T, error) {
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
			logger.Errorf("decoding image failed: %s", err)
			return nil, fmt.Errorf("error processing png image: %w", ImageProcessingError)
		}

		if isContextCancelled(ctx) {
			return nil, InternalProcessingError{fmt.Errorf("context cancelled")}
		}

		// step2: convert to ascii string
		logger.Infof("converting image %s to ascii", id)
		convertOpts := convert.Options{
			Ratio:       1,
			FixedWidth:  -1,
			FixedHeight: -1,
			FitScreen:   false,
		}
		image := i.imageConverter.Image2ASCIIString(m, &convertOpts)

		if isContextCancelled(ctx) {
			return nil, InternalProcessingError{fmt.Errorf("context cancelled")}
		}

		// step3: push to image store
		logger.Infof("storing image %s", id)
		err = i.imageStore.PushASCIIImage(image, id)
		if err != nil {
			logger.Errorf("saving image failed: %s", err)
			return nil, fmt.Errorf("error storing ascii image: %w", ImageStorageError)
		}

		// step4: remove self from the asyncTask map
		logger.Infof("processing successful")
		delete(i.asyncTasks, id)
		return id, nil
	}

	task := async.CreateTask(func() context.Context { return ctx }, async.WorkFn(worker))
	i.asyncTasks[id] = task
	return task
}

func (i *Service) GetASCIIImage(ctx context.Context, id uuid.UUID) (bool, []byte, error) {
	logger := getLogger(ctx)
	logger.Infof("attempting to fetch ascii image for imageID = %s", id)
	processingTask, k := i.asyncTasks[id]
	// if processing task is still running
	if k {
		logger.Infof("image has not yet finished processing")
		status := processingTask.State()
		if status.IsRunning() {
			return false, nil, nil
		}
		result, taskErr := processingTask.Result()
		if taskErr != nil {
			return false, nil, InternalProcessingError{fmt.Errorf("internal application processing failed: %w", taskErr)}
		}
		if result.Error != nil {
			return false, nil, InternalProcessingError{fmt.Errorf("image processing failed: %w", result.Error)}
		}
		if status == async.Cancelled || status == async.InternalError || status == async.Cancelling {
			return false, nil, InternalProcessingError{fmt.Errorf("unknown image processing failed/cancelled")}
		}
	}
	logger.Infof("grabbing image from image store")
	exists, image, err := i.imageStore.GetASCIIImage(id)
	if err != nil {
		return false, nil, err
	}
	if !exists {
		logger.Warnf("image %s does not exist", id)
		return false, nil, NewResourceNotFoundError(fmt.Errorf("image %s does not exist", id.String()))
	}
	logger.Infof("found image from image store")
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
		return logrus.NewEntry(logging.Logger.Logger)
	}
	return logger.(*logrus.Entry)
}
