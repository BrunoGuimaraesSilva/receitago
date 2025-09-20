package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/BrunoGuimaraesSilva/receitago/services/downloader/domain/dataset"
)

type Local struct {
	BaseDir string
}

func NewLocal(baseDir string) *Local { return &Local{BaseDir: baseDir} }

func (s *Local) Save(ctx context.Context, name string, r dataset.ReadSeekCloser) error {
	path := filepath.Join(s.BaseDir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := r.Seek(0, 0); err != nil {
		return err
	}
	_, err = io.Copy(out, r)
	return err
}
