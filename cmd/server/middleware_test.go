package server

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWithLoggingContext(t *testing.T) {
	req, err := http.NewRequest("GET", "/someurl", nil)
	if err != nil {
		t.Fatal(err)
	}

	var dummyHandler httpMiddleWare = func(writer http.ResponseWriter, request *http.Request) {
		rContext := request.Context()
		correlationID := rContext.Value(CorrelationID)
		assert.NotNil(t, correlationID)
		_, err := uuid.Parse(correlationID.(string))
		assert.NoError(t, err)
		logger := rContext.Value(Logger)
		assert.NotNil(t, logger)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(dummyHandler.WithLoggingContext("dummyHandler"))
	handler.ServeHTTP(rr, req)
}

func TestWithTimeout_Cancels(t *testing.T) {
	req, err := http.NewRequest("GET", "/someurl", nil)
	if err != nil {
		t.Fatal(err)
	}

	var dummyHandler httpMiddleWare = func(writer http.ResponseWriter, request *http.Request) {
		rContext := request.Context()
		select {
		case <-time.After(5 * time.Second):
			t.Errorf("timeout wasn't caught")
		case <-rContext.Done():
			return
		}
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(dummyHandler.WithTimeout(1))
	handler.ServeHTTP(rr, req)
}

func TestWithTimeout_Finishes(t *testing.T) {
	req, err := http.NewRequest("GET", "/someurl", nil)
	if err != nil {
		t.Fatal(err)
	}

	var dummyHandler httpMiddleWare = func(writer http.ResponseWriter, request *http.Request) {
		rContext := request.Context()
		select {
		case <-time.After(1 * time.Second):
			return
		case <-rContext.Done():
			t.Errorf("timeout shouldn't have been caught")
		}
	}

	rr := httptest.NewRecorder()
	// this bit should finish after 1sec
	handler := http.HandlerFunc(dummyHandler.WithTimeout(5000))
	handler.ServeHTTP(rr, req)
}

func TestChainedMiddlewares(t *testing.T) {
	req, err := http.NewRequest("GET", "/someurl", nil)
	if err != nil {
		t.Fatal(err)
	}

	var dummyHandler httpMiddleWare = func(writer http.ResponseWriter, request *http.Request) {
		rContext := request.Context()
		correlationID := rContext.Value(CorrelationID)
		assert.NotNil(t, correlationID)
		_, err := uuid.Parse(correlationID.(string))
		assert.NoError(t, err)
		logger := rContext.Value(Logger)
		assert.NotNil(t, logger)
		select {
		case <-time.After(1 * time.Second):
			return
		case <-rContext.Done():
			t.Errorf("timeout shouldn't have been caught")
		}
	}

	rr := httptest.NewRecorder()
	// this bit should finish after 1sec
	handler := http.HandlerFunc(dummyHandler.WithLoggingContext("dummyHandler").WithTimeout(5000))
	handler.ServeHTTP(rr, req)
}
