package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestNewHealthHandler(t *testing.T) {
	handler := NewHealthHandler()
	if handler == nil {
		t.Fatal("NewHealthHandler() returned nil")
	}
}

func TestHealthHandler_Health(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	handler := NewHealthHandler()

	// Create a test router
	router := gin.New()
	router.GET("/health", handler.Health)

	// Create a test request
	req, err := http.NewRequest(http.MethodGet, "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Health() status = %v, want %v", w.Code, http.StatusOK)
	}

	// Parse response body
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal("Failed to parse response body:", err)
	}

	// Check status field
	if status, ok := response["status"].(string); !ok || status != "healthy" {
		t.Errorf("Health() status = %v, want 'healthy'", response["status"])
	}

	// Check version field
	if version, ok := response["version"].(string); !ok || version != "0.1.0" {
		t.Errorf("Health() version = %v, want '0.1.0'", response["version"])
	}

	// Check timestamp field exists and is valid
	if timestamp, ok := response["timestamp"].(string); !ok {
		t.Error("Health() response missing timestamp field")
	} else {
		// Try to parse the timestamp
		if _, err := time.Parse(time.RFC3339, timestamp); err != nil {
			t.Errorf("Health() timestamp is not valid RFC3339: %v", err)
		}
	}

	// Check Content-Type header
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("Health() Content-Type = %v, want 'application/json; charset=utf-8'", contentType)
	}
}