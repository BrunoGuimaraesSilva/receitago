package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/BrunoGuimaraesSilva/receitago/internal/downloader/domain/dataset"
	"github.com/rs/zerolog"
)

const (
	ckanPkgPath  = "/ckan/api/3/action/package_show?id="
	tesouroBase  = "https://www.tesourotransparente.gov.br"
	tesouroPkgID = "abb968cb-3710-4f85-89cf-875c91b9c7f6"
)

type ckanResource struct {
	URL          string `json:"url"`
	Name         string `json:"name"`
	LastModified string `json:"last_modified"`
}

type ckanResult struct {
	Resources []ckanResource `json:"resources"`
}

type ckanPkg struct {
	Success bool       `json:"success"`
	Result  ckanResult `json:"result"`
}

type TesouroProvider struct {
	baseDir string
	logger  zerolog.Logger
}

func NewTesouroProvider(baseDir string, logger zerolog.Logger) *TesouroProvider {
	return &TesouroProvider{
		baseDir: baseDir,
		logger:  logger,
	}
}

func (p *TesouroProvider) fetchPackage(baseURL, pkgID string) (*ckanPkg, error) {
	url := baseURL + ckanPkgPath + pkgID
	p.logger.Debug().Str("url", url).Msg("Fetching CKAN package metadata")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s responded with %s", url, resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var pkg ckanPkg
	if err := json.Unmarshal(body, &pkg); err != nil {
		return nil, err
	}
	if !pkg.Success {
		return nil, fmt.Errorf("CKAN API returned success=false")
	}
	return &pkg, nil
}

func (p *TesouroProvider) ListNeeded(ctx context.Context) ([]dataset.Dataset, error) {
	// fetch dataset metadata
	pkg, err := p.fetchPackage(tesouroBase, tesouroPkgID)
	if err != nil {
		p.logger.Error().Err(err).Msg("Failed to fetch Tesouro package")
		return nil, err
	}
	if len(pkg.Result.Resources) == 0 {
		return nil, fmt.Errorf("no resources found in Tesouro package")
	}

	// Find latest modification timestamp for deterministic ID
	latest := pkg.Result.Resources[0].LastModified
	for _, r := range pkg.Result.Resources {
		if r.LastModified > latest {
			latest = r.LastModified
		}
	}

	p.logger.Info().Str("latest", latest).Int("count", len(pkg.Result.Resources)).Msg("Listing Tesouro datasets")

	out := make([]dataset.Dataset, len(pkg.Result.Resources))
	for i, r := range pkg.Result.Resources {
		ts, _ := time.Parse(time.RFC3339, r.LastModified)
		out[i] = dataset.Dataset{
			ID:        fmt.Sprintf("tesouro-%s-%s", latest, r.Name),
			URL:       r.URL,
			Filename:  filepath.Base(r.URL),
			Published: ts,
		}
	}
	return out, nil
}
