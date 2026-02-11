package pr

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/BrunoGuimaraesSilva/receitago/internal/downloader/domain/dataset"
)

const (
	ativosURL     = "http://processos.fazenda.pr.gov.br/arquivos/ativos"
	canceladosURL = "http://processos.fazenda.pr.gov.br/arquivos/cancelados"
)

type Provider struct {
	baseDir string
}

func NewProvider(baseDir string) *Provider {
	return &Provider{
		baseDir: baseDir,
	}
}

func (p *Provider) ListNeeded(ctx context.Context) ([]dataset.Dataset, error) {
	today := time.Now().Format("2006-01-02")

	urls := []struct {
		url  string
		kind string
	}{
		{ativosURL, "ativos"},
		{canceladosURL, "cancelados"},
	}

	out := make([]dataset.Dataset, 0, len(urls))
	for _, item := range urls {
		id := fmt.Sprintf("icmspr-%s-%s", item.kind, today)

		out = append(out, dataset.Dataset{
			ID:        id,
			URL:       item.url,
			Filename:  filepath.Base(item.url),
			Published: time.Now(),
		})
	}
	return out, nil
}
