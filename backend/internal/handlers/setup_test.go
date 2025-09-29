package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/burndler/burndler/internal/config"
	"github.com/burndler/burndler/internal/models"
	"github.com/burndler/burndler/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDBForSetupHandler(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.Setup{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestNewSetupHandler(t *testing.T) {
	db := setupTestDBForSetupHandler(t)
	cfg := &config.Config{}
	setupService := services.NewSetupService(db, cfg)

	handler := NewSetupHandler(setupService, db)

	assert.NotNil(t, handler)
	assert.Equal(t, setupService, handler.setupService)
	assert.Equal(t, db, handler.db)
}

func TestSetupHandler_GetStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDBForSetupHandler(t)
	cfg := &config.Config{}
	setupService := services.NewSetupService(db, cfg)
	handler := NewSetupHandler(setupService, db)

	router := gin.New()
	router.GET("/setup/status", handler.GetStatus)

	req, _ := http.NewRequest(http.MethodGet, "/setup/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SetupStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.IsCompleted)
	assert.True(t, response.RequiresSetup)
	assert.False(t, response.AdminExists)
	assert.NotEmpty(t, response.SetupToken)
}

func TestSetupHandler_Initialize(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDBForSetupHandler(t)
	cfg := &config.Config{}
	setupService := services.NewSetupService(db, cfg)
	handler := NewSetupHandler(setupService, db)

	router := gin.New()
	router.POST("/setup/init", handler.Initialize)

	tests := []struct {
		name           string
		request        InitializeRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid token",
			request: InitializeRequest{
				SetupToken: "1234567890abcdef1234567890abcdef12345678",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid token - too short",
			request: InitializeRequest{
				SetupToken: "short",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing token",
			request: InitializeRequest{
				SetupToken: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest(http.MethodPost, "/setup/init", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSetupHandler_CreateAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDBForSetupHandler(t)
	cfg := &config.Config{}
	setupService := services.NewSetupService(db, cfg)
	handler := NewSetupHandler(setupService, db)

	router := gin.New()
	router.POST("/setup/admin", handler.CreateAdmin)

	tests := []struct {
		name           string
		request        CreateAdminRequest
		expectedStatus int
	}{
		{
			name: "valid admin creation",
			request: CreateAdminRequest{
				Email:    "admin@example.com",
				Password: "adminPassword123!",
				Name:     "Admin User",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid email",
			request: CreateAdminRequest{
				Email:    "invalid-email",
				Password: "adminPassword123!",
				Name:     "Admin User",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "short password",
			request: CreateAdminRequest{
				Email:    "admin@example.com",
				Password: "short",
				Name:     "Admin User",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty name",
			request: CreateAdminRequest{
				Email:    "admin@example.com",
				Password: "adminPassword123!",
				Name:     "",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest(http.MethodPost, "/setup/admin", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response CreateAdminResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotNil(t, response.User)
			}
		})
	}
}

func TestSetupHandler_Complete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDBForSetupHandler(t)
	cfg := &config.Config{}
	setupService := services.NewSetupService(db, cfg)
	handler := NewSetupHandler(setupService, db)

	// Create admin first
	_, err := setupService.CreateInitialAdmin("admin@example.com", "password123!", "Admin User")
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/setup/complete", handler.Complete)

	tests := []struct {
		name           string
		request        CompleteSetupRequest
		expectedStatus int
	}{
		{
			name: "valid setup completion",
			request: CompleteSetupRequest{
				CompanyName: "Test Company",
				SystemSettings: map[string]string{
					"theme": "dark",
					"lang":  "ko",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "empty company name",
			request: CompleteSetupRequest{
				CompanyName: "",
				SystemSettings: map[string]string{
					"theme": "dark",
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest(http.MethodPost, "/setup/complete", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response CompleteSetupResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Completed)
				assert.NotEmpty(t, response.Message)
			}
		})
	}
}

func TestSetupHandler_CreateAdmin_AdminAlreadyExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDBForSetupHandler(t)
	cfg := &config.Config{}
	setupService := services.NewSetupService(db, cfg)
	handler := NewSetupHandler(setupService, db)

	// Create first admin
	_, err := setupService.CreateInitialAdmin("admin1@example.com", "password123!", "Admin 1")
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/setup/admin", handler.CreateAdmin)

	// Try to create second admin
	request := CreateAdminRequest{
		Email:    "admin2@example.com",
		Password: "password123!",
		Name:     "Admin 2",
	}

	body, _ := json.Marshal(request)
	req, _ := http.NewRequest(http.MethodPost, "/setup/admin", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ADMIN_ALREADY_EXISTS", response.Error)
}
