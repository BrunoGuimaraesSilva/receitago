package download

import "time"

type DownloadErrorType string

const (
	ErrDownload DownloadErrorType = "download"
	ErrSave     DownloadErrorType = "save"
	ErrUnknown  DownloadErrorType = "unknown"
)

type ErrorDetail struct {
	Type    DownloadErrorType `json:"type"`
	Message string            `json:"message"`
}

type Result struct {
	ID           string        `json:"id"`
	Filename     string        `json:"filename"`
	Success      bool          `json:"success"`
	Error        *ErrorDetail  `json:"error,omitempty"`
	DownloadTime time.Duration `json:"download_time"`
	UnzipTime    time.Duration `json:"unzip_time"`
	Attempts     int           `json:"attempts"`
}
