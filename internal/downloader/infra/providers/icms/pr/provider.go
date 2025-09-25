package pr

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BrunoGuimaraesSilva/receitago/internal/downloader/domain/dataset"
)

const (
	ativosURL     = "http://processos.fazenda.pr.gov.br/arquivos/ativos"
	canceladosURL = "http://processos.fazenda.pr.gov.br/arquivos/cancelados"

	lastFile = "icmspr_last_downloaded.txt"
)

type Provider struct {
	baseDir string
	client  *http.Client
}

func NewProvider(baseDir string) *Provider {
	return &Provider{
		baseDir: baseDir,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *Provider) saveLastDownloaded(ts string) error {
	path := filepath.Join(p.baseDir, lastFile)
	if err := os.MkdirAll(p.baseDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(ts), 0o644)
}

func (p *Provider) isAlreadyDownloaded(ts string) bool {
	path := filepath.Join(p.baseDir, lastFile)
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(b)) == ts
}

func (p *Provider) ListNeeded(ctx context.Context) ([]dataset.Dataset, error) {
	today := time.Now().Format("2006-01-02")

	if p.isAlreadyDownloaded(today) {
		return nil, fmt.Errorf("ICMS PR already downloaded today: %s", today)
	}
	if err := p.saveLastDownloaded(today); err != nil {
		return nil, err
	}

	urls := []string{ativosURL, canceladosURL}
	out := make([]dataset.Dataset, len(urls))
	for i, u := range urls {
		out[i] = dataset.Dataset{
			ID:        fmt.Sprintf("icmspr-%s-%d", today, i),
			URL:       u,
			Filename:  filepath.Base(u),
			Published: time.Now(),
		}
	}
	return out, nil
}
