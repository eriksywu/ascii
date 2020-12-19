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
	Finished bool
}

type GetImageListResponse struct {
	Response
	ImageIDList []string
}
