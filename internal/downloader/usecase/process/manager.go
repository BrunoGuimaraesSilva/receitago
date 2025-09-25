package process

import (
	"context"
	"fmt"

	"github.com/BrunoGuimaraesSilva/receitago/internal/downloader/domain/dataset"
	"github.com/BrunoGuimaraesSilva/receitago/internal/downloader/usecase/download"
)

type Manager struct {
	Steps      []Step
	Downloader dataset.DownloaderPort
	Filestorer dataset.FilestorerPort
}

func NewManager(steps []Step, d dataset.DownloaderPort, f dataset.FilestorerPort) *Manager {
	return &Manager{
		Steps:      steps,
		Downloader: d,
		Filestorer: f,
	}
}

func (m *Manager) Run(ctx context.Context) (Report, error) {
	report := Report{Steps: make([]StepReport, 0, len(m.Steps))}
	allOK := true

	for _, st := range m.Steps {
		stepReport := m.runStep(ctx, st)
		report.Steps = append(report.Steps, stepReport)

		if !stepReport.Success {
			allOK = false
			return report, fmt.Errorf("step %q failed: %s", st.Name, stepReport.Error)
		}
	}

	report.Success = allOK
	return report, nil
}

func (m *Manager) runStep(ctx context.Context, st Step) StepReport {
	uc := download.NewInteractor(st.Provider, m.Downloader, m.Filestorer, st.MaxRetries, st.RetryDelay)
	files, err := uc.Run(ctx)

	srep := StepReport{
		Name:     st.Name,
		FileRuns: files,
		Success:  true,
	}

	if err != nil {
		srep.Success = false
		srep.Error = err.Error()
		return srep
	}

	for _, f := range files {
		if !f.Success {
			srep.Success = false
			if srep.Error == "" {
				srep.Error = fmt.Sprintf("file %s failed: %s", f.Filename, f.Error)
			}
		}
	}

	return srep
}
