package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/BrunoGuimaraesSilva/receitago/pkg/iox"
)

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

// Download fetches a file over HTTP and stores it in a temporary file.
func (d *HTTPDownloader) Download(ctx context.Context, url string) (iox.ReadSeekCloser, error) {
	ctx, cancel := context.WithTimeout(ctx, d.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	if d.UserAgent != "" {
		req.Header.Set("User-Agent", d.UserAgent)
	}

	resp, err := d.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	tf, err := os.CreateTemp("", "receitago-*")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}

	if _, err := io.Copy(tf, resp.Body); err != nil {
		tf.Close()
		os.Remove(tf.Name())
		return nil, fmt.Errorf("copy response: %w", err)
	}

	if _, err := tf.Seek(0, 0); err != nil {
		tf.Close()
		os.Remove(tf.Name())
		return nil, fmt.Errorf("seek temp file: %w", err)
	}

	return &tempFile{tf}, nil
}
