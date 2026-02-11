package models

import "github.com/BrunoGuimaraesSilva/receitago/internal/downloader/usecase/download"

type SuccessResponse struct {
	Status  string `json:"status" example:"success"`
	Message string `json:"message" example:"Operation completed successfully"`
}
type ErrorResponse struct {
	Error   string        `json:"error" example:"Internal Server Error"`
	Message string        `json:"message,omitempty" example:"Failed to process request"`
	Details []ErrorDetail `json:"details,omitempty"`
}

type ErrorDetail struct {
	Field   string `json:"field,omitempty" example:"year"`
	Message string `json:"message" example:"Invalid year format"`
}
type UnauthorizedResponse struct {
	Error string `json:"error" example:"Unauthorized"`
}

type BadRequestResponse struct {
	Error   string `json:"error" example:"Bad Request"`
	Message string `json:"message" example:"Invalid request parameters"`
}

type NotFoundResponse struct {
	Error   string `json:"error" example:"Not Found"`
	Message string `json:"message" example:"Resource not found"`
}

type HealthResponse struct {
	Status    string `json:"status" example:"healthy"`
	Timestamp string `json:"timestamp" example:"2026-02-10T21:53:00Z"`
	Version   string `json:"version" example:"1.0.0"`
}

type DownloadResults struct {
	Results []download.Result `json:"results"`
}
