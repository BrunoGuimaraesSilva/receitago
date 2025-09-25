package downloader

import "os"

type tempFile struct{ *os.File }

func (t *tempFile) Close() error {
	return t.File.Close()
}
