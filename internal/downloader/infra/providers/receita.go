package providers

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/BrunoGuimaraesSilva/receitago/internal/downloader/domain/dataset"
	"github.com/rs/zerolog"
)

const (
	receitaUserAgent         = "ReceitaGo/0.0.1"
	lastDownloadedFile       = "last_downloaded.txt"
	httpTimeout              = 60 * time.Second
	federalRevenueURL        = "https://arquivos.receitafederal.gov.br/dados/cnpj/"
	federalRevenueSourcePath = "dados_abertos_cnpj"
	federalRevenueTaxesPath  = "regime_tributario"
)

var (
	yearMonthPattern = regexp.MustCompile(`href="(\d{4}-\d{2}/)"`)
	filePattern      = regexp.MustCompile(`href="(\w+\d?\.zip)"`)
	taxFilePattern   = regexp.MustCompile(`href="((Imune|Lucro).+\.zip)"`)
)

type ReceitaProvider struct {
	client  *http.Client
	baseDir string
	logger  zerolog.Logger
}

func NewReceitaProvider(baseDir string, logger zerolog.Logger) *ReceitaProvider {
	return &ReceitaProvider{
		client:  &http.Client{Timeout: httpTimeout},
		baseDir: baseDir,
		logger:  logger,
	}
}

func (p *ReceitaProvider) get(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("User-Agent", receitaUserAgent)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http get %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http %s: %s", url, resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	return string(b), nil
}

func (p *ReceitaProvider) mostRecentFolder(ctx context.Context, url string) (string, string, error) {
	p.logger.Debug().Str("url", url).Msg("Fetching folder list")

	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	b, err := p.get(ctx, url)
	if err != nil {
		return "", "", err
	}
	var batches []string
	for _, m := range yearMonthPattern.FindAllStringSubmatch(b, -1) {
		batches = append(batches, m[1])
	}
	slices.Sort(batches)
	if len(batches) == 0 {
		return "", "", fmt.Errorf("no batches found in %s", url)
	}
	latest := batches[len(batches)-1]
	p.logger.Info().Str("latest", latest).Msg("Found most recent CNPJ batch")
	return strings.TrimSuffix(latest, "/"), url + latest, nil
}

func (p *ReceitaProvider) taxRegimeURLs(ctx context.Context, url string) ([]string, error) {
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	b, err := p.get(ctx, url)
	if err != nil {
		return nil, err
	}
	var urls []string
	for _, m := range taxFilePattern.FindAllStringSubmatch(b, -1) {
		urls = append(urls, url+m[1])
	}
	return urls, nil
}

func (p *ReceitaProvider) sourceURLs(ctx context.Context, baseURL string) (string, []string, error) {
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	folderName, folderURL, err := p.mostRecentFolder(ctx, baseURL+federalRevenueSourcePath)
	if err != nil {
		return "", nil, fmt.Errorf("get most recent folder: %w", err)
	}

	if p.isAlreadyDownloaded(folderName) {
		return "", nil, fmt.Errorf("latest base already downloaded: %s", folderName)
	}
	if err := p.saveLastDownloaded(folderName); err != nil {
		return "", nil, fmt.Errorf("save last downloaded: %w", err)
	}

	b, err := p.get(ctx, folderURL)
	if err != nil {
		return "", nil, fmt.Errorf("list files: %w", err)
	}
	var urls []string
	for _, m := range filePattern.FindAllStringSubmatch(b, -1) {
		urls = append(urls, folderURL+m[1])
	}

	ts, err := p.taxRegimeURLs(ctx, baseURL+federalRevenueTaxesPath)
	if err != nil {
		return "", nil, fmt.Errorf("list tax regime: %w", err)
	}
	urls = append(urls, ts...)

	return folderName, urls, nil
}

func (p *ReceitaProvider) saveLastDownloaded(name string) error {
	path := filepath.Join(p.baseDir, lastDownloadedFile)
	if err := os.MkdirAll(p.baseDir, 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	if _, err := w.WriteString(name); err != nil {
		return err
	}
	return w.Flush()
}

func (p *ReceitaProvider) isAlreadyDownloaded(name string) bool {
	path := filepath.Join(p.baseDir, lastDownloadedFile)
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	last := strings.TrimSpace(string(b))
	return last == name
}

func (p *ReceitaProvider) ListNeeded(ctx context.Context) ([]dataset.Dataset, error) {
	folderName, urls, err := p.sourceURLs(ctx, federalRevenueURL)
	if err != nil {
		return nil, err
	}

	// Parse YYYY-MM format to time.Time
	publishedDate, err := time.Parse("2006-01", folderName)
	if err != nil {
		p.logger.Warn().Str("folderName", folderName).Err(err).Msg("Failed to parse published date, using zero time")
	}

	p.logger.Info().Int("count", len(urls)).Str("batch", folderName).Msg("Listing Receita datasets")

	out := make([]dataset.Dataset, len(urls))
	for i, u := range urls {
		out[i] = dataset.Dataset{
			ID:        fmt.Sprintf("receita-%s-%s", folderName, filepath.Base(u)),
			URL:       u,
			Filename:  filepath.Base(u),
			Published: publishedDate,
		}
	}
	return out, nil
}
