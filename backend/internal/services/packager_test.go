package services

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/burndler/burndler/internal/storage"
)

// MockStorage implements storage.Storage for testing
type MockStorage struct {
	UploadCalled   bool
	DownloadCalled bool
	DeleteCalled   bool
	UploadError    error
	DownloadError  error
	DeleteError    error
}

func (m *MockStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64) (string, error) {
	m.UploadCalled = true
	if m.UploadError != nil {
		return "", m.UploadError
	}
	return "http://mock-storage/" + key, nil
}

func (m *MockStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	m.DownloadCalled = true
	if m.DownloadError != nil {
		return nil, m.DownloadError
	}
	return io.NopCloser(bytes.NewReader([]byte("mock content"))), nil
}

func (m *MockStorage) Delete(ctx context.Context, key string) error {
	m.DeleteCalled = true
	return m.DeleteError
}

func (m *MockStorage) Exists(ctx context.Context, key string) (bool, error) {
	return true, nil
}

func (m *MockStorage) List(ctx context.Context, prefix string) ([]storage.FileInfo, error) {
	return []storage.FileInfo{}, nil
}

func (m *MockStorage) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return "http://mock-storage/" + key, nil
}

// Test NewPackager constructor
func TestNewPackager(t *testing.T) {
	mockStorage := &MockStorage{}
	packager := NewPackager(mockStorage)

	if packager == nil {
		t.Error("Expected NewPackager to return non-nil packager")
	}

	// We can't directly compare storage interface, but we can verify it's set
	if packager.storage == nil {
		t.Error("Expected packager storage to be set")
	}
}

// Test CreatePackage basic functionality
func TestPackager_CreatePackage(t *testing.T) {
	mockStorage := &MockStorage{}
	packager := NewPackager(mockStorage)

	ctx := context.Background()
	req := &PackageRequest{
		Name: "test-package",
		Compose: `version: '3'
services:
  web:
    image: nginx:latest`,
		Resources: []Resource{},
	}

	packagePath, err := packager.CreatePackage(ctx, req)
	if err != nil {
		t.Fatalf("CreatePackage failed: %v", err)
	}

	if packagePath == "" {
		t.Error("Expected non-empty package path")
	}

	if !mockStorage.UploadCalled {
		t.Error("Expected storage Upload to be called")
	}
}

// Test CreatePackage with storage error
func TestPackager_CreatePackage_StorageError(t *testing.T) {
	mockStorage := &MockStorage{
		UploadError: errors.New("storage error"),
	}
	packager := NewPackager(mockStorage)

	ctx := context.Background()
	req := &PackageRequest{
		Name: "test-package",
		Compose: `version: '3'
services:
  web:
    image: nginx:latest`,
		Resources: []Resource{},
	}

	packagePath, err := packager.CreatePackage(ctx, req)
	if err == nil {
		t.Error("Expected error when storage fails")
	}

	// Even on error, we might get a partial path
	_ = packagePath
}