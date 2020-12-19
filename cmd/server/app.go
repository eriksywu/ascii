package server

import (
	"context"
	"encoding/json"
	"github.com/eriksywu/ascii/pkg/logging"
	"github.com/eriksywu/ascii/pkg/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

// singleton instance of the ascii app server
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

// Portions of skeleton http/router setup code was lifted from my own personal project
// https://github.com/eriksywu/yamlr/blob/master/server/server.go
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
		WithTimeout(120)).
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

func (s *appServer) newImageBaseHandler() httpMiddleWare {
	return func(rw http.ResponseWriter, r *http.Request) {
		baseResponse := models.Response{
			CorrelationID: s.tryGetCorrelationID(r.Context()),
		}
		uid, err := s.service.NewASCIIImage(r.Body)
		if err != nil {
			s.writeErrorResponse(err, rw, baseResponse)
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
		image, err := s.service.GetASCIIImage(imageUID)
		if err != nil {
			s.writeErrorResponse(err, rw, baseResponse)
			return
		}
		response := models.GetImageResponse{
			Response:   baseResponse,
			ASCIIValue: string(image),
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
		imageIDList, err := s.service.GetImageList()
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

// TODO implement
func (s *appServer) writeErrorResponse(err error, rw http.ResponseWriter, baseResponse models.Response) {
	rw.WriteHeader(http.StatusBadRequest)
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
