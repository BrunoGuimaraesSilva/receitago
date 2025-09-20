package download

import (
	"context"
	"fmt"
	"time"

	"github.com/BrunoGuimaraesSilva/receitago/services/downloader/domain/dataset"
)

type Result struct {
	ID           string        `json:"id"`
	Filename     string        `json:"filename"`
	Success      bool          `json:"success"`
	Error        string        `json:"error,omitempty"`
	DownloadTime time.Duration `json:"download_time"`
	UnzipTime    time.Duration `json:"unzip_time"`
	Attempts     int           `json:"attempts"`
}

type Interactor struct {
	Provider   dataset.DatasetProvider
	Downloader dataset.DownloaderPort
	Filestorer dataset.FilestorerPort

	MaxRetries int
	RetryDelay time.Duration
}

func NewInteractor(p dataset.DatasetProvider, d dataset.DownloaderPort, f dataset.FilestorerPort, maxRetries int, retryDelay time.Duration) *Interactor {
	return &Interactor{
		Provider:   p,
		Downloader: d,
		Filestorer: f,
		MaxRetries: maxRetries,
		RetryDelay: retryDelay,
	}
}

func (uc *Interactor) Run(ctx context.Context) ([]Result, error) {
	items, err := uc.Provider.ListNeeded(ctx)
	if err != nil {
		return nil, fmt.Errorf("list datasets: %w", err)
	}

	results := make([]Result, 0, len(items))
	for _, ds := range items {
		res := Result{ID: ds.ID, Filename: ds.Filename}

		var lastErr error
		for attempt := 1; attempt <= uc.MaxRetries; attempt++ {
			startDownload := time.Now()
			r, err := uc.Downloader.Download(ctx, ds.URL)
			downloadElapsed := time.Since(startDownload)

			if err != nil {
				lastErr = fmt.Errorf("download failed: %w", err)
			} else {
				startUnzip := time.Now()
				if err := uc.Filestorer.Save(ctx, ds.Filename, r); err != nil {
					lastErr = fmt.Errorf("save failed: %w", err)
				} else {
					res.Success = true
					lastErr = nil
					res.DownloadTime = downloadElapsed
					res.UnzipTime = time.Since(startUnzip)
					res.Attempts = attempt
					_ = r.Close()
					break
				}
				res.DownloadTime = downloadElapsed
				res.UnzipTime = time.Since(startUnzip)
				_ = r.Close()
			}

			if attempt < uc.MaxRetries {
				time.Sleep(uc.RetryDelay)
			}
		}

		if lastErr != nil {
			res.Success = false
			res.Error = lastErr.Error()
			res.Attempts = uc.MaxRetries
		}

		results = append(results, res)
	}
	return results, nil
}
