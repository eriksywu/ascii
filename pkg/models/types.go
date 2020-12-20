package models

type ErrorResponse struct {
	ErrorMessage  string
	CorrelationID string
}

type NewImageResponse struct {
	ImageID string
}

type GetImageResponse struct {
	ASCIIValue   string
	Finished     bool
	ErrorMessage string
}

type GetImageListResponse struct {
	ImageIDList []string
}
