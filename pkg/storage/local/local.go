package local

import (
	"context"
	"io"
	"path/filepath"

	"github.com/BrunoGuimaraesSilva/receitago/pkg/iox"
)

type Local struct {
	BaseDir string
	FS      FileWriter
}

func NewLocal(baseDir string, fs FileWriter) *Local {
	return &Local{
		BaseDir: baseDir,
		FS:      fs,
	}
}

func (s *Local) Save(ctx context.Context, name string, r iox.ReadSeekCloser) error {
	path := filepath.Join(s.BaseDir, name)

	if err := s.FS.MkdirAll(filepath.Dir(path)); err != nil {
		return err
	}

	out, err := s.FS.CreateFile(path)
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
