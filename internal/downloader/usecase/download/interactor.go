package download

import (
	"context"
	"fmt"
	"time"

	"github.com/BrunoGuimaraesSilva/receitago/internal/downloader/domain/dataset"
	"github.com/BrunoGuimaraesSilva/receitago/pkg/iox"
)

type Interactor struct {
	Provider   dataset.DatasetProvider
	Downloader dataset.DownloaderPort
	Filestorer dataset.FilestorerPort

	MaxRetries int
	RetryDelay time.Duration
}

func NewInteractor(p dataset.DatasetProvider, d dataset.DownloaderPort, f dataset.FilestorerPort, maxRetries int, retryDelay time.Duration) *Interactor {
	return &Interactor{Provider: p, Downloader: d, Filestorer: f, MaxRetries: maxRetries, RetryDelay: retryDelay}
}

func (uc *Interactor) Run(ctx context.Context) ([]Result, error) {
	items, err := uc.Provider.ListNeeded(ctx)
	if err != nil {
		return nil, fmt.Errorf("list datasets: %w", err)
	}

	results := make([]Result, 0, len(items))
	for _, ds := range items {
		results = append(results, uc.runDataset(ctx, ds))
	}
	return results, nil
}

func (uc *Interactor) runDataset(ctx context.Context, ds dataset.Dataset) Result {
	res := Result{ID: ds.ID, Filename: ds.Filename}
	var lastErr *ErrorDetail

	for attempt := 1; attempt <= uc.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			res.Error = &ErrorDetail{Type: ErrUnknown, Message: "context cancelled"}
			res.Attempts = attempt
			return res
		default:
		}

		startDownload := time.Now()
		r, err := uc.Downloader.Download(ctx, ds.URL)
		res.DownloadTime = time.Since(startDownload)

		if err != nil {
			lastErr = &ErrorDetail{Type: ErrDownload, Message: err.Error()}
		} else {
			startSave := time.Now()
			saveErr := uc.Filestorer.Save(ctx, ds.Filename, r.(iox.ReadSeekCloser))
			_ = r.Close()

			if saveErr != nil {
				lastErr = &ErrorDetail{Type: ErrSave, Message: saveErr.Error()}
				res.UnzipTime = time.Since(startSave)
			} else {
				res.Success = true
				res.UnzipTime = time.Since(startSave)
				res.Attempts = attempt
				return res
			}
		}

		if attempt < uc.MaxRetries {
			select {
			case <-ctx.Done():
				return res
			case <-time.After(uc.RetryDelay):
			}
		}
	}

	res.Success = false
	res.Error = lastErr
	res.Attempts = uc.MaxRetries
	return res
}
