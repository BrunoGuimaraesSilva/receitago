package providers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BrunoGuimaraesSilva/receitago/services/downloader/domain/dataset"
)

const (
	ckanPkgPath  = "/ckan/api/3/action/package_show?id="
	tesouroBase  = "https://www.tesourotransparente.gov.br"
	tesouroPkgID = "abb968cb-3710-4f85-89cf-875c91b9c7f6"
	lastFile     = "tesouro_last_downloaded.txt"
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
	client  *http.Client
}

func NewTesouroProvider(baseDir string) *TesouroProvider {
	return &TesouroProvider{
		baseDir: baseDir,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *TesouroProvider) fetchPackage(baseURL, pkgID string) (*ckanPkg, error) {
	url := baseURL + ckanPkgPath + pkgID
	resp, err := p.client.Get(url)
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

func (p *TesouroProvider) saveLastDownloaded(ts string) error {
	path := filepath.Join(p.baseDir, lastFile)
	if err := os.MkdirAll(p.baseDir, 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	if _, err := w.WriteString(ts); err != nil {
		return err
	}
	return w.Flush()
}

func (p *TesouroProvider) isAlreadyDownloaded(ts string) bool {
	path := filepath.Join(p.baseDir, lastFile)
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	last := strings.TrimSpace(string(b))
	return last == ts
}

func (p *TesouroProvider) ListNeeded(ctx context.Context) ([]dataset.Dataset, error) {
	// fetch dataset metadata
	pkg, err := p.fetchPackage(tesouroBase, tesouroPkgID)
	if err != nil {
		return nil, err
	}
	if len(pkg.Result.Resources) == 0 {
		return nil, fmt.Errorf("no resources found in Tesouro package")
	}

	latest := pkg.Result.Resources[0].LastModified
	for _, r := range pkg.Result.Resources {
		if r.LastModified > latest {
			latest = r.LastModified
		}
	}

	if p.isAlreadyDownloaded(latest) {
		return nil, fmt.Errorf("Latest Tesouro dataset already downloaded: %s", latest)
	}
	if err := p.saveLastDownloaded(latest); err != nil {
		return nil, err
	}

	out := make([]dataset.Dataset, len(pkg.Result.Resources))
	for i, r := range pkg.Result.Resources {
		ts, _ := time.Parse(time.RFC3339, r.LastModified)
		out[i] = dataset.Dataset{
			ID:        fmt.Sprintf("tesouro-%s-%d", latest, i),
			URL:       r.URL,
			Filename:  filepath.Base(r.URL),
			Published: ts,
		}
	}
	return out, nil
}
