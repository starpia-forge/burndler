package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/burndler/burndler/internal/config"
	"github.com/burndler/burndler/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestNew(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		ServerHost:         "localhost",
		ServerPort:         "8080",
		ServerReadTimeout:  30 * time.Second,
		ServerWriteTimeout: 30 * time.Second,
		CORSAllowedOrigins: []string{"http://localhost:3000"},
	}

	merger := services.NewMerger()
	linter := services.NewLinter()
	packager := services.NewPackager(nil)

	srv := New(cfg, nil, merger, linter, packager)

	assert.NotNil(t, srv)
	assert.Equal(t, cfg, srv.config)
	assert.Equal(t, merger, srv.merger)
	assert.Equal(t, linter, srv.linter)
	assert.Equal(t, packager, srv.packager)
	assert.NotNil(t, srv.router)
}

func TestServer_setupRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		CORSAllowedOrigins: []string{"http://localhost:3000"},
	}

	srv := &Server{
		config:   cfg,
		merger:   services.NewMerger(),
		linter:   services.NewLinter(),
		packager: services.NewPackager(nil),
		db:       &gorm.DB{},
	}

	srv.setupRouter()

	// Test that health endpoint exists
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	srv.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"status\":\"healthy\"")
}

func TestServer_Run(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		ServerHost:         "localhost",
		ServerPort:         "0", // Use port 0 to let OS assign available port
		ServerReadTimeout:  30 * time.Second,
		ServerWriteTimeout: 30 * time.Second,
		CORSAllowedOrigins: []string{"http://localhost:3000"},
	}

	srv := New(cfg, nil, services.NewMerger(), services.NewLinter(), services.NewPackager(nil))

	// Start server in goroutine
	done := make(chan bool)
	go func() {
		err := srv.Run()
		// Server exits normally when interrupted
		assert.NoError(t, err)
		done <- true
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Send interrupt signal to stop the server
	process, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	err = process.Signal(syscall.SIGINT)
	require.NoError(t, err)

	// Wait for server to finish
	select {
	case <-done:
		// Server shut down successfully
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Server did not shut down in time")
	}
}