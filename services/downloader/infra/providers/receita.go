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

	"github.com/BrunoGuimaraesSilva/receitago/services/downloader/domain/dataset"
)

const (
	receitaUserAgent         = "ReceitaGo/0.0.1"
	lastDownloadedFile       = "last_downloaded.txt"
	federalRevenueURL        = "https://arquivos.receitafederal.gov.br/dados/cnpj/"
	federalRevenueSourcePath = "dados_abertos_cnpj"
	federalRevenueTaxesPath  = "regime_tributario"
)

var yearMonthPattern = regexp.MustCompile(`href="(\d{4}-\d{2}/)"`)
var filePattern = regexp.MustCompile(`href="(\w+\d?\.zip)"`)
var taxFilePattern = regexp.MustCompile(`href="((Imune|Lucro).+\.zip)"`)

type ReceitaProvider struct {
	client  *http.Client
	baseDir string
}

func NewReceitaProvider(baseDir string) *ReceitaProvider {
	return &ReceitaProvider{
		client:  &http.Client{Timeout: 60 * time.Second},
		baseDir: baseDir,
	}
}

func (p *ReceitaProvider) get(url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", receitaUserAgent)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s responded with %s", url, resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (p *ReceitaProvider) mostRecentFolder(url string) (string, string, error) {
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	b, err := p.get(url)
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
	return strings.TrimSuffix(latest, "/"), url + latest, nil
}

func (p *ReceitaProvider) taxRegimeURLs(url string) ([]string, error) {
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	b, err := p.get(url)
	if err != nil {
		return nil, err
	}
	var urls []string
	for _, m := range taxFilePattern.FindAllStringSubmatch(b, -1) {
		urls = append(urls, url+m[1])
	}
	return urls, nil
}

func (p *ReceitaProvider) sourceURLs(baseURL string) (string, []string, error) {
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	folderName, folderURL, err := p.mostRecentFolder(baseURL + federalRevenueSourcePath)
	if err != nil {
		return "", nil, err
	}

	if p.isAlreadyDownloaded(folderName) {
		return "", nil, fmt.Errorf("Latest base was already downloaded: %s", folderName)
	}
	if err := p.saveLastDownloaded(folderName); err != nil {
		return "", nil, err
	}

	b, err := p.get(folderURL)
	if err != nil {
		return "", nil, err
	}
	var urls []string
	for _, m := range filePattern.FindAllStringSubmatch(b, -1) {
		urls = append(urls, folderURL+m[1])
	}

	ts, err := p.taxRegimeURLs(baseURL + federalRevenueTaxesPath)
	if err != nil {
		return "", nil, err
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
	folderName, urls, err := p.sourceURLs(federalRevenueURL)
	if err != nil {
		return nil, err
	}

	out := make([]dataset.Dataset, len(urls))
	for i, u := range urls {
		out[i] = dataset.Dataset{
			ID:       fmt.Sprintf("receita-%s-%d", folderName, i),
			URL:      u,
			Filename: filepath.Base(u),
		}
	}
	return out, nil
}
