package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/BrunoGuimaraesSilva/receitago/pkg/iox"
)

type ChunkDownloader struct {
	Client        *http.Client
	Timeout       time.Duration
	ChunkSize     int64
	MaxRetries    int
	PrintProgress bool
}

func NewChunkDownloader(timeout time.Duration, chunkSizeMB int, maxRetries int, printProgress bool) *ChunkDownloader {
	return &ChunkDownloader{
		Client:        &http.Client{},
		Timeout:       timeout,
		ChunkSize:     int64(chunkSizeMB) * 1024 * 1024,
		MaxRetries:    maxRetries,
		PrintProgress: printProgress,
	}
}

func (d *ChunkDownloader) Download(ctx context.Context, url string) (iox.ReadSeekCloser, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := d.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("head request failed: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HEAD %s failed with %s", url, resp.Status)
	}
	size := resp.ContentLength
	if size <= 0 {
		return nil, fmt.Errorf("invalid content length for %s", url)
	}

	tmpPath := filepath.Join(os.TempDir(), "receitago-"+filepath.Base(url))
	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, err
	}

	var downloaded int64
	filename := filepath.Base(url)

	for downloaded < size {
		start := downloaded
		end := start + d.ChunkSize - 1
		if end >= size {
			end = size - 1
		}

		rangeHeader := fmt.Sprintf("bytes=%d-%d", start, end)
		var lastErr error

		for attempt := 1; attempt <= d.MaxRetries; attempt++ {
			select {
			case <-ctx.Done():
				tmpFile.Close()
				return nil, ctx.Err()
			default:
			}

			chunkReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			chunkReq.Header.Set("Range", rangeHeader)

			chunkResp, err := d.Client.Do(chunkReq)
			if err != nil {
				lastErr = err
				d.waitRetry(ctx, attempt)
				continue
			}
			if chunkResp.StatusCode != http.StatusPartialContent && chunkResp.StatusCode != http.StatusOK {
				lastErr = fmt.Errorf("unexpected status %s", chunkResp.Status)
				chunkResp.Body.Close()
				d.waitRetry(ctx, attempt)
				continue
			}

			if _, err := tmpFile.Seek(start, 0); err != nil {
				chunkResp.Body.Close()
				tmpFile.Close()
				return nil, err
			}

			n, err := io.Copy(tmpFile, chunkResp.Body)
			chunkResp.Body.Close()
			if err != nil {
				lastErr = fmt.Errorf("copy failed: %w", err)
				d.waitRetry(ctx, attempt)
				continue
			}

			downloaded += n
			lastErr = nil

			if d.PrintProgress {
				percent := float64(downloaded) / float64(size) * 100
				fmt.Printf("[%s] %.1f%% downloaded (%s/%s)\n",
					filename,
					percent,
					humanSize(downloaded),
					humanSize(size),
				)
			}
			break
		}

		if lastErr != nil {
			tmpFile.Close()
			return nil, fmt.Errorf("failed chunk %s after %d retries: %w", rangeHeader, d.MaxRetries, lastErr)
		}
	}

	if _, err := tmpFile.Seek(0, 0); err != nil {
		tmpFile.Close()
		return nil, err
	}

	return &tempFile{tmpFile}, nil
}

// waitRetry pauses before retrying, respecting context
func (d *ChunkDownloader) waitRetry(ctx context.Context, attempt int) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(time.Second * time.Duration(attempt)):
	}
}

func humanSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(b)/float64(div), "KMGTPE"[exp])
}
