package services

import (
	"testing"

	"github.com/burndler/burndler/internal/config"
	"github.com/burndler/burndler/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDBForSetup(t *testing.T) *gorm.DB {
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

func TestNewSetupService(t *testing.T) {
	db := setupTestDBForSetup(t)
	cfg := &config.Config{}

	service := NewSetupService(db, cfg)

	assert.NotNil(t, service)
	assert.Equal(t, db, service.db)
	assert.Equal(t, cfg, service.config)
}

func TestSetupService_CheckSetupStatus(t *testing.T) {
	db := setupTestDBForSetup(t)
	cfg := &config.Config{}
	service := NewSetupService(db, cfg)

	// Test initial state
	status, err := service.CheckSetupStatus()

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.False(t, status.IsCompleted)
	assert.True(t, status.RequiresSetup)
	assert.False(t, status.AdminExists)
	assert.NotEmpty(t, status.SetupToken)
}

func TestSetupService_CreateInitialAdmin(t *testing.T) {
	db := setupTestDBForSetup(t)
	cfg := &config.Config{}
	service := NewSetupService(db, cfg)

	// Test creating initial admin
	email := "admin@example.com"
	password := "adminPassword123!"
	name := "Admin User"

	admin, err := service.CreateInitialAdmin(email, password, name)

	assert.NoError(t, err)
	assert.NotNil(t, admin)
	assert.Equal(t, email, admin.Email)
	assert.Equal(t, name, admin.Name)
	assert.Equal(t, "Admin", admin.Role)
	assert.True(t, admin.Active)
	assert.True(t, admin.CheckPassword(password))
}

func TestSetupService_CreateInitialAdmin_AlreadyExists(t *testing.T) {
	db := setupTestDBForSetup(t)
	cfg := &config.Config{}
	service := NewSetupService(db, cfg)

	// Create first admin
	_, err := service.CreateInitialAdmin("admin1@example.com", "password1", "Admin 1")
	assert.NoError(t, err)

	// Try to create second admin
	_, err = service.CreateInitialAdmin("admin2@example.com", "password2", "Admin 2")
	assert.Error(t, err)
	assert.Equal(t, ErrAdminAlreadyExists, err)
}

func TestSetupService_CompleteSetup(t *testing.T) {
	db := setupTestDBForSetup(t)
	cfg := &config.Config{}
	service := NewSetupService(db, cfg)

	// Create admin first
	_, err := service.CreateInitialAdmin("admin@example.com", "password", "Admin")
	assert.NoError(t, err)

	// Complete setup
	config := SetupConfig{
		CompanyName: "Test Company",
		SystemSettings: map[string]string{
			"theme": "dark",
			"lang":  "ko",
		},
	}

	err = service.CompleteSetup(config)
	assert.NoError(t, err)

	// Check status after completion
	status, err := service.CheckSetupStatus()
	assert.NoError(t, err)
	assert.True(t, status.IsCompleted)
	assert.False(t, status.RequiresSetup)
	assert.True(t, status.AdminExists)
	assert.Empty(t, status.SetupToken) // No token when completed
}

func TestSetupService_CompleteSetup_NoAdmin(t *testing.T) {
	db := setupTestDBForSetup(t)
	cfg := &config.Config{}
	service := NewSetupService(db, cfg)

	// Try to complete setup without admin
	config := SetupConfig{
		CompanyName: "Test Company",
	}

	err := service.CompleteSetup(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "admin user must be created")
}

func TestSetupService_CompleteSetup_AlreadyCompleted(t *testing.T) {
	db := setupTestDBForSetup(t)
	cfg := &config.Config{}
	service := NewSetupService(db, cfg)

	// Create admin and complete setup
	_, err := service.CreateInitialAdmin("admin@example.com", "password", "Admin")
	assert.NoError(t, err)

	config := SetupConfig{
		CompanyName: "Test Company",
	}

	err = service.CompleteSetup(config)
	assert.NoError(t, err)

	// Try to complete again
	err = service.CompleteSetup(config)
	assert.Error(t, err)
	assert.Equal(t, ErrSetupAlreadyCompleted, err)
}

func TestSetupService_IsSetupCompleted(t *testing.T) {
	db := setupTestDBForSetup(t)
	cfg := &config.Config{}
	service := NewSetupService(db, cfg)

	// Initially not completed
	completed, err := service.IsSetupCompleted()
	assert.NoError(t, err)
	assert.False(t, completed)

	// Complete setup
	_, err = service.CreateInitialAdmin("admin@example.com", "password", "Admin")
	assert.NoError(t, err)

	config := SetupConfig{
		CompanyName: "Test Company",
	}
	err = service.CompleteSetup(config)
	assert.NoError(t, err)

	// Should be completed now
	completed, err = service.IsSetupCompleted()
	assert.NoError(t, err)
	assert.True(t, completed)
}

func TestSetupService_ValidateSetupToken(t *testing.T) {
	db := setupTestDBForSetup(t)
	cfg := &config.Config{}
	service := NewSetupService(db, cfg)

	// Test valid token (simple length validation for now)
	validToken := "1234567890abcdef1234567890abcdef12345678"
	err := service.ValidateSetupToken(validToken)
	assert.NoError(t, err)

	// Test invalid token (too short)
	invalidToken := "short"
	err = service.ValidateSetupToken(invalidToken)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidSetupToken, err)

	// Complete setup and test token validation
	_, err = service.CreateInitialAdmin("admin@example.com", "password", "Admin")
	assert.NoError(t, err)

	config := SetupConfig{CompanyName: "Test Company"}
	err = service.CompleteSetup(config)
	assert.NoError(t, err)

	// Should return error when setup is completed
	err = service.ValidateSetupToken(validToken)
	assert.Error(t, err)
	assert.Equal(t, ErrSetupAlreadyCompleted, err)
}
