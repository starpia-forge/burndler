package services

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"testing"
	"time"

	"github.com/burndler/burndler/internal/models"
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

func (m *MockStorage) UploadMultipart(ctx context.Context, key string, file *multipart.FileHeader) (storage.UploadResult, error) {
	m.UploadCalled = true
	if m.UploadError != nil {
		return storage.UploadResult{}, m.UploadError
	}
	return storage.UploadResult{
		Key:         key,
		URL:         "http://mock-storage/" + key,
		Size:        file.Size,
		ContentType: "application/octet-stream",
		Checksum:    "mock-checksum",
	}, nil
}

func (m *MockStorage) DownloadBatch(ctx context.Context, keys []string) (map[string][]byte, error) {
	m.DownloadCalled = true
	if m.DownloadError != nil {
		return nil, m.DownloadError
	}
	results := make(map[string][]byte)
	for _, key := range keys {
		results[key] = []byte("mock content")
	}
	return results, nil
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
		t.Fatal("Expected NewPackager to return non-nil packager")
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
		Resources:          []ResourceFile{},
		ContainerResources: []ContainerResourceGroup{},
		DownloadAssets:     []DownloadAssetInfo{},
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
		Resources:          []ResourceFile{},
		ContainerResources: []ContainerResourceGroup{},
		DownloadAssets:     []DownloadAssetInfo{},
	}

	packagePath, err := packager.CreatePackage(ctx, req)
	if err == nil {
		t.Error("Expected error when storage fails")
	}

	// Even on error, we might get a partial path
	_ = packagePath
}

// Test CreatePackage with resources and download assets
func TestPackager_CreatePackage_WithResources(t *testing.T) {
	mockStorage := &MockStorage{}
	packager := NewPackager(mockStorage)

	ctx := context.Background()
	req := &PackageRequest{
		Name: "test-package",
		Compose: `version: '3'
services:
  web:
    image: nginx:latest`,
		Resources: []ResourceFile{
			{
				Path:    "resources/service_1/container1/config.yaml",
				Content: []byte("test: config"),
			},
			{
				Path:    "resources/service_1/container2/data.json",
				Content: []byte("{\"key\": \"value\"}"),
			},
		},
		DownloadAssets: []DownloadAssetInfo{
			{
				FilePath:    "resources/service_1/container1/large-file.bin",
				DownloadURL: "https://example.com/download/large-file.bin",
				Checksum:    "abc123def456",
				FileSize:    1024000,
			},
		},
		ContainerResources: []ContainerResourceGroup{},
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

// Test CreatePackage with ContainerResources
func TestPackager_CreatePackage_WithContainerResources(t *testing.T) {
	mockStorage := &MockStorage{}
	packager := NewPackager(mockStorage)

	ctx := context.Background()

	// Mock ContainerResources with two containers
	containerResources := []ContainerResourceGroup{
		{
			ContainerID:   1,
			ContainerName: "nginx-proxy",
			Resources: []models.ContainerResource{
				{
					ID:                 1,
					ContainerVersionID: 1,
					Path:               "config/app.yaml",
					StorageKey:         "containers/1/versions/1/resources/config/app.yaml",
				},
				{
					ID:                 2,
					ContainerVersionID: 1,
					Path:               "scripts/init.sh",
					StorageKey:         "containers/1/versions/1/resources/scripts/init.sh",
				},
			},
		},
		{
			ContainerID:   2,
			ContainerName: "postgres-db",
			Resources: []models.ContainerResource{
				{
					ID:                 3,
					ContainerVersionID: 2,
					Path:               "data/config.json",
					StorageKey:         "containers/2/versions/1/resources/data/config.json",
				},
			},
		},
	}

	req := &PackageRequest{
		Name: "test-package",
		Compose: `version: '3'
services:
  web:
    image: nginx:latest`,
		Resources:          []ResourceFile{},
		ContainerResources: containerResources,
		DownloadAssets:     []DownloadAssetInfo{},
	}

	packagePath, err := packager.CreatePackage(ctx, req)
	if err != nil {
		t.Fatalf("CreatePackage failed: %v", err)
	}

	if packagePath == "" {
		t.Error("Expected non-empty package path")
	}

	if !mockStorage.DownloadCalled {
		t.Error("Expected storage Download to be called for container resources")
	}

	if !mockStorage.UploadCalled {
		t.Error("Expected storage Upload to be called")
	}
}
