package dataset

import (
	"context"
	"time"

	"github.com/BrunoGuimaraesSilva/receitago/pkg/iox"
)

type Dataset struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Filename  string    `json:"filename"`
	Published time.Time `json:"published,omitempty"`
}

type DatasetProvider interface {
	ListNeeded(ctx context.Context) ([]Dataset, error)
}

type DownloaderPort interface {
	Download(ctx context.Context, url string) (iox.ReadSeekCloser, error)
}

type FilestorerPort interface {
	Save(ctx context.Context, name string, r iox.ReadSeekCloser) error
}
