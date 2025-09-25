package process

import (
	"time"

	"github.com/BrunoGuimaraesSilva/receitago/internal/downloader/domain/dataset"
	"github.com/BrunoGuimaraesSilva/receitago/internal/downloader/usecase/download"
)

// Step defines a single ingestion step in the pipeline
type Step struct {
	Name       string
	Provider   dataset.DatasetProvider
	MaxRetries int
	RetryDelay time.Duration
}

// StepReport contains the result of a single step execution
type StepReport struct {
	Name     string            `json:"name"`
	Success  bool              `json:"success"`
	Error    string            `json:"error,omitempty"`
	FileRuns []download.Result `json:"files"`
}

// Report aggregates results for the entire pipeline
type Report struct {
	Steps   []StepReport `json:"steps"`
	Success bool         `json:"success"`
}
