package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/burndler/burndler/internal/config"
	"github.com/burndler/burndler/internal/middleware"
	"github.com/burndler/burndler/internal/models"
	"github.com/burndler/burndler/internal/services"
	"github.com/burndler/burndler/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDBForModule(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.Module{}, &models.ModuleVersion{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func setupTestUser(t *testing.T, db *gorm.DB, role string) *models.User {
	user := &models.User{
		Email: fmt.Sprintf("test-%s@example.com", role),
		Name:  fmt.Sprintf("Test %s", role),
		Role:  role,
	}
	err := user.SetPassword("testPassword123!")
	assert.NoError(t, err)

	err = db.Create(user).Error
	assert.NoError(t, err)

	return user
}

func setupModuleHandler(t *testing.T) (*ModuleHandler, *gorm.DB, *services.AuthService) {
	gin.SetMode(gin.TestMode)

	db := setupTestDBForModule(t)

	cfg := &config.Config{
		JWTSecret:            "test-secret-key",
		JWTIssuer:            "burndler",
		JWTAudience:          "burndler-api",
		JWTExpiration:        time.Hour * 24,
		JWTRefreshExpiration: time.Hour * 168,
	}

	authService := services.NewAuthService(cfg, db)

	// Create storage instance (local for testing)
	localStorage, err := storage.NewLocalFSStorage(&config.Config{
		LocalStoragePath:    "./test_storage",
		LocalStorageMaxSize: "1GB",
	})
	assert.NoError(t, err)

	// Create linter service
	linter := services.NewLinter()

	// Create module service
	moduleService := services.NewModuleService(db, localStorage, linter)

	// Create module handler
	moduleHandler := NewModuleHandler(moduleService, db)

	return moduleHandler, db, authService
}

func getAuthToken(t *testing.T, authService *services.AuthService, user *models.User) string {
	token, err := authService.GenerateToken(user)
	assert.NoError(t, err)
	return token
}

func TestModuleHandler_ListModules(t *testing.T) {
	moduleHandler, db, authService := setupModuleHandler(t)

	// Create test users
	developer := setupTestUser(t, db, "Developer")
	engineer := setupTestUser(t, db, "Engineer")

	// Create test modules
	modules := []models.Module{
		{
			Name:        "test-module-1",
			Description: "Test Module 1",
			Author:      "Author 1",
			Active:      true,
		},
		{
			Name:        "test-module-2",
			Description: "Test Module 2",
			Author:      "Author 2",
			Active:      false,
		},
	}

	for _, module := range modules {
		err := db.Create(&module).Error
		assert.NoError(t, err)
	}

	tests := []struct {
		name           string
		user           *models.User
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "developer can list modules",
			user:           developer,
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  2, // Both modules (active and inactive)
		},
		{
			name:           "engineer can list modules",
			user:           engineer,
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  2, // Both modules (active and inactive)
		},
		{
			name:           "list with show_deleted=true",
			user:           developer,
			queryParams:    "?show_deleted=true",
			expectedStatus: http.StatusOK,
			expectedCount:  2, // Both active and inactive
		},
		{
			name:           "list with pagination",
			user:           developer,
			queryParams:    "?page=1&page_size=1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "list with invalid page size",
			user:           developer,
			queryParams:    "?page_size=101", // Over limit, should clamp to 100
			expectedStatus: http.StatusOK,
			expectedCount:  2, // Should show all 2 modules
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			token := getAuthToken(t, authService, tt.user)

			router.GET("/modules", middleware.JWTAuth(&config.Config{
				JWTSecret:   "test-secret-key",
				JWTIssuer:   "burndler",
				JWTAudience: "burndler-api",
			}), moduleHandler.ListModules)

			req, _ := http.NewRequest(http.MethodGet, "/modules"+tt.queryParams, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Contains(t, response, "data")
				assert.Contains(t, response, "total")
				assert.Contains(t, response, "page")
				assert.Contains(t, response, "page_size")

				data, ok := response["data"].([]interface{})
				assert.True(t, ok)
				assert.Equal(t, tt.expectedCount, len(data))
			}
		})
	}
}

