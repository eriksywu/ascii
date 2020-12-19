package server

import (
	"context"
	"github.com/eriksywu/ascii/pkg/logging"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)
// singleton instance of the ascii app server
var server *appServer

// appServer is a simple rest controller for the ascii service
type appServer struct {
	router  *mux.Router
	port    int
	service ASCIIImageService
	logger  *logging.StandardEventLogger
}

// Portions of skeleton http/router setup code was lifted from my own personal project
// https://github.com/eriksywu/yamlr/blob/master/server/server.go
func BuildServer(service ASCIIImageService, port int) *appServer {
	if server == nil {
		server = &appServer{
			port: port,
			logger: logging.Logger,
			service: service,
		}
		server.buildRouter()
	}
	return server
}

func (s *appServer) Run() {
	p := strconv.Itoa(s.port)
	s.logger.Println("Starting service...")
	s.logger.Fatal(http.ListenAndServe(":"+p, s.router))
}

// TODO: pass in builder opts to specify which middleware to use?
func (s *appServer) buildRouter() {

	s.logger.Println("registering handlers routes")

	router := mux.NewRouter()

	router.HandleFunc("/api/image", s.newImageBaseHandler().
		WithLoggingContext("newImageHandler").
		WithTimeout(120)).
		Methods("POST")

	router.HandleFunc("/api/image", s.getImageBaseHandler().
		WithLoggingContext("newImageHandler").
		WithTimeout(60)).
		Methods("GET")
	//router.HandleFunc("/api/search", s.getImageHandler).Methods("GET")
}

type wrappedHandlerFunc http.HandlerFunc

const CorrelationID = "correlationID"
const Logger = "logger"

// WithLoggingAndTimeContext is a simple middleware to inject a correlationID and logrus entry into the incoming request
// TODO handle timeout context somehow?
func (w wrappedHandlerFunc) WithLoggingContext(operationName string) wrappedHandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		correlationID := uuid.New()
		rContext := context.WithValue(r.Context(), CorrelationID, correlationID)
		logEntry := logrus.WithFields(logrus.Fields{
			"correlationID": correlationID,
			"operation": operationName,
		})

		rContext = context.WithValue(rContext, Logger, logEntry)
		w(rw, r.WithContext(rContext))
	}
}

func (w wrappedHandlerFunc) WithTimeout(timeoutSeconds int) wrappedHandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rContext, cancelFunc := context.WithCancel(r.Context())
		done := make(chan struct{})
		go func() {
			w(rw, r.WithContext(rContext))
			done <- struct{}{}
		}()
		select {
		case <- time.After(time.Duration(timeoutSeconds) * time.Second):
			cancelFunc()
			// TODO implement timeout response
			return
		case <- done:
			return
		}
	}
}

func(s *appServer) newImageBaseHandler() wrappedHandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

	}
}

func(s *appServer) getImageBaseHandler() wrappedHandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

	}
}