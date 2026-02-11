package downloader

import "time"

const (
	DefaultChunkTimeout   = 10 * time.Minute
	DefaultMaxConcurrency = 50
	DefaultMaxRetries     = 3
	DefaultChunkSizeMB    = 50
)

const (
	DefaultHTTPTimeout = 2 * time.Minute
)

type ChunkConfig struct {
	Timeout      time.Duration
	ChunkSizeMB  int
	MaxRetries   int
	ShowProgress bool
}

func DefaultChunkConfig() ChunkConfig {
	return ChunkConfig{
		Timeout:      DefaultChunkTimeout,
		ChunkSizeMB:  DefaultChunkSizeMB,
		MaxRetries:   DefaultMaxRetries,
		ShowProgress: false,
	}
}

type HTTPConfig struct {
	Timeout   time.Duration
	UserAgent string
}

func DefaultHTTPConfig() HTTPConfig {
	return HTTPConfig{
		Timeout:   DefaultHTTPTimeout,
		UserAgent: "",
	}
}
