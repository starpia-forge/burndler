package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/burndler/burndler/internal/config"
	"github.com/burndler/burndler/internal/models"
	"github.com/burndler/burndler/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDBForAuth(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDBForAuth(t)
	cfg := &config.Config{
		JWTSecret:            "test-secret-key",
		JWTIssuer:            "burndler",
		JWTAudience:          "burndler-api",
		JWTExpiration:        time.Hour * 24,
		JWTRefreshExpiration: time.Hour * 168,
	}

	authService := services.NewAuthService(cfg, db)
	authHandler := NewAuthHandler(authService, db)

	// Create test user
	user := &models.User{
		Email: "test@example.com",
		Name:  "Test User",
		Role:  "Developer",
	}
	err := user.SetPassword("testPassword123!")
	assert.NoError(t, err)

	err = db.Create(user).Error
	assert.NoError(t, err)

	tests := []struct {
		name            string
		requestBody     interface{}
		expectedStatus  int
		expectedFields  []string // Fields that should be present in response
		shouldHaveToken bool
	}{
		{
			name: "valid login",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "testPassword123!",
			},
			expectedStatus:  http.StatusOK,
			expectedFields:  []string{"accessToken", "refreshToken", "user"},
			shouldHaveToken: true,
		},
		{
			name: "invalid password",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "wrongPassword",
			},
			expectedStatus:  http.StatusUnauthorized,
			expectedFields:  []string{"error"},
			shouldHaveToken: false,
		},
		{
			name: "user not found",
			requestBody: map[string]string{
				"email":    "nonexistent@example.com",
				"password": "testPassword123!",
			},
			expectedStatus:  http.StatusUnauthorized,
			expectedFields:  []string{"error"},
			shouldHaveToken: false,
		},
		{
			name: "missing email",
			requestBody: map[string]string{
				"password": "testPassword123!",
			},
			expectedStatus:  http.StatusBadRequest,
			expectedFields:  []string{"error"},
			shouldHaveToken: false,
		},
		{
			name: "missing password",
			requestBody: map[string]string{
				"email": "test@example.com",
			},
			expectedStatus:  http.StatusBadRequest,
			expectedFields:  []string{"error"},
			shouldHaveToken: false,
		},
		{
			name: "empty email",
			requestBody: map[string]string{
				"email":    "",
				"password": "testPassword123!",
			},
			expectedStatus:  http.StatusBadRequest,
			expectedFields:  []string{"error"},
			shouldHaveToken: false,
		},
		{
			name: "empty password",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "",
			},
			expectedStatus:  http.StatusBadRequest,
			expectedFields:  []string{"error"},
			shouldHaveToken: false,
		},
		{
			name:            "invalid JSON",
			requestBody:     "invalid json",
			expectedStatus:  http.StatusBadRequest,
			expectedFields:  []string{"error"},
			shouldHaveToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/auth/login", authHandler.Login)

			var reqBody []byte
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, _ = json.Marshal(tt.requestBody)
			}

			req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Check expected fields are present
			for _, field := range tt.expectedFields {
				assert.Contains(t, response, field, "Response should contain field: %s", field)
			}

			// Check token presence
			if tt.shouldHaveToken {
				assert.NotEmpty(t, response["accessToken"])
				assert.NotEmpty(t, response["refreshToken"])
				assert.Contains(t, response, "user")

				// Verify user data doesn't contain password
				user, ok := response["user"].(map[string]interface{})
				assert.True(t, ok)
				assert.NotContains(t, user, "password")
			} else {
				assert.NotContains(t, response, "accessToken")
				assert.NotContains(t, response, "refreshToken")
			}
		})
	}
}

func TestAuthHandler_Login_InactiveUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDBForAuth(t)
	cfg := &config.Config{
		JWTSecret:            "test-secret-key",
		JWTIssuer:            "burndler",
		JWTAudience:          "burndler-api",
		JWTExpiration:        time.Hour * 24,
		JWTRefreshExpiration: time.Hour * 168,
	}

	authService := services.NewAuthService(cfg, db)
	authHandler := NewAuthHandler(authService, db)

	// Create inactive test user
	user := &models.User{
		Email: "inactive@example.com",
		Name:  "Inactive User",
		Role:  "Engineer",
	}
	err := user.SetPassword("testPassword123!")
	assert.NoError(t, err)

	err = db.Create(user).Error
	assert.NoError(t, err)

	// Set user as inactive
	err = db.Model(user).Update("active", false).Error
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/auth/login", authHandler.Login)

	requestBody := map[string]string{
		"email":    "inactive@example.com",
		"password": "testPassword123!",
	}
	reqBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.NotContains(t, response, "accessToken")
	assert.NotContains(t, response, "refreshToken")
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDBForAuth(t)
	cfg := &config.Config{
		JWTSecret:            "test-secret-key",
		JWTIssuer:            "burndler",
		JWTAudience:          "burndler-api",
		JWTExpiration:        time.Hour * 24,
		JWTRefreshExpiration: time.Hour * 168,
	}

	authService := services.NewAuthService(cfg, db)
	authHandler := NewAuthHandler(authService, db)

	// Create test user
	user := &models.User{
		Email: "refresh@example.com",
		Name:  "Refresh User",
		Role:  "Developer",
	}
	err := user.SetPassword("testPassword123!")
	assert.NoError(t, err)

	err = db.Create(user).Error
	assert.NoError(t, err)

	// Generate a valid refresh token
	validRefreshToken, err := authService.GenerateRefreshToken(user)
	assert.NoError(t, err)

	tests := []struct {
		name            string
		requestBody     interface{}
		expectedStatus  int
		expectedFields  []string
		shouldHaveToken bool
	}{
		{
			name: "valid refresh token",
			requestBody: map[string]string{
				"refreshToken": validRefreshToken,
			},
			expectedStatus:  http.StatusOK,
			expectedFields:  []string{"accessToken", "refreshToken"},
			shouldHaveToken: true,
		},
		{
			name:            "missing refresh token",
			requestBody:     map[string]string{},
			expectedStatus:  http.StatusBadRequest,
			expectedFields:  []string{"error"},
			shouldHaveToken: false,
		},
		{
			name: "empty refresh token",
			requestBody: map[string]string{
				"refreshToken": "",
			},
			expectedStatus:  http.StatusBadRequest,
			expectedFields:  []string{"error"},
			shouldHaveToken: false,
		},
		{
			name: "invalid refresh token",
			requestBody: map[string]string{
				"refreshToken": "invalid.token.string",
			},
			expectedStatus:  http.StatusUnauthorized,
			expectedFields:  []string{"error"},
			shouldHaveToken: false,
		},
		{
			name:            "invalid JSON",
			requestBody:     "invalid json",
			expectedStatus:  http.StatusBadRequest,
			expectedFields:  []string{"error"},
			shouldHaveToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/auth/refresh", authHandler.RefreshToken)

			var reqBody []byte
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, _ = json.Marshal(tt.requestBody)
			}

			req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Check expected fields are present
			for _, field := range tt.expectedFields {
				assert.Contains(t, response, field, "Response should contain field: %s", field)
			}

			// Check token presence
			if tt.shouldHaveToken {
				assert.NotEmpty(t, response["accessToken"])
				assert.NotEmpty(t, response["refreshToken"])
			} else {
				assert.NotContains(t, response, "accessToken")
				assert.NotContains(t, response, "refreshToken")
			}
		})
	}
}
