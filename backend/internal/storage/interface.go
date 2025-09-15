package storage

import (
	"context"
	"io"
	"time"
)

// Storage defines the interface for artifact storage
// Implementations: S3 (production) and LocalFS (development/offline)
type Storage interface {
	// Upload stores a file and returns its URL/path
	Upload(ctx context.Context, key string, reader io.Reader, size int64) (string, error)

	// Download retrieves a file
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes a file
	Delete(ctx context.Context, key string) error

	// Exists checks if a file exists
	Exists(ctx context.Context, key string) (bool, error)

	// List returns all files with the given prefix
	List(ctx context.Context, prefix string) ([]FileInfo, error)

	// GetURL returns a signed/accessible URL for the file
	GetURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}

// FileInfo contains metadata about a stored file
type FileInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	ContentType  string
}
