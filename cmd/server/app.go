package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eriksywu/ascii/pkg/image"
	"github.com/eriksywu/ascii/pkg/logging"
	"github.com/eriksywu/ascii/pkg/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

// singleton instance of the ascii image server
var server *appServer

const baseURL = "/images"

// appServer is the rest server for the ascii service
// appServer builds and holds the rest endpoint handlers and constructs logging/context middleware
// appServer's handlers construct the external API responses from ASCIIImageService's output
type appServer struct {
	router  *mux.Router
	port    int
	service ASCIIImageService
	logger  *logging.StandardEventLogger
}

func BuildServer(service ASCIIImageService, port int) *appServer {
	if server == nil {
		server = &appServer{
			port:    port,
			logger:  logging.Logger,
			service: service,
		}
		buildRouter(server)
	}
	return server
}

func (s *appServer) Run() {
	p := strconv.Itoa(s.port)
	s.logger.Printf("Starting service on port %d" , s.port)
	s.logger.Fatal(http.ListenAndServe(":"+p, s.router))
}

func buildRouter (s *appServer) {
	s.logger.Println("registering handlers routes")

	router := mux.NewRouter()

	router.HandleFunc(baseURL, s.newImageBaseHandler().
		WithLoggingContext("newImageHandler").
		WithDynamicTimeout(s.dynamicTimeoutFunc)).
		Methods("POST")

	router.HandleFunc(baseURL, s.getImageListBaseHandler().
		WithLoggingContext("getImageListHandler").
		WithTimeout(30)).
		Methods("GET")

	router.HandleFunc(baseURL+"/{imageId}", s.getImageBaseHandler().
		WithLoggingContext("getImageHandler").
		WithTimeout(60)).
		Methods("GET")

	// Health is contextless
	// Typically timeouts for health checks are specified on the client side (i.e HTTPProbe on K8S)
	router.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
		s.logger.Println("received health check request")
		writer.Write([]byte("running"))
	}).Methods("GET")

	s.router = router
}

func (s *appServer) dynamicTimeoutFunc(r *http.Request) int {
	// this could be a bad idea :)
	// i.e if this service runs through any type of layer2/3 router/lb then we'd have to worry about things like tcp-reset
	return 1000
}

func (s *appServer) newImageBaseHandler() httpMiddleWare {
	return func(rw http.ResponseWriter, r *http.Request) {
		var uid *uuid.UUID
		var err error
		if r.Header.Get("async") == "true" {
			uid, err = s.service.NewASCIIImageAsync(r.Context(), r.Body)
		} else {
			uid, err = s.service.NewASCIIImageSync(r.Context(), r.Body)
		}
		if err != nil {
			s.writeErrorResponse(r.Context(), err, rw)
		} else if uid == nil {
			s.writeErrorResponse(r.Context(), fmt.Errorf("internal error: could not generate uuid"), rw)
		} else {
			response := models.NewImageResponse{
				ImageID:  uid.String(),
			}
			responseBody, _ := json.Marshal(response)
			rw.Write([]byte(responseBody))
		}
	}
}

func (s *appServer) getImageBaseHandler() httpMiddleWare {
	return func(rw http.ResponseWriter, r *http.Request) {
		imageId := mux.Vars(r)["imageId"]
		imageUID, err := uuid.Parse(imageId)
		if err != nil {
			s.writeErrorResponse(r.Context(), err, rw)
			return
		}
		finished, imageBytes, err := s.service.GetASCIIImage(r.Context(), imageUID)
		// if it's an internalprocessingerror, return the error in the response body
		if err != nil && !errors.Is(err, image.InternalProcessingError{}) {
			s.writeErrorResponse(r.Context(), err, rw)
			return
		}
		response := models.GetImageResponse{
			ASCIIValue: string(imageBytes),
			Finished:   finished,
		}
		if err != nil {
			response.ErrorMessage = err.Error()
		}
		responseBody, _ := json.Marshal(response)
		rw.Write([]byte(responseBody))
	}
}

func (s *appServer) getImageListBaseHandler() httpMiddleWare {
	return func(rw http.ResponseWriter, r *http.Request) {
		imageIDList, err := s.service.GetImageList(r.Context())
		if err != nil {
			s.writeErrorResponse(r.Context(), err, rw)
			return
		}
		imageList := make([]string, 0, len(imageIDList))
		for _, id := range imageIDList {
			imageList = append(imageList, id.String())
		}
		response := models.GetImageListResponse{
			ImageIDList: imageList,
		}
		responseBody, _ := json.Marshal(response)
		rw.Write([]byte(responseBody))
	}
}

func (s *appServer) writeErrorResponse(ctx context.Context, err error, rw http.ResponseWriter) {
	switch err.(type) {
	case image.InternalProcessingError:
		rw.WriteHeader(http.StatusInternalServerError)
	case image.ResourceNotFoundError:
		rw.WriteHeader(http.StatusNotFound)
	case image.InvalidInputError:
		rw.WriteHeader(http.StatusBadRequest)
	default:
		rw.WriteHeader(http.StatusBadRequest)
	}
	response := models.ErrorResponse{ErrorMessage: err.Error(), CorrelationID: s.tryGetCorrelationID(ctx)}
	responseBody, _ := json.Marshal(response)
	rw.Write(responseBody)
}

func (s *appServer) tryGetCorrelationID(context context.Context) string {
	correlationIDObject := context.Value(CorrelationID)
	if correlationIDObject == nil {
		return ""
	}
	if correlationID, k := correlationIDObject.(string); !k {
		return ""
	} else {
		return correlationID
	}
}
