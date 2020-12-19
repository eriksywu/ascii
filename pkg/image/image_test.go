package image

import (
	async "github.com/eriksywu/go-async"
	"github.com/google/uuid"
	"github.com/qeesung/image2ascii/convert"
	"image"
	"testing"
)

var _ ImageConverter = (*MockImageConverter)(nil)

type MockImageConverter struct {
	f func() string
}

func (m *MockImageConverter) Image2ASCIIString(image image.Image, options *convert.Options) string {
	if m.f != nil {
		return m.f()
	}
	return ""
}

type MockImageStore struct {
	push func() error
	get func() (bool, string, error)
	list func() ([]uuid.UUID, error)
}

func (m *MockImageStore) PushASCIIImage(asciiImage string, id uuid.UUID) error {
	if m.push != nil {
		return m.push()
	}
	return nil
}

func (m *MockImageStore) GetASCIIImage(id uuid.UUID) (bool, string, error) {
	if m.get != nil {
		m.get()
	}
	return false, "", nil
}

func (m *MockImageStore) ListASCIIImages() ([]uuid.UUID, error) {
	if m.list != nil {
		m.list()
	}
	return nil, nil
}

var _ ImageStore = (*MockImageStore)(nil)

func TestService_NewASCIIImageAsync(t *testing.T) {
	service := &Service{imageStore: &MockImageStore{}}
	service.imageConverter = &MockImageConverter{}
	service.asyncTasks = make(map[uuid.UUID]*async.Task)


}
