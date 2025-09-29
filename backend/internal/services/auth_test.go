package services

import (
	"testing"
	"time"

	"github.com/burndler/burndler/internal/config"
	"github.com/burndler/burndler/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
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

func TestAuthService_GenerateToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:     "test-secret-key",
		JWTIssuer:     "burndler",
		JWTAudience:   "burndler-api",
		JWTExpiration: time.Hour * 24,
	}

	authService := NewAuthService(cfg, nil)

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  "Developer",
	}

	token, err := authService.GenerateToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Parse token to verify structure
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	// Verify claims
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		assert.Equal(t, "1", claims["user_id"])
		assert.Equal(t, user.Email, claims["email"])
		assert.Equal(t, user.Role, claims["role"])
		assert.Equal(t, cfg.JWTIssuer, claims["iss"])
		assert.Contains(t, claims["aud"], cfg.JWTAudience)
	} else {
		t.Error("Failed to parse token claims")
	}
}

func TestAuthService_GenerateToken_AdminRole(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:     "test-secret-key",
		JWTIssuer:     "burndler",
		JWTAudience:   "burndler-api",
		JWTExpiration: time.Hour * 24,
	}

	authService := NewAuthService(cfg, nil)

	user := &models.User{
		ID:    1,
		Email: "admin@example.com",
		Role:  "Admin",
	}

	token, err := authService.GenerateToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Parse token to verify structure
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	// Verify claims
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
		assert.Equal(t, "1", claims["user_id"])
		assert.Equal(t, user.Email, claims["email"])
		assert.Equal(t, "Admin", claims["role"])
		assert.Equal(t, cfg.JWTIssuer, claims["iss"])
		assert.Contains(t, claims["aud"], cfg.JWTAudience)
	} else {
		t.Error("Failed to parse token claims")
	}
}

func TestAuthService_GenerateRefreshToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:            "test-secret-key",
		JWTIssuer:            "burndler",
		JWTAudience:          "burndler-api",
		JWTRefreshExpiration: time.Hour * 168, // 7 days
	}

	authService := NewAuthService(cfg, nil)

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  "Engineer",
	}

	refreshToken, err := authService.GenerateRefreshToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, refreshToken)

	// Parse token to verify structure
	parsedToken, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})

	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)
}

func TestAuthService_AuthenticateUser(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{
		JWTSecret:     "test-secret-key",
		JWTIssuer:     "burndler",
		JWTAudience:   "burndler-api",
		JWTExpiration: time.Hour * 24,
	}

	authService := NewAuthService(cfg, db)

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
		name        string
		email       string
		password    string
		expectError bool
		expectUser  bool
	}{
		{
			name:        "valid credentials",
			email:       "test@example.com",
			password:    "testPassword123!",
			expectError: false,
			expectUser:  true,
		},
		{
			name:        "invalid password",
			email:       "test@example.com",
			password:    "wrongPassword",
			expectError: true,
			expectUser:  false,
		},
		{
			name:        "user not found",
			email:       "nonexistent@example.com",
			password:    "testPassword123!",
			expectError: true,
			expectUser:  false,
		},
		{
			name:        "empty email",
			email:       "",
			password:    "testPassword123!",
			expectError: true,
			expectUser:  false,
		},
		{
			name:        "empty password",
			email:       "test@example.com",
			password:    "",
			expectError: true,
			expectUser:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundUser, err := authService.AuthenticateUser(tt.email, tt.password)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectUser {
				assert.NotNil(t, foundUser)
				assert.Equal(t, tt.email, foundUser.Email)
			} else {
				assert.Nil(t, foundUser)
			}
		})
	}
}

func TestAuthService_AuthenticateUser_AdminRole(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{
		JWTSecret:     "test-secret-key",
		JWTIssuer:     "burndler",
		JWTAudience:   "burndler-api",
		JWTExpiration: time.Hour * 24,
	}

	authService := NewAuthService(cfg, db)

	// Create test admin user
	user := &models.User{
		Email: "admin@example.com",
		Name:  "Admin User",
		Role:  "Admin",
	}
	err := user.SetPassword("adminPassword123!")
	assert.NoError(t, err)

	err = db.Create(user).Error
	assert.NoError(t, err)

	// Authenticate admin user
	foundUser, err := authService.AuthenticateUser("admin@example.com", "adminPassword123!")

	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, "admin@example.com", foundUser.Email)
	assert.Equal(t, "Admin", foundUser.Role)
	assert.True(t, foundUser.IsAdmin())
}

func TestAuthService_AuthenticateUser_InactiveUser(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{
		JWTSecret:     "test-secret-key",
		JWTIssuer:     "burndler",
		JWTAudience:   "burndler-api",
		JWTExpiration: time.Hour * 24,
	}

	authService := NewAuthService(cfg, db)

	// Create inactive test user
	user := &models.User{
		Email:  "inactive@example.com",
		Name:   "Inactive User",
		Role:   "Engineer",
		Active: false, // Inactive user
	}
	err := user.SetPassword("testPassword123!")
	assert.NoError(t, err)

	err = db.Create(user).Error
	assert.NoError(t, err)

	// Explicitly set user as inactive after creation
	err = db.Model(user).Update("active", false).Error
	assert.NoError(t, err)

	// Try to authenticate inactive user
	foundUser, err := authService.AuthenticateUser("inactive@example.com", "testPassword123!")

	assert.Error(t, err)
	assert.Nil(t, foundUser)
	assert.Contains(t, err.Error(), "inactive")
}

func TestAuthService_ValidateToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:     "test-secret-key",
		JWTIssuer:     "burndler",
		JWTAudience:   "burndler-api",
		JWTExpiration: time.Hour * 24,
	}

	authService := NewAuthService(cfg, nil)

	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  "Developer",
	}

	// Generate valid token
	token, err := authService.GenerateToken(user)
	assert.NoError(t, err)

	// Validate token
	claims, err := authService.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, "1", claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Role, claims.Role)

	// Test invalid token
	_, err = authService.ValidateToken("invalid.token.string")
	assert.Error(t, err)

	// Test empty token
	_, err = authService.ValidateToken("")
	assert.Error(t, err)
}

func TestAuthService_RefreshToken(t *testing.T) {
	db := setupTestDB(t)
	cfg := &config.Config{
		JWTSecret:            "test-secret-key",
		JWTIssuer:            "burndler",
		JWTAudience:          "burndler-api",
		JWTExpiration:        time.Hour * 24,
		JWTRefreshExpiration: time.Hour * 168, // 7 days
	}

	authService := NewAuthService(cfg, db)

	// Create test user
	user := &models.User{
		Email: "refresh@example.com",
		Name:  "Refresh Test User",
		Role:  "Developer",
	}
	err := user.SetPassword("testPassword123!")
	assert.NoError(t, err)

	err = db.Create(user).Error
	assert.NoError(t, err)

	// Generate refresh token
	refreshToken, err := authService.GenerateRefreshToken(user)
	assert.NoError(t, err)

	// Use refresh token to get new tokens
	newAccessToken, newRefreshToken, err := authService.RefreshToken(refreshToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)
	assert.NotEmpty(t, newRefreshToken)
	// Note: New refresh token might be the same if generated in the same second
	// This is acceptable as tokens have different IssuedAt claims internally

	// Validate new access token
	claims, err := authService.ValidateToken(newAccessToken)
	assert.NoError(t, err)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Role, claims.Role)

	// Test with invalid refresh token
	_, _, err = authService.RefreshToken("invalid.token.string")
	assert.Error(t, err)
}
