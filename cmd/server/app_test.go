package server

import (
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

var _ ASCIIImageService = (*ASCIIImageServiceMock)(nil)

type ASCIIImageServiceMock struct {
	GetASCIIImageFn    func() ([]byte, error)
	GetNewASCIIImageFn func() (uuid.UUID, error)
	GetImageListFn     func() ([]uuid.UUID, error)
}

func (A ASCIIImageServiceMock) GetASCIIImage(id uuid.UUID) ([]byte, error) {
	if A.GetASCIIImageFn == nil {
		return nil, nil
	}
	return A.GetASCIIImageFn()
}

func (A ASCIIImageServiceMock) NewASCIIImage(r io.ReadCloser) (uuid.UUID, error) {
	if A.GetNewASCIIImageFn == nil {
		return uuid.New(), nil
	}
	return A.GetNewASCIIImageFn()
}

func (A ASCIIImageServiceMock) GetImageList() ([]uuid.UUID, error) {
	if A.GetImageListFn == nil {
		return nil, nil
	}
	return A.GetImageListFn()
}

func TestGetASCIIImageHandler_BadUID(t *testing.T) {
	req, err := http.NewRequest("GET", "/images/NOT-A-UID", nil)
	if err != nil {
		t.Fatal(err)
	}

	mockService := &ASCIIImageServiceMock{}
	testSubject := BuildServer(mockService, 8080)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(testSubject.getImageBaseHandler())

	handler.ServeHTTP(rr, req)

	status := rr.Code
	assert.NotEqual(t, status, http.StatusOK)
}

func TestGetASCIIImageHandler_ServiceFails(t *testing.T) {
	req, err := http.NewRequest("GET", "/images/NOT-A-UID", nil)
	if err != nil {
		t.Fatal(err)
	}

	mockService := &ASCIIImageServiceMock{}
	mockService.GetASCIIImageFn = func() ([]byte, error) {
		return nil, errors.New("some error")
	}
	testSubject := BuildServer(mockService, 8080)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(testSubject.getImageBaseHandler())

	handler.ServeHTTP(rr, req)

	status := rr.Code
	assert.NotEqual(t, status, http.StatusOK)
}

// Not much need to test the other handlers since they're all business logic