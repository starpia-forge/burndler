package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/burndler/burndler/internal/config"
)

// Test NewLocalFSStorage constructor
func TestNewLocalFSStorage(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()

	cfg := &config.Config{
		LocalStoragePath:    tempDir,
		LocalStorageMaxSize: "100MB",
	}

	fs, err := NewLocalFSStorage(cfg)
	if err != nil {
		t.Fatalf("NewLocalFSStorage failed: %v", err)
	}

	if fs == nil {
		t.Error("Expected non-nil LocalFSStorage")
	}

	// Check that base directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("Expected base directory to exist")
	}
}

// Test Upload method
func TestLocalFS_Upload(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		LocalStoragePath:    tempDir,
		LocalStorageMaxSize: "100MB",
	}

	fs, err := NewLocalFSStorage(cfg)
	if err != nil {
		t.Fatalf("NewLocalFSStorage failed: %v", err)
	}

	ctx := context.Background()
	content := []byte("test content")
	reader := bytes.NewReader(content)

	path, err := fs.Upload(ctx, "test/file.txt", reader, int64(len(content)))
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	if path == "" {
		t.Error("Expected non-empty path")
	}

	// Verify file was created
	fullPath := filepath.Join(tempDir, "test", "file.txt")
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Failed to read uploaded file: %v", err)
	}

	if string(data) != string(content) {
		t.Errorf("File content mismatch: got %s, want %s", data, content)
	}
}

// Test Download method
func TestLocalFS_Download(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		LocalStoragePath:    tempDir,
		LocalStorageMaxSize: "100MB",
	}

	fs, err := NewLocalFSStorage(cfg)
	if err != nil {
		t.Fatalf("NewLocalFSStorage failed: %v", err)
	}

	ctx := context.Background()
	content := []byte("download test content")

	// First upload a file
	reader := bytes.NewReader(content)
	_, err = fs.Upload(ctx, "download/test.txt", reader, int64(len(content)))
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	// Now download it
	downloadReader, err := fs.Download(ctx, "download/test.txt")
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	defer downloadReader.Close()

	downloaded, err := io.ReadAll(downloadReader)
	if err != nil {
		t.Fatalf("Failed to read download: %v", err)
	}

	if string(downloaded) != string(content) {
		t.Errorf("Downloaded content mismatch: got %s, want %s", downloaded, content)
	}
}

// Test Delete method
func TestLocalFS_Delete(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		LocalStoragePath:    tempDir,
		LocalStorageMaxSize: "100MB",
	}

	fs, err := NewLocalFSStorage(cfg)
	if err != nil {
		t.Fatalf("NewLocalFSStorage failed: %v", err)
	}

	ctx := context.Background()
	content := []byte("delete test")

	// Upload a file
	reader := bytes.NewReader(content)
	_, err = fs.Upload(ctx, "delete/test.txt", reader, int64(len(content)))
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	// Verify it exists
	exists, err := fs.Exists(ctx, "delete/test.txt")
	if err != nil {
		t.Fatalf("Exists check failed: %v", err)
	}
	if !exists {
		t.Error("Expected file to exist before delete")
	}

	// Delete it
	err = fs.Delete(ctx, "delete/test.txt")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	exists, err = fs.Exists(ctx, "delete/test.txt")
	if err != nil {
		t.Fatalf("Exists check failed: %v", err)
	}
	if exists {
		t.Error("Expected file to not exist after delete")
	}
}

// Test Exists method
func TestLocalFS_Exists(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		LocalStoragePath:    tempDir,
		LocalStorageMaxSize: "100MB",
	}

	fs, err := NewLocalFSStorage(cfg)
	if err != nil {
		t.Fatalf("NewLocalFSStorage failed: %v", err)
	}

	ctx := context.Background()

	// Check non-existent file
	exists, err := fs.Exists(ctx, "nonexistent.txt")
	if err != nil {
		t.Fatalf("Exists check failed: %v", err)
	}
	if exists {
		t.Error("Expected false for non-existent file")
	}

	// Upload a file
	content := []byte("exists test")
	reader := bytes.NewReader(content)
	_, err = fs.Upload(ctx, "exists/test.txt", reader, int64(len(content)))
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	// Check it exists
	exists, err = fs.Exists(ctx, "exists/test.txt")
	if err != nil {
		t.Fatalf("Exists check failed: %v", err)
	}
	if !exists {
		t.Error("Expected true for existing file")
	}
}

// Test GetURL method
func TestLocalFS_GetURL(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		LocalStoragePath:    tempDir,
		LocalStorageMaxSize: "100MB",
	}

	fs, err := NewLocalFSStorage(cfg)
	if err != nil {
		t.Fatalf("NewLocalFSStorage failed: %v", err)
	}

	ctx := context.Background()

	// First upload a file so it exists
	content := []byte("url test")
	reader := bytes.NewReader(content)
	_, err = fs.Upload(ctx, "test/file.txt", reader, int64(len(content)))
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	// Now get its URL
	url, err := fs.GetURL(ctx, "test/file.txt", 1*time.Hour)
	if err != nil {
		t.Fatalf("GetURL failed: %v", err)
	}

	if url == "" {
		t.Error("Expected non-empty URL")
	}

	// For LocalFS, URL should be file:// scheme
	if !bytes.HasPrefix([]byte(url), []byte("file://")) {
		t.Errorf("Expected file:// URL, got %s", url)
	}
}