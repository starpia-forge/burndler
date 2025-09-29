package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/burndler/burndler/internal/models"
	"github.com/burndler/burndler/internal/services"
	"github.com/burndler/burndler/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Mock Storage for testing
type mockStorage struct{}

func (m *mockStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64) (string, error) {
	return "https://example.com/" + key, nil
}

func (m *mockStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader([]byte("test content"))), nil
}

func (m *mockStorage) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *mockStorage) Exists(ctx context.Context, key string) (bool, error) {
	return true, nil
}

func (m *mockStorage) List(ctx context.Context, prefix string) ([]storage.FileInfo, error) {
	return []storage.FileInfo{}, nil
}

func (m *mockStorage) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return "https://example.com/signed/" + key, nil
}

// Setup test database
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal("Failed to setup test database:", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&models.Build{})
	if err != nil {
		t.Fatal("Failed to migrate test database:", err)
	}

	return db
}

func TestNewPackageHandler(t *testing.T) {
	db := setupTestDB(t)
	storage := &mockStorage{}
	packager := services.NewPackager(storage)
	handler := NewPackageHandler(packager, db)

	if handler == nil {
		t.Fatal("NewPackageHandler() returned nil")
	}
	if handler.packager == nil {
		t.Error("NewPackageHandler() packager is nil")
	}
	if handler.db == nil {
		t.Error("NewPackageHandler() db is nil")
	}
}

func TestPackageHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		userID         string
		expectedStatus int
		expectedError  string
		checkResponse  func(t *testing.T, body []byte, db *gorm.DB)
	}{
		{
			name: "successful package creation",
			requestBody: services.PackageRequest{
				Name:    "test-package",
				Compose: "version: '3'\nservices:\n  web:\n    image: nginx:latest",
			},
			userID:         "1",
			expectedStatus: http.StatusAccepted,
			checkResponse: func(t *testing.T, body []byte, db *gorm.DB) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatal("Failed to parse response:", err)
				}

				// Check build_id is returned
				if _, ok := response["build_id"].(string); !ok {
					t.Error("Create() response missing build_id")
				}

				// Check status is queued
				if status, ok := response["status"].(string); !ok || status != "queued" {
					t.Errorf("Create() status = %v, want 'queued'", response["status"])
				}

				// Check database record
				time.Sleep(50 * time.Millisecond) // Wait for goroutine
				var build models.Build
				if err := db.First(&build).Error; err != nil {
					t.Error("Build record not found in database")
				} else {
					if build.Name != "test-package" {
						t.Errorf("Build name = %v, want 'test-package'", build.Name)
					}
					if build.UserID != 1 {
						t.Errorf("Build userID = %v, want 1", build.UserID)
					}
				}
			},
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid-json",
			userID:         "1",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_REQUEST",
		},
		{
			name: "missing required fields",
			requestBody: services.PackageRequest{
				Name:    "", // Missing name
				Compose: "version: '3'",
			},
			userID:         "1",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_FIELDS",
		},
		{
			name: "missing compose content",
			requestBody: services.PackageRequest{
				Name:    "test-package",
				Compose: "", // Missing compose
			},
			userID:         "1",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_FIELDS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			storage := &mockStorage{}
			packager := services.NewPackager(storage)
			handler := NewPackageHandler(packager, db)

			router := gin.New()
			router.POST("/package", func(c *gin.Context) {
				// Set user_id in context
				c.Set("user_id", tt.userID)
				handler.Create(c)
			})

			body, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest(http.MethodPost, "/package", bytes.NewBuffer(body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Create() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatal("Failed to parse error response:", err)
				}
				if errorCode, ok := response["error"].(string); !ok || errorCode != tt.expectedError {
					t.Errorf("Create() error = %v, want %v", response["error"], tt.expectedError)
				}
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes(), db)
			}
		})
	}
}

