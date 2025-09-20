package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/BrunoGuimaraesSilva/receitago/services/downloader/domain/dataset"
)

type tempFile struct{ *os.File }

func (t *tempFile) Close() error { return t.File.Close() }

type HTTPDownloader struct {
	Client    *http.Client
	Timeout   time.Duration
	UserAgent string
}

func NewHTTPDownloader(timeout time.Duration, userAgent string) *HTTPDownloader {
	return &HTTPDownloader{
		Client:    &http.Client{},
		Timeout:   timeout,
		UserAgent: userAgent,
	}
}

func (d *HTTPDownloader) Download(ctx context.Context, url string) (dataset.ReadSeekCloser, error) {
	ctx, cancel := context.WithTimeout(ctx, d.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if d.UserAgent != "" {
		req.Header.Set("User-Agent", d.UserAgent)
	}

	resp, err := d.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	tf, err := os.CreateTemp("", "receitago-*")
	if err != nil {
		resp.Body.Close()
		return nil, err
	}

	if _, err := io.Copy(tf, resp.Body); err != nil {
		resp.Body.Close()
		tf.Close()
		os.Remove(tf.Name())
		return nil, err
	}
	resp.Body.Close()

	if _, err := tf.Seek(0, 0); err != nil {
		return nil, err
	}
	return &tempFile{tf}, nil
}
