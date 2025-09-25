package local

import (
	"io"
	"os"
)

type LocalFS struct{}

func (LocalFS) MkdirAll(path string) error {
	return os.MkdirAll(path, 0o755)
}

func (LocalFS) CreateFile(path string) (io.WriteCloser, error) {
	return os.Create(path)
}
