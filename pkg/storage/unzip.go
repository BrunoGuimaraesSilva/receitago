package storage

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/BrunoGuimaraesSilva/receitago/pkg/iox"
	"github.com/BrunoGuimaraesSilva/receitago/pkg/storage/local"
)

const (
	receitaDir = "receita"
)

type UnzippingLocal struct {
	BaseDir string
	FS      local.FileWriter
	ZR      ZipReaderFactory
}

func NewUnzippingLocal(baseDir string, fs local.FileWriter, zr ZipReaderFactory) *UnzippingLocal {
	return &UnzippingLocal{
		BaseDir: baseDir,
		FS:      fs,
		ZR:      zr,
	}
}

func (s *UnzippingLocal) Save(ctx context.Context, name string, r iox.ReadSeekCloser) error {
	if _, err := r.Seek(0, 0); err != nil {
		return fmt.Errorf("seek: %w", err)
	}

	tmpPath := filepath.Join(s.BaseDir, fmt.Sprintf("tmp-%s", name))
	tmpFile, err := s.FS.CreateFile(tmpPath)
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	if _, err := io.Copy(tmpFile, r); err != nil {
		tmpFile.Close()
		return fmt.Errorf("copy to temp: %w", err)
	}
	tmpFile.Close()

	zr, err := s.ZR.Open(tmpPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer zr.Close()

	if len(zr.File) == 0 {
		return fmt.Errorf("empty zip: %s", name)
	}

	outDir := filepath.Join(s.BaseDir, receitaDir)
	if err := s.FS.MkdirAll(outDir); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	for _, f := range zr.File {
		if err := s.extractFile(f, outDir); err != nil {
			return err
		}
	}

	return nil
}

func (s *UnzippingLocal) extractFile(f *zip.File, outDir string) error {
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("open file %s in zip: %w", f.Name, err)
	}
	defer rc.Close()

	outPath := filepath.Join(outDir, f.Name)
	outFile, err := s.FS.CreateFile(outPath)
	if err != nil {
		return fmt.Errorf("create output %s: %w", outPath, err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, rc); err != nil {
		return fmt.Errorf("copy to output %s: %w", outPath, err)
	}
	return nil
}
