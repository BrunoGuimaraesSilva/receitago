package local

import "io"

type FileWriter interface {
	MkdirAll(path string) error
	CreateFile(path string) (io.WriteCloser, error)
}