func TestModuleHandler_CreateModule(t *testing.T) {
	moduleHandler, db, authService := setupModuleHandler(t)

	developer := setupTestUser(t, db, "Developer")
	engineer := setupTestUser(t, db, "Engineer")

	tests := []struct {
		name           string
		user           *models.User
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "developer can create module",
			user: developer,
			requestBody: map[string]string{
				"name":        "new-module",
				"description": "New Module Description",
				"author":      "Test Author",
				"repository":  "https://github.com/test/repo",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "engineer cannot create module",
			user: engineer,
			requestBody: map[string]string{
				"name":        "engineer-module",
				"description": "Engineer Module",
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "missing required name",
			user: developer,
			requestBody: map[string]string{
				"description": "Module without name",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "duplicate module name",
			user: developer,
			requestBody: map[string]string{
				"name":        "new-module", // Same as first test
				"description": "Duplicate module",
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			token := getAuthToken(t, authService, tt.user)

			cfg := &config.Config{
				JWTSecret:   "test-secret-key",
				JWTIssuer:   "burndler",
				JWTAudience: "burndler-api",
			}

			router.POST("/modules",
				middleware.JWTAuth(cfg),
				middleware.RequireRole("Developer"),
				moduleHandler.CreateModule)

			reqBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/modules", bytes.NewBuffer(reqBody))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Contains(t, response, "id")
				assert.Contains(t, response, "name")
				assert.Equal(t, tt.requestBody.(map[string]string)["name"], response["name"])
			}
		})
	}
}

func TestModuleHandler_GetModule(t *testing.T) {
	moduleHandler, db, authService := setupModuleHandler(t)

	developer := setupTestUser(t, db, "Developer")

	// Create test module
	module := models.Module{
		Name:        "get-test-module",
		Description: "Get Test Module",
		Author:      "Test Author",
		Active:      true,
	}
	err := db.Create(&module).Error
	assert.NoError(t, err)

	tests := []struct {
		name           string
		moduleID       string
		expectedStatus int
	}{
		{
			name:           "get existing module",
			moduleID:       fmt.Sprintf("%d", module.ID),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get non-existent module",
			moduleID:       "99999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid module ID",
			moduleID:       "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			token := getAuthToken(t, authService, developer)

			router.GET("/modules/:id", middleware.JWTAuth(&config.Config{
				JWTSecret:   "test-secret-key",
				JWTIssuer:   "burndler",
				JWTAudience: "burndler-api",
			}), moduleHandler.GetModule)

			req, _ := http.NewRequest(http.MethodGet, "/modules/"+tt.moduleID, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Contains(t, response, "id")
				assert.Contains(t, response, "name")
				assert.Equal(t, "get-test-module", response["name"])
			}
		})
	}
}

func TestModuleHandler_UpdateModule(t *testing.T) {
	moduleHandler, db, authService := setupModuleHandler(t)

	developer := setupTestUser(t, db, "Developer")
	engineer := setupTestUser(t, db, "Engineer")

	// Create test module
	module := models.Module{
		Name:        "update-test-module",
		Description: "Original Description",
		Author:      "Original Author",
		Active:      true,
	}
	err := db.Create(&module).Error
	assert.NoError(t, err)

	tests := []struct {
		name           string
		user           *models.User
		moduleID       string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name:     "developer can update module",
			user:     developer,
			moduleID: fmt.Sprintf("%d", module.ID),
			requestBody: map[string]interface{}{
				"description": "Updated Description",
				"author":      "Updated Author",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "engineer cannot update module",
			user:     engineer,
			moduleID: fmt.Sprintf("%d", module.ID),
			requestBody: map[string]string{
				"description": "Engineer Update",
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "update non-existent module",
			user:           developer,
			moduleID:       "99999",
			requestBody:    map[string]string{"description": "Update"},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			token := getAuthToken(t, authService, tt.user)

			cfg := &config.Config{
				JWTSecret:   "test-secret-key",
				JWTIssuer:   "burndler",
				JWTAudience: "burndler-api",
			}

			router.PUT("/modules/:id",
				middleware.JWTAuth(cfg),
				middleware.RequireRole("Developer"),
				moduleHandler.UpdateModule)

			reqBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPut, "/modules/"+tt.moduleID, bytes.NewBuffer(reqBody))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Contains(t, response, "description")
				assert.Equal(t, "Updated Description", response["description"])
			}
		})
	}
}

