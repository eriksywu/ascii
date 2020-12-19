package models

type Response struct {
	CorrelationID string
}

type ErrorResponse struct {
	Response
	ErrorMessage string
}

type NewImageResponse struct {
	Response
	ImageID string
}

type GetImageResponse struct {
	Response
	ASCIIValue string
}

type GetImageListResponse struct {
	Response
	ImageIDList []string
}
