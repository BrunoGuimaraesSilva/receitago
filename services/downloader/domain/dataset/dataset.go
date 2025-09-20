package dataset

import (
	"context"
	"io"
	"time"
)

type Dataset struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Filename  string    `json:"filename"`
	Published time.Time `json:"published,omitempty"`
}

type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

type DatasetProvider interface {
	ListNeeded(ctx context.Context) ([]Dataset, error)
}

type DownloaderPort interface {
	Download(ctx context.Context, url string) (ReadSeekCloser, error)
}

type FilestorerPort interface {
	Save(ctx context.Context, name string, r ReadSeekCloser) error
}
