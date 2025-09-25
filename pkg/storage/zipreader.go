package storage

import "archive/zip"

type ZipReaderFactory interface {
	Open(path string) (*zip.ReadCloser, error)
}

type StdZipReader struct{}

func (StdZipReader) Open(path string) (*zip.ReadCloser, error) {
	return zip.OpenReader(path)
}
