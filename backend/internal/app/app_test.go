package app

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/burndler/burndler/internal/config"
	"github.com/burndler/burndler/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Test that New creates an App instance with all dependencies
	app, err := New()

	// For testing, we expect an error since we don't have a real database
	// In a real test environment, we'd mock the database connection
	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

func TestNewWithConfig(t *testing.T) {
	// Test with a test configuration
	cfg := &config.Config{
		DBHost:               "localhost",
		DBPort:               "5432",
		DBUser:               "test",
		DBPassword:           "test",
		DBName:               "test_db",
		DBSSLMode:            "disable",
		DBMaxConnections:     10,
		DBMaxIdleConnections: 5,
		StorageMode:          "local",
		LocalStoragePath:     t.TempDir(),
		LocalStorageMaxSize:  "100MB",
	}

	app, err := NewWithConfig(cfg)

	// We expect an error since we don't have a real database connection
	assert.Error(t, err)
	assert.Nil(t, app)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

func TestApp_Close(t *testing.T) {
	// Test that Close handles nil DB gracefully
	app := &App{
		DB: nil,
	}

	err := app.Close()
	assert.NoError(t, err)
}

func TestInitStorage_LocalFS(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		StorageMode:         "local",
		LocalStoragePath:    tempDir,
		LocalStorageMaxSize: "100MB",
	}

	storage, err := initStorage(cfg)
	require.NoError(t, err)
	assert.NotNil(t, storage)
}

func TestInitStorage_S3(t *testing.T) {
	cfg := &config.Config{
		StorageMode:       "s3",
		S3Bucket:          "test-bucket",
		S3Region:          "us-east-1",
		S3Endpoint:        "",
		S3AccessKeyID:     "test-access-key",
		S3SecretAccessKey: "test-secret-key",
	}

	storage, err := initStorage(cfg)
	// S3 storage should initialize successfully with test credentials
	require.NoError(t, err)
	assert.NotNil(t, storage)
}

func TestInitStorage_UnknownMode(t *testing.T) {
	cfg := &config.Config{
		StorageMode: "unknown",
	}

	storage, err := initStorage(cfg)
	assert.Error(t, err)
	assert.Nil(t, storage)
	assert.Contains(t, err.Error(), "unknown storage mode")
}

func TestApp_Run(t *testing.T) {
	// Test that Run method delegates to server correctly
	// This is a simple test since the actual server logic is tested in server_test.go

	testApp := &App{
		Config: &config.Config{
			ServerHost:         "localhost",
			ServerPort:         "0", // Use port 0 to let OS assign available port
			ServerReadTimeout:  30 * time.Second,
			ServerWriteTimeout: 30 * time.Second,
			CORSAllowedOrigins: []string{"http://localhost:3000"},
		},
		Merger:   &services.Merger{},
		Linter:   &services.Linter{},
		Packager: &services.Packager{},
	}

	// Start the app in a goroutine
	go func() {
		// We expect this to block until interrupted
		err := testApp.Run()
		// Should exit cleanly when interrupted
		assert.NoError(t, err)
	}()

	// Give the server time to start
	time.Sleep(100 * time.Millisecond)

	// Send interrupt signal to trigger shutdown
	process, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	err = process.Signal(syscall.SIGINT)
	require.NoError(t, err)

	// Give time for graceful shutdown
	time.Sleep(200 * time.Millisecond)
}
