package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/burndler/burndler/internal/app"
	"github.com/burndler/burndler/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Create a test app with minimal configuration
	testApp := &app.App{
		Config: &config.Config{
			ServerHost:         "localhost",
			ServerPort:         "8080",
			ServerReadTimeout:  30 * time.Second,
			ServerWriteTimeout: 30 * time.Second,
			CORSAllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	srv := New(testApp)

	assert.NotNil(t, srv)
	assert.Equal(t, testApp, srv.app)
	assert.NotNil(t, srv.router)
	// httpServer is created on Start, not in New
	assert.Nil(t, srv.httpServer)
}

func TestServer_setupRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testApp := &app.App{
		Config: &config.Config{
			CORSAllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	srv := &Server{
		app:    testApp,
		router: gin.New(),
	}

	srv.setupRouter()

	// Test that health endpoint exists
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	srv.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "\"status\":\"healthy\"")
}

func TestServer_Start(t *testing.T) {
	testApp := &app.App{
		Config: &config.Config{
			ServerHost:         "localhost",
			ServerPort:         "0", // Use port 0 to let OS assign available port
			ServerReadTimeout:  30 * time.Second,
			ServerWriteTimeout: 30 * time.Second,
			CORSAllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	srv := New(testApp)

	// Start server in goroutine
	go func() {
		err := srv.Start()
		// Should return error when we shutdown
		assert.Error(t, err)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := srv.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestServer_Shutdown(t *testing.T) {
	testApp := &app.App{
		Config: &config.Config{
			ServerHost:         "localhost",
			ServerPort:         "0",
			ServerReadTimeout:  30 * time.Second,
			ServerWriteTimeout: 30 * time.Second,
			CORSAllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	srv := New(testApp)

	// Start server
	go srv.Start()
	time.Sleep(100 * time.Millisecond)

	// Test shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := srv.Shutdown(ctx)
	require.NoError(t, err)
}

func TestServer_healthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testApp := &app.App{
		Config: &config.Config{
			CORSAllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	srv := &Server{
		app:    testApp,
		router: gin.New(),
	}

	srv.setupRouter()

	// Test health check endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	srv.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check response body
	expectedBody := `{"status":"healthy","timestamp":`
	assert.Contains(t, w.Body.String(), expectedBody)
}