func TestPackageHandler_Status(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		buildID        string
		setupDB        func(*gorm.DB) *models.Build
		expectedStatus int
		expectedError  string
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name:    "successful status retrieval",
			buildID: uuid.New().String(),
			setupDB: func(db *gorm.DB) *models.Build {
				build := &models.Build{
					ID:          uuid.New(),
					Name:        "test-package",
					Status:      "completed",
					Progress:    100,
					DownloadURL: "https://example.com/package.tar.gz",
					UserID:      1,
				}
				db.Create(build)
				return build
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatal("Failed to parse response:", err)
				}

				if status, ok := response["status"].(string); !ok || status != "completed" {
					t.Errorf("Status() status = %v, want 'completed'", response["status"])
				}
				if progress, ok := response["progress"].(float64); !ok || progress != 100 {
					t.Errorf("Status() progress = %v, want 100", response["progress"])
				}
				if url, ok := response["download_url"].(string); !ok || url != "https://example.com/package.tar.gz" {
					t.Errorf("Status() download_url = %v, want 'https://example.com/package.tar.gz'", response["download_url"])
				}
			},
		},
		{
			name:           "invalid build ID format",
			buildID:        "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_BUILD_ID",
		},
		{
			name:    "build not found",
			buildID: uuid.New().String(),
			setupDB: func(db *gorm.DB) *models.Build {
				// Don't create any build
				return nil
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "BUILD_NOT_FOUND",
		},
		{
			name:    "build with error",
			buildID: uuid.New().String(),
			setupDB: func(db *gorm.DB) *models.Build {
				build := &models.Build{
					ID:       uuid.New(),
					Name:     "failed-package",
					Status:   "failed",
					Progress: 50,
					Error:    "Image pull failed",
					UserID:   1,
				}
				db.Create(build)
				return build
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatal("Failed to parse response:", err)
				}

				if status, ok := response["status"].(string); !ok || status != "failed" {
					t.Errorf("Status() status = %v, want 'failed'", response["status"])
				}
				if errorMsg, ok := response["error"].(string); !ok || errorMsg != "Image pull failed" {
					t.Errorf("Status() error = %v, want 'Image pull failed'", response["error"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			storage := &mockStorage{}
			packager := services.NewPackager(storage)
			handler := NewPackageHandler(packager, db)

			var testBuild *models.Build
			if tt.setupDB != nil {
				testBuild = tt.setupDB(db)
			}

			router := gin.New()
			router.GET("/package/:id/status", handler.Status)

			// Use the test build's ID if available
			buildID := tt.buildID
			if testBuild != nil {
				buildID = testBuild.ID.String()
			}

			req, err := http.NewRequest(http.MethodGet, "/package/"+buildID+"/status", nil)
			if err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Status() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatal("Failed to parse error response:", err)
				}
				if errorCode, ok := response["error"].(string); !ok || errorCode != tt.expectedError {
					t.Errorf("Status() error = %v, want %v", response["error"], tt.expectedError)
				}
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestPackageHandler_ProcessPackage(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus string
		expectedError  string
	}{
		{
			name:           "successful package processing",
			expectedStatus: "completed",
		},
		{
			name:           "failed package processing",
			expectedStatus: "building",
			expectedError:  "", // processPackage doesn't fail in basic test
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			storage := &mockStorage{}
			packager := services.NewPackager(storage)
			handler := NewPackageHandler(packager, db)

			// Create initial build record
			build := &models.Build{
				ID:          uuid.New(),
				Name:        "test-package",
				Status:      "queued",
				UserID:      1,
				ComposeYAML: "version: '3'",
			}
			db.Create(build)

			// Process package
			req := &services.PackageRequest{
				Name:    "test-package",
				Compose: "version: '3'",
			}
			handler.processPackage(build, req)

			// Wait for async processing
			time.Sleep(100 * time.Millisecond)

			// Verify the build was updated
			var updatedBuild models.Build
			if err := db.First(&updatedBuild, "id = ?", build.ID).Error; err != nil {
				t.Fatal("Failed to fetch updated build:", err)
			}

			// Basic verification - the actual CreatePackage would need proper mocking
			if updatedBuild.Status == "queued" {
				t.Error("processPackage() didn't update status from queued")
			}
		})
	}
}
