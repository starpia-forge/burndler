package storage

import (
	"context"
	"io"
	"mime/multipart"
	"time"
)

// Storage defines the interface for artifact storage
// Implementations: S3 (production) and LocalFS (development/offline)
type Storage interface {
	// Upload stores a file and returns its URL/path
	Upload(ctx context.Context, key string, reader io.Reader, size int64) (string, error)

	// UploadMultipart handles multipart file upload with automatic content type detection
	UploadMultipart(ctx context.Context, key string, file *multipart.FileHeader) (UploadResult, error)

	// DownloadBatch retrieves multiple files efficiently
	DownloadBatch(ctx context.Context, keys []string) (map[string][]byte, error)

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

// UploadResult contains the result of an upload operation
type UploadResult struct {
	Key         string // Storage key
	URL         string // Accessible URL
	Size        int64  // File size in bytes
	ContentType string // MIME type
	Checksum    string // SHA256 checksum
}
