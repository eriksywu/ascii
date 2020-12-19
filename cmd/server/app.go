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
		writer.WriteHeader(200)
	}).Methods("GET")
}

func (s *appServer) dynamicTimeoutFunc(r *http.Request) int {
	// this could be a bad idea :)
	// i.e if this service runs through any type of layer2/3 router/lb then we'd have to worry about things like tcp-reset
	return 1000
}

func (s *appServer) newImageBaseHandler() httpMiddleWare {
	return func(rw http.ResponseWriter, r *http.Request) {
		baseResponse := models.Response{
			CorrelationID: s.tryGetCorrelationID(r.Context()),
		}
		var uid *uuid.UUID
		var err error
		if r.Header.Get("async") == "true" {
			uid, err = s.service.NewASCIIImageAsync(r.Context(), r.Body)
		} else {
			uid, err = s.service.NewASCIIImageSync(r.Context(), r.Body)
		}
		if err != nil {
			s.writeErrorResponse(err, rw, baseResponse)
		} else if uid == nil {
			s.writeErrorResponse(fmt.Errorf("internal error: could not generate uuid"), rw, baseResponse)
		} else {
			response := models.NewImageResponse{
				Response: baseResponse,
				ImageID:  uid.String(),
			}
			responseBody, _ := json.Marshal(response)
			rw.Write([]byte(responseBody))
		}
	}
}

func (s *appServer) getImageBaseHandler() httpMiddleWare {
	return func(rw http.ResponseWriter, r *http.Request) {
		baseResponse := models.Response{
			CorrelationID: s.tryGetCorrelationID(r.Context()),
		}
		imageId := mux.Vars(r)["imageId"]
		imageUID, err := uuid.Parse(imageId)
		if err != nil {
			s.writeErrorResponse(err, rw, baseResponse)
			return
		}
		finished, image, err := s.service.GetASCIIImage(r.Context(), imageUID)
		if err != nil {
			s.writeErrorResponse(err, rw, baseResponse)
			return
		}
		response := models.GetImageResponse{
			Response:   baseResponse,
			ASCIIValue: string(image),
			Finished:   finished,
		}
		responseBody, _ := json.Marshal(response)
		rw.Write([]byte(responseBody))
	}
}

func (s *appServer) getImageListBaseHandler() httpMiddleWare {
	return func(rw http.ResponseWriter, r *http.Request) {
		baseResponse := models.Response{
			CorrelationID: s.tryGetCorrelationID(r.Context()),
		}
		imageIDList, err := s.service.GetImageList(r.Context())
		if err != nil {
			s.writeErrorResponse(err, rw, baseResponse)
			return
		}
		imageList := make([]string, 0, len(imageIDList))
		for _, id := range imageIDList {
			imageList = append(imageList, id.String())
		}
		response := models.GetImageListResponse{
			Response:    baseResponse,
			ImageIDList: imageList,
		}
		responseBody, _ := json.Marshal(response)
		rw.Write([]byte(responseBody))
	}
}

func (s *appServer) writeErrorResponse(err error, rw http.ResponseWriter, baseResponse models.Response) {
	if errors.Is(err, image.InternalProcessingError{}) {
		rw.WriteHeader(http.StatusInternalServerError)
	} else if errors.Is(err, image.ResourceNotFoundError{}) {
		rw.WriteHeader(http.StatusNotFound)
	} else {
		rw.WriteHeader(http.StatusBadRequest)
	}
	response := models.ErrorResponse{Response: baseResponse, ErrorMessage: err.Error()}
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
