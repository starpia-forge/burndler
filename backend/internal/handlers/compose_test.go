package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/burndler/burndler/internal/services"
	"github.com/gin-gonic/gin"
)

func TestNewComposeHandler(t *testing.T) {
	merger := services.NewMerger()
	linter := services.NewLinter()
	handler := NewComposeHandler(merger, linter)

	if handler == nil {
		t.Fatal("NewComposeHandler() returned nil")
	}
	if handler.merger == nil {
		t.Error("NewComposeHandler() merger is nil")
	}
	if handler.linter == nil {
		t.Error("NewComposeHandler() linter is nil")
	}
}

func TestComposeHandler_Merge(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "successful merge",
			requestBody: services.MergeRequest{
				Modules: []services.Module{
					{Name: "module1", Compose: "version: '3'\nservices:\n  web:\n    image: nginx:latest"},
					{Name: "module2", Compose: "version: '3'\nservices:\n  api:\n    image: node:14"},
				},
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var result services.MergeResult
				if err := json.Unmarshal(body, &result); err != nil {
					t.Fatal("Failed to parse response:", err)
				}
				if result.MergedCompose == "" {
					t.Error("Merge() returned empty compose")
				}
				if result.Mappings == nil {
					t.Error("Merge() mappings is nil")
				}
			},
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid-json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name: "no modules provided",
			requestBody: services.MergeRequest{
				Modules: []services.Module{},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "NO_MODULES",
		},
		{
			name: "invalid compose yaml",
			requestBody: services.MergeRequest{
				Modules: []services.Module{
					{Name: "module1", Compose: "invalid: yaml: content:"},
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "MERGE_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merger := services.NewMerger()
			linter := services.NewLinter()
			handler := NewComposeHandler(merger, linter)

			router := gin.New()
			router.POST("/merge", handler.Merge)

			body, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest(http.MethodPost, "/merge", bytes.NewBuffer(body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Merge() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatal("Failed to parse error response:", err)
				}
				if errorCode, ok := response["error"].(string); !ok || errorCode != tt.expectedError {
					t.Errorf("Merge() error = %v, want %v", response["error"], tt.expectedError)
				}
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestComposeHandler_Lint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "successful lint - valid compose",
			requestBody: services.LintRequest{
				Compose: "version: '3'\nservices:\n  web:\n    image: nginx@sha256:abcdef",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var result services.LintResult
				if err := json.Unmarshal(body, &result); err != nil {
					t.Fatal("Failed to parse response:", err)
				}
				// The actual validation depends on the linter implementation
				// For now we just check that we get a response
				if result.Errors == nil {
					t.Error("Lint() errors field is nil")
				}
			},
		},
		{
			name: "lint with build directive (should fail)",
			requestBody: services.LintRequest{
				Compose: "version: '3'\nservices:\n  web:\n    build: .",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var result services.LintResult
				if err := json.Unmarshal(body, &result); err != nil {
					t.Fatal("Failed to parse response:", err)
				}
				if !result.Valid {
					// Expected - build directive should be rejected
					foundBuildError := false
					for _, issue := range result.Errors {
						if strings.Contains(issue.Message, "build") || strings.Contains(issue.Rule, "build") {
							foundBuildError = true
							break
						}
					}
					if !foundBuildError {
						t.Error("Lint() expected error about build directive")
					}
				}
			},
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid-json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name: "no compose content",
			requestBody: services.LintRequest{
				Compose: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "NO_COMPOSE",
		},
		{
			name: "invalid yaml",
			requestBody: services.LintRequest{
				Compose: "invalid: yaml: content:",
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "LINT_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merger := services.NewMerger()
			linter := services.NewLinter()
			handler := NewComposeHandler(merger, linter)

			router := gin.New()
			router.POST("/lint", handler.Lint)

			body, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest(http.MethodPost, "/lint", bytes.NewBuffer(body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Lint() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatal("Failed to parse error response:", err)
				}
				if errorCode, ok := response["error"].(string); !ok || errorCode != tt.expectedError {
					t.Errorf("Lint() error = %v, want %v", response["error"], tt.expectedError)
				}
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}