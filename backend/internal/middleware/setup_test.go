package middleware

import (
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

func setupTestDBForSetupMiddleware(t *testing.T) *gorm.DB {
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

func TestSetupGuard_SetupNotCompleted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDBForSetupMiddleware(t)
	cfg := &config.Config{}
	setupService := services.NewSetupService(db, cfg)

	router := gin.New()
	router.Use(SetupGuard(setupService))
	router.GET("/api/v1/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "protected"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "SETUP_REQUIRED", response["error"])
	assert.Equal(t, true, response["requires_setup"])
}

func TestSetupGuard_SetupCompleted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDBForSetupMiddleware(t)
	cfg := &config.Config{}
	setupService := services.NewSetupService(db, cfg)

	// Complete setup
	_, err := setupService.CreateInitialAdmin("admin@example.com", "password123!", "Admin")
	assert.NoError(t, err)

	config := services.SetupConfig{
		CompanyName: "Test Company",
	}
	err = setupService.CompleteSetup(config)
	assert.NoError(t, err)

	router := gin.New()
	router.Use(SetupGuard(setupService))
	router.GET("/api/v1/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "protected"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "protected", response["message"])
}

func TestSetupGuard_SetupEndpointsAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDBForSetupMiddleware(t)
	cfg := &config.Config{}
	setupService := services.NewSetupService(db, cfg)

	router := gin.New()
	router.Use(SetupGuard(setupService))

	setupEndpoints := []string{
		"/api/v1/setup/status",
		"/api/v1/setup/init",
		"/api/v1/setup/admin",
		"/api/v1/setup/complete",
	}

	for _, endpoint := range setupEndpoints {
		router.GET(endpoint, func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "setup endpoint"})
		})
		router.POST(endpoint, func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "setup endpoint"})
		})
	}

	for _, endpoint := range setupEndpoints {
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			req, _ := http.NewRequest(method, endpoint, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Endpoint %s %s should be allowed", method, endpoint)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "setup endpoint", response["message"])
		}
	}
}

func TestSetupGuard_HealthEndpointAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDBForSetupMiddleware(t)
	cfg := &config.Config{}
	setupService := services.NewSetupService(db, cfg)

	router := gin.New()
	router.Use(SetupGuard(setupService))
	router.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "healthy"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["message"])
}

func TestSetupCompleteGuard_SetupNotCompleted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDBForSetupMiddleware(t)
	cfg := &config.Config{}
	setupService := services.NewSetupService(db, cfg)

	router := gin.New()
	router.Use(SetupCompleteGuard(setupService))
	router.POST("/api/v1/setup/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "setup endpoint"})
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/setup/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "setup endpoint", response["message"])
}

func TestSetupCompleteGuard_SetupCompleted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDBForSetupMiddleware(t)
	cfg := &config.Config{}
	setupService := services.NewSetupService(db, cfg)

	// Complete setup
	_, err := setupService.CreateInitialAdmin("admin@example.com", "password123!", "Admin")
	assert.NoError(t, err)

	config := services.SetupConfig{
		CompanyName: "Test Company",
	}
	err = setupService.CompleteSetup(config)
	assert.NoError(t, err)

	router := gin.New()
	router.Use(SetupCompleteGuard(setupService))
	router.POST("/api/v1/setup/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "setup endpoint"})
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/setup/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "SETUP_ALREADY_COMPLETED", response["error"])
	assert.Equal(t, true, response["is_completed"])
}

func TestSetupCompleteGuard_StatusEndpointAlwaysAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDBForSetupMiddleware(t)
	cfg := &config.Config{}
	setupService := services.NewSetupService(db, cfg)

	// Complete setup
	_, err := setupService.CreateInitialAdmin("admin@example.com", "password123!", "Admin")
	assert.NoError(t, err)

	config := services.SetupConfig{
		CompanyName: "Test Company",
	}
	err = setupService.CompleteSetup(config)
	assert.NoError(t, err)

	router := gin.New()
	router.Use(SetupCompleteGuard(setupService))
	router.GET("/api/v1/setup/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "status endpoint"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/setup/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "status endpoint", response["message"])
}

func TestSetupCompleteGuard_NonSetupEndpointAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDBForSetupMiddleware(t)
	cfg := &config.Config{}
	setupService := services.NewSetupService(db, cfg)

	// Complete setup
	_, err := setupService.CreateInitialAdmin("admin@example.com", "password123!", "Admin")
	assert.NoError(t, err)

	config := services.SetupConfig{
		CompanyName: "Test Company",
	}
	err = setupService.CompleteSetup(config)
	assert.NoError(t, err)

	router := gin.New()
	router.Use(SetupCompleteGuard(setupService))
	router.GET("/api/v1/other", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "other endpoint"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/other", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "other endpoint", response["message"])
}

func TestSetupCompleteGuard_InconsistentState_AllowsAdminCreation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDBForSetupMiddleware(t)
	cfg := &config.Config{}
	setupService := services.NewSetupService(db, cfg)

	// Manually create an inconsistent state: setup marked complete but no admin exists
	// This simulates a scenario where setup completion was marked but admin creation failed

	// First trigger setup record creation by checking status
	_, err := setupService.CheckSetupStatus()
	assert.NoError(t, err)

	// Force setup to be marked as completed without creating an admin
	var setupRecord models.Setup
	err = db.First(&setupRecord).Error
	assert.NoError(t, err)

	setupRecord.MarkCompleted()
	err = db.Save(&setupRecord).Error
	assert.NoError(t, err)

	// Verify setup is marked complete but no admin exists
	isCompleted, err := setupService.IsSetupCompleted()
	assert.NoError(t, err)
	assert.True(t, isCompleted)

	adminExists, err := setupService.CheckAdminExists()
	assert.NoError(t, err)
	assert.False(t, adminExists)

	router := gin.New()
	router.Use(SetupCompleteGuard(setupService))
	router.POST("/api/v1/setup/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin creation allowed"})
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/setup/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should allow admin creation despite setup being marked complete
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "admin creation allowed", response["message"])
}

func TestSetupCompleteGuard_InconsistentState_BlocksOtherSetupEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDBForSetupMiddleware(t)
	cfg := &config.Config{}
	setupService := services.NewSetupService(db, cfg)

	// Manually create an inconsistent state: setup marked complete but no admin exists
	// First trigger setup record creation by checking status
	_, err := setupService.CheckSetupStatus()
	assert.NoError(t, err)

	var setupRecord models.Setup
	err = db.First(&setupRecord).Error
	assert.NoError(t, err)

	setupRecord.MarkCompleted()
	err = db.Save(&setupRecord).Error
	assert.NoError(t, err)

	router := gin.New()
	router.Use(SetupCompleteGuard(setupService))
	router.POST("/api/v1/setup/complete", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "setup complete endpoint"})
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/setup/complete", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should still block other setup endpoints
	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "SETUP_ALREADY_COMPLETED", response["error"])
}

func TestIsSetupEndpoint(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/api/v1/setup/status", true},
		{"/api/v1/setup/init", true},
		{"/api/v1/setup/admin", true},
		{"/api/v1/setup/complete", true},
		{"/api/v1/auth/login", false},
		{"/api/v1/health", false},
		{"/api/v1/compose/merge", false},
		{"/setup/something", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isSetupEndpoint(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}
