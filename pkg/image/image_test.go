package image

import (
	"bytes"
	"context"
	"encoding/base64"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	_ "image/png" // register png decoder
	"io"
	"io/ioutil"
	"strings"
	"testing"
)


type MockImageStore struct {
	data map[uuid.UUID]string
}

func (m *MockImageStore) PushASCIIImage(asciiImage string, id uuid.UUID) error {
	m.data[id] = asciiImage
	return nil
}

func (m *MockImageStore) GetASCIIImage(id uuid.UUID) (bool, string, error) {
	d, k:= m.data[id]
	return k, d, nil
}

func (m *MockImageStore) ListASCIIImages() ([]uuid.UUID, error) {
	return nil, nil
}

var _ ImageStore = (*MockImageStore)(nil)

// E2E logic and error handling tests
func TestService_NewASCIIImageAsyncE2E_BadImage(t *testing.T) {
	service := NewService(&MockImageStore{data: make(map[uuid.UUID]string)})

	r := ioutil.NopCloser(strings.NewReader("this is not an image"))

	id, err := service.NewASCIIImageAsync(context.Background(), r)

	// explanation: the actual processing task itself will error out but this is an async call so all it does is creat the Task
	assert.NoError(t, err)
	assert.NotNil(t, id)
	assert.Equal(t, 1, len(service.asyncTasks))
	task := service.asyncTasks[*id]
	result, taskErr := task.Result()
	assert.NoError(t, taskErr)
	assert.True(t, (*task).State().IsTerminal())
	err = result.Error
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "processing error"))

	finished, _,  err := service.GetASCIIImage(context.Background(), *id)
	assert.Error(t, err)
	_, isProcessError := err.(InternalProcessingError)
	assert.True(t, isProcessError)
	assert.False(t, finished)
}

func TestService_NewASCIIImageSyncE2E_BadImage(t *testing.T) {
	service := NewService(&MockImageStore{data: make(map[uuid.UUID]string)})

	r := ioutil.NopCloser(strings.NewReader("this is not an image"))

	id, err := service.NewASCIIImageSync(context.Background(), r)

	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "processing erro"))
	assert.Nil(t, id)
	// we will keep error'ed out tasks in our local cache
	assert.Equal(t, 1, len(service.asyncTasks))

}

// base64 string rep of an actual png image: http://www.schaik.com/pngsuite/basn0g01.png
const testPNGBase64Representation = "iVBORw0KGgoAAAANSUhEUgAAACAAAAAgAQAAAABbAUdZAAAABGdBTUEAAYagMeiWXwAAAFtJREFUeJwtzLEJAzAMBdHr0gSySiALejRvkBU8gsGNCmFFB1Hx4IovqurSpIRszqklUwbnUzRXEuIRsiG/SyY9G0JzJSVei9qynm9qyjBpLp0pYW7pbzBl8L8fEIdJL9AvFMkAAAAASUVORK5CYII="

func getGoodImageRCloser() io.ReadCloser {
	data, err := base64.StdEncoding.DecodeString(testPNGBase64Representation)
	if err != nil {
		// not supposed to get here
		panic(err)
	}
	return ioutil.NopCloser(bytes.NewReader(data))
}

func TestService_NewASCIIImageAsyncE2E_GoodImage(t *testing.T) {
	service := NewService(&MockImageStore{data: make(map[uuid.UUID]string)})

	id, err := service.NewASCIIImageAsync(context.Background(), getGoodImageRCloser())

	// explanation: the actual processing task itself will error out but this is an async call so all it does is creat the Task
	assert.NoError(t, err)
	assert.NotNil(t, id)
	assert.Equal(t, 1, len(service.asyncTasks))
	task := service.asyncTasks[*id]
	result, taskErr := task.Result()
	assert.NoError(t, taskErr)
	assert.True(t, (*task).State().IsTerminal())
	err = result.Error
	assert.NoError(t, err)

	finished, asciiImage,  err := service.GetASCIIImage(context.Background(), *id)
	assert.NoError(t, err)
	assert.True(t, finished)
	t.Logf("generated ascci image: \n%s", asciiImage)
}

func TestService_NewASCIIImageSyncE2E_GoodImage(t *testing.T) {
	service := NewService(&MockImageStore{data: make(map[uuid.UUID]string)})

	id, err := service.NewASCIIImageSync(context.Background(), getGoodImageRCloser())

	assert.NoError(t, err)
	assert.NotNil(t, id)
	assert.Equal(t, 0, len(service.asyncTasks))

	finished, asciiImage,  err := service.GetASCIIImage(context.Background(), *id)
	assert.NoError(t, err)
	assert.True(t, finished)
	t.Logf("generated ascci image: \n%s", asciiImage)
}
