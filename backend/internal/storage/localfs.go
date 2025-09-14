package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/burndler/burndler/internal/config"
)

// LocalFSStorage implements Storage interface using local filesystem
// Used for development and offline deployments
type LocalFSStorage struct {
	basePath string
	maxSize  string
}

// NewLocalFSStorage creates a new local filesystem storage instance
func NewLocalFSStorage(cfg *config.Config) (*LocalFSStorage, error) {
	// Ensure base path exists
	if err := os.MkdirAll(cfg.LocalStoragePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalFSStorage{
		basePath: cfg.LocalStoragePath,
		maxSize:  cfg.LocalStorageMaxSize,
	}, nil
}

func (l *LocalFSStorage) getFullPath(key string) string {
	// Sanitize key to prevent directory traversal
	key = strings.ReplaceAll(key, "..", "")
	key = filepath.Clean(key)
	return filepath.Join(l.basePath, key)
}

func (l *LocalFSStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64) (string, error) {
	fullPath := l.getFullPath(key)

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content
	written, err := io.Copy(file, reader)
	if err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	if written != size && size > 0 {
		return "", fmt.Errorf("size mismatch: expected %d, wrote %d", size, written)
	}

	return fullPath, nil
}

func (l *LocalFSStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath := l.getFullPath(key)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", key)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

func (l *LocalFSStorage) Delete(ctx context.Context, key string) error {
	fullPath := l.getFullPath(key)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Try to remove empty parent directories
	dir := filepath.Dir(fullPath)
	for dir != l.basePath && dir != "/" && dir != "." {
		if err := os.Remove(dir); err != nil {
			break // Directory not empty or other error
		}
		dir = filepath.Dir(dir)
	}

	return nil
}

func (l *LocalFSStorage) Exists(ctx context.Context, key string) (bool, error) {
	fullPath := l.getFullPath(key)

	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}

func (l *LocalFSStorage) List(ctx context.Context, prefix string) ([]FileInfo, error) {
	searchPath := l.getFullPath(prefix)
	var files []FileInfo

	// If prefix is a directory, append wildcard
	if stat, err := os.Stat(searchPath); err == nil && stat.IsDir() {
		searchPath = filepath.Join(searchPath, "*")
	} else {
		searchPath = searchPath + "*"
	}

	matches, err := filepath.Glob(searchPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	for _, match := range matches {
		stat, err := os.Stat(match)
		if err != nil {
			continue
		}

		if !stat.IsDir() {
			// Get relative path from base
			relPath, err := filepath.Rel(l.basePath, match)
			if err != nil {
				continue
			}

			files = append(files, FileInfo{
				Key:          relPath,
				Size:         stat.Size(),
				LastModified: stat.ModTime(),
			})
		}
	}

	// Also walk subdirectories if prefix is a directory
	prefixDir := l.getFullPath(prefix)
	if stat, err := os.Stat(prefixDir); err == nil && stat.IsDir() {
		err = filepath.Walk(prefixDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors
			}

			if !info.IsDir() {
				relPath, err := filepath.Rel(l.basePath, path)
				if err != nil {
					return nil
				}

				// Check if not already added
				found := false
				for _, f := range files {
					if f.Key == relPath {
						found = true
						break
					}
				}

				if !found {
					files = append(files, FileInfo{
						Key:          relPath,
						Size:         info.Size(),
						LastModified: info.ModTime(),
					})
				}
			}

			return nil
		})
	}

	return files, nil
}

func (l *LocalFSStorage) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	// For local storage, return a file:// URL
	// In production, this would be served through the API
	fullPath := l.getFullPath(key)

	// Check if file exists
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file not found: %s", key)
		}
		return "", fmt.Errorf("failed to check file: %w", err)
	}

	// Return file URL (for local development)
	// In production, this would return an API endpoint URL
	return fmt.Sprintf("file://%s", fullPath), nil
}