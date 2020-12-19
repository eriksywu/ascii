package server

import (
	"context"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// decorate the existing HandlerFunc with our own middleware on-top of base request handlers for image logic
type httpMiddleWare http.HandlerFunc

const (
	CorrelationID = "correlationID"
	Logger        = "logger"
	Operation     = "operation"
)

// WithLoggingAndTimeContext is a simple middleware to inject some request context info into a logrus entry
func (w httpMiddleWare) WithLoggingContext(operationName string) httpMiddleWare {
	return func(rw http.ResponseWriter, r *http.Request) {
		correlationID := uuid.New().String()
		rContext := context.WithValue(r.Context(), CorrelationID, correlationID)
		logEntry := logrus.WithFields(logrus.Fields{
			CorrelationID: correlationID,
			Operation:     operationName,
		})
		rContext = context.WithValue(rContext, Logger, logEntry)
		w(rw, r.WithContext(rContext))
	}
}

// WithLoggingAndTimeContext is a simple middleware to wrap the request with a timeout context
// this assumes the underlying base handler respects the context.Done channel from the request
func (w httpMiddleWare) WithTimeout(timeoutSeconds int) httpMiddleWare {
	return func(rw http.ResponseWriter, r *http.Request) {
		rContext, cancelFunc := context.WithCancel(r.Context())
		done := make(chan struct{})
		go func() {
			w(rw, r.WithContext(rContext))
			done <- struct{}{}
		}()
		select {
		case <-time.After(time.Duration(timeoutSeconds) * time.Second):
			cancelFunc()
			// TODO implement timeout response
			return
		case <-done:
			return
		}
	}
}
