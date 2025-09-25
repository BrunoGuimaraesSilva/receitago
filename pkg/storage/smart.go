package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/BrunoGuimaraesSilva/receitago/pkg/iox"
	"github.com/BrunoGuimaraesSilva/receitago/pkg/storage/local"
)

type SmartFilestorer struct {
	BaseDir string
	FS      local.FileWriter
	ZR      ZipReaderFactory
}

func NewSmartFilestorer(baseDir string, fs local.FileWriter, zr ZipReaderFactory) *SmartFilestorer {
	return &SmartFilestorer{
		BaseDir: baseDir,
		FS:      fs,
		ZR:      zr,
	}
}

func (s *SmartFilestorer) Save(ctx context.Context, name string, r iox.ReadSeekCloser) error {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == ".zip" {
		return s.saveZipAndUnzip(ctx, name, r)
	}
	return s.saveRaw(ctx, name, r)
}

func (s *SmartFilestorer) saveRaw(ctx context.Context, name string, r iox.ReadSeekCloser) error {
	if _, err := r.Seek(0, 0); err != nil {
		return err
	}

	outDir := filepath.Join(s.BaseDir, "tesouro")
	if err := s.FS.MkdirAll(outDir); err != nil {
		return err
	}

	outPath := filepath.Join(outDir, name)
	outFile, err := s.FS.CreateFile(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, r)
	fmt.Printf("✅ saved raw file: %s\n", outPath)
	return err
}

// saveZipAndUnzip stores the original zip AND extracts its contents
func (s *SmartFilestorer) saveZipAndUnzip(ctx context.Context, name string, r iox.ReadSeekCloser) error {
	if _, err := r.Seek(0, 0); err != nil {
		return err
	}

	// 1. Save raw .zip
	zipDir := filepath.Join(s.BaseDir, "zips")
	if err := s.FS.MkdirAll(zipDir); err != nil {
		return err
	}
	zipPath := filepath.Join(zipDir, name)

	zipFile, err := s.FS.CreateFile(zipPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(zipFile, r); err != nil {
		zipFile.Close()
		return err
	}
	zipFile.Close()
	fmt.Printf("✅ saved zip: %s\n", zipPath)

	// reopen zip file to extract
	zr, err := s.ZR.Open(zipPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer zr.Close()

	if len(zr.File) == 0 {
		return fmt.Errorf("empty zip %s", name)
	}

	// 2. Extract to /receita
	outDir := filepath.Join(s.BaseDir, "receita")
	if err := s.FS.MkdirAll(outDir); err != nil {
		return err
	}

	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}

		outPath := filepath.Join(outDir, f.Name)
		outFile, err := s.FS.CreateFile(outPath)
		if err != nil {
			rc.Close()
			return err
		}

		if _, err := io.Copy(outFile, rc); err != nil {
			rc.Close()
			outFile.Close()
			return err
		}

		rc.Close()
		outFile.Close()
		fmt.Printf("📂 extracted: %s\n", outPath)
	}

	return nil
}
