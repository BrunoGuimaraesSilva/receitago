package storage

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/BrunoGuimaraesSilva/receitago/services/downloader/domain/dataset"
)

type SmartFilestorer struct {
	BaseDir string
}

func NewSmartFilestorer(baseDir string) *SmartFilestorer {
	return &SmartFilestorer{BaseDir: baseDir}
}

func (s *SmartFilestorer) Save(ctx context.Context, name string, r dataset.ReadSeekCloser) error {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == ".zip" {
		return s.saveUnzipped(ctx, name, r)
	}
	return s.saveRaw(ctx, name, r)
}

func (s *SmartFilestorer) saveRaw(ctx context.Context, name string, r dataset.ReadSeekCloser) error {
	if _, err := r.Seek(0, 0); err != nil {
		return err
	}
	outDir := filepath.Join(s.BaseDir, "tesouro")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	outPath := filepath.Join(outDir, name)

	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, r)
	return err
}

func (s *SmartFilestorer) saveUnzipped(ctx context.Context, name string, r dataset.ReadSeekCloser) error {
	if _, err := r.Seek(0, 0); err != nil {
		return err
	}

	tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("receitago-%s", name))
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(tmpFile, r); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	zr, err := zip.OpenReader(tmpPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer zr.Close()

	if len(zr.File) == 0 {
		return fmt.Errorf("empty zip %s", name)
	}

	outDir := filepath.Join(s.BaseDir, "receita")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		outPath := filepath.Join(outDir, f.Name)
		outFile, err := os.Create(outPath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(outFile, rc); err != nil {
			outFile.Close()
			return err
		}
		outFile.Close()
	}

	return nil
}