func TestModuleHandler_DeleteModule(t *testing.T) {
	moduleHandler, db, authService := setupModuleHandler(t)

	developer := setupTestUser(t, db, "Developer")
	engineer := setupTestUser(t, db, "Engineer")

	// Create test module
	module := models.Module{
		Name:        "delete-test-module",
		Description: "Delete Test Module",
		Author:      "Test Author",
		Active:      true,
	}
	err := db.Create(&module).Error
	assert.NoError(t, err)

	tests := []struct {
		name           string
		user           *models.User
		moduleID       string
		expectedStatus int
	}{
		{
			name:           "engineer cannot delete module",
			user:           engineer,
			moduleID:       fmt.Sprintf("%d", module.ID),
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "developer can delete module",
			user:           developer,
			moduleID:       fmt.Sprintf("%d", module.ID),
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "delete non-existent module",
			user:           developer,
			moduleID:       "99999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			token := getAuthToken(t, authService, tt.user)

			cfg := &config.Config{
				JWTSecret:   "test-secret-key",
				JWTIssuer:   "burndler",
				JWTAudience: "burndler-api",
			}

			router.DELETE("/modules/:id",
				middleware.JWTAuth(cfg),
				middleware.RequireRole("Developer"),
				moduleHandler.DeleteModule)

			req, _ := http.NewRequest(http.MethodDelete, "/modules/"+tt.moduleID, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusNoContent {
				// Verify module is soft deleted
				var deletedModule models.Module
				err := db.Unscoped().First(&deletedModule, module.ID).Error
				assert.NoError(t, err)
				assert.NotNil(t, deletedModule.DeletedAt)
			}
		})
	}
}

func TestModuleHandler_CreateVersion(t *testing.T) {
	moduleHandler, db, authService := setupModuleHandler(t)

	developer := setupTestUser(t, db, "Developer")

	// Create test module
	module := models.Module{
		Name:        "version-test-module",
		Description: "Version Test Module",
		Author:      "Test Author",
		Active:      true,
	}
	err := db.Create(&module).Error
	assert.NoError(t, err)

	tests := []struct {
		name           string
		moduleID       string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name:     "create valid version",
			moduleID: fmt.Sprintf("%d", module.ID),
			requestBody: map[string]interface{}{
				"version": "v1.0.0",
				"compose": "version: '3.8'\nservices:\n  app:\n    image: nginx:latest",
				"variables": map[string]interface{}{
					"port": 8080,
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:     "invalid semantic version",
			moduleID: fmt.Sprintf("%d", module.ID),
			requestBody: map[string]string{
				"version": "invalid-version",
				"compose": "version: '3.8'\nservices:\n  app:\n    image: nginx:latest",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "duplicate version",
			moduleID: fmt.Sprintf("%d", module.ID),
			requestBody: map[string]string{
				"version": "v1.0.0", // Same as first test
				"compose": "version: '3.8'\nservices:\n  app:\n    image: nginx:latest",
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:     "missing compose content",
			moduleID: fmt.Sprintf("%d", module.ID),
			requestBody: map[string]string{
				"version": "v1.1.0",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			token := getAuthToken(t, authService, developer)

			cfg := &config.Config{
				JWTSecret:   "test-secret-key",
				JWTIssuer:   "burndler",
				JWTAudience: "burndler-api",
			}

			router.POST("/modules/:id/versions",
				middleware.JWTAuth(cfg),
				middleware.RequireRole("Developer"),
				moduleHandler.CreateVersion)

			reqBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/modules/"+tt.moduleID+"/versions", bytes.NewBuffer(reqBody))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Contains(t, response, "id")
				assert.Contains(t, response, "version")
				assert.Contains(t, response, "published")
				assert.Equal(t, false, response["published"]) // Should be unpublished by default
			}
		})
	}
}

func TestModuleHandler_Unauthenticated(t *testing.T) {
	moduleHandler, _, _ := setupModuleHandler(t)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "list modules without auth",
			method:         http.MethodGet,
			path:           "/modules",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "create module without auth",
			method:         http.MethodPost,
			path:           "/modules",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "get module without auth",
			method:         http.MethodGet,
			path:           "/modules/1",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			cfg := &config.Config{
				JWTSecret:   "test-secret-key",
				JWTIssuer:   "burndler",
				JWTAudience: "burndler-api",
			}

			// Setup routes with auth middleware
			router.GET("/modules", middleware.JWTAuth(cfg), moduleHandler.ListModules)
			router.POST("/modules", middleware.JWTAuth(cfg), moduleHandler.CreateModule)
			router.GET("/modules/:id", middleware.JWTAuth(cfg), moduleHandler.GetModule)

			var reqBody []byte
			if tt.method == http.MethodPost {
				body := map[string]string{"name": "test"}
				reqBody, _ = json.Marshal(body)
			}

			req, _ := http.NewRequest(tt.method, tt.path, bytes.NewBuffer(reqBody))
			if tt.method == http.MethodPost {
				req.Header.Set("Content-Type", "application/json")
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response, "error")
		})
	}
}