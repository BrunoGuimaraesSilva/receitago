package process

import (
	"context"
	"fmt"
	"time"

	"github.com/BrunoGuimaraesSilva/receitago/services/downloader/domain/dataset"
	"github.com/BrunoGuimaraesSilva/receitago/services/downloader/usecase/download"
)

type Step struct {
	Name       string
	Provider   dataset.DatasetProvider
	MaxRetries int
	RetryDelay time.Duration
}

type StepReport struct {
	Name    string            `json:"name"`
	Success bool              `json:"success"`
	Error   string            `json:"error,omitempty"`
	Files   []download.Result `json:"files"`
}

type Report struct {
	Steps   []StepReport `json:"steps"`
	Success bool         `json:"success"`
}

type Manager struct {
	Steps      []Step
	Downloader dataset.DownloaderPort
	Filestorer dataset.FilestorerPort
}

func NewManager(steps []Step, d dataset.DownloaderPort, f dataset.FilestorerPort) *Manager {
	return &Manager{Steps: steps, Downloader: d, Filestorer: f}
}

func (m *Manager) Run(ctx context.Context) (Report, error) {
	r := Report{Steps: make([]StepReport, 0, len(m.Steps))}
	for _, st := range m.Steps {
		uc := download.NewInteractor(st.Provider, m.Downloader, m.Filestorer, st.MaxRetries, st.RetryDelay)
		files, err := uc.Run(ctx)
		srep := StepReport{Name: st.Name, Files: files}

		if err != nil {
			srep.Success = false
			srep.Error = err.Error()
			r.Steps = append(r.Steps, srep)
			r.Success = false
			return r, fmt.Errorf("step %q failed: %w", st.Name, err)
		}

		ok := true
		for _, f := range files {
			if !f.Success {
				ok = false
				if srep.Error == "" {
					srep.Error = fmt.Sprintf("file %s failed: %s", f.Filename, f.Error)
				}
			}
		}
		srep.Success = ok
		r.Steps = append(r.Steps, srep)

		if !ok {
			r.Success = false
			return r, fmt.Errorf("step %q failed", st.Name)
		}
	}
	r.Success = true
	return r, nil
}
