package services

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/burndler/burndler/internal/config"
	"github.com/burndler/burndler/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SetupService provides setup-related operations
type SetupService struct {
	db     *gorm.DB
	config *config.Config
}

// NewSetupService creates a new setup service
func NewSetupService(db *gorm.DB, cfg *config.Config) *SetupService {
	return &SetupService{
		db:     db,
		config: cfg,
	}
}

// SetupConfig represents the system configuration during setup
type SetupConfig struct {
	CompanyName    string            `json:"company_name"`
	SystemSettings map[string]string `json:"system_settings"`
}

// SetupStatus represents the current setup status
type SetupStatus struct {
	IsCompleted   bool   `json:"is_completed"`
	RequiresSetup bool   `json:"requires_setup"`
	AdminExists   bool   `json:"admin_exists"`
	SetupToken    string `json:"setup_token,omitempty"`
}

var (
	// ErrSetupAlreadyCompleted is returned when trying to setup an already completed system
	ErrSetupAlreadyCompleted = errors.New("setup already completed")
	// ErrInvalidSetupToken is returned when an invalid setup token is provided
	ErrInvalidSetupToken = errors.New("invalid setup token")
	// ErrAdminAlreadyExists is returned when trying to create admin but one already exists
	ErrAdminAlreadyExists = errors.New("admin user already exists")
)

// CheckSetupStatus returns the current setup status
func (s *SetupService) CheckSetupStatus() (*SetupStatus, error) {
	setup, err := s.getOrCreateSetup()
	if err != nil {
		return nil, fmt.Errorf("failed to check setup status: %w", err)
	}

	// Check if any admin users exist
	var adminCount int64
	err = s.db.Model(&models.User{}).Where("role = ?", "Admin").Count(&adminCount).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count admin users: %w", err)
	}

	status := &SetupStatus{
		IsCompleted:   setup.IsCompleted,
		RequiresSetup: !setup.IsCompleted,
		AdminExists:   adminCount > 0,
	}

	// Generate setup token if setup is not completed
	if !setup.IsCompleted {
		token, err := s.generateSetupToken()
		if err != nil {
			return nil, fmt.Errorf("failed to generate setup token: %w", err)
		}
		status.SetupToken = token
	}

	return status, nil
}

// ValidateSetupToken validates a setup token
func (s *SetupService) ValidateSetupToken(token string) error {
	setup, err := s.getOrCreateSetup()
	if err != nil {
		return fmt.Errorf("failed to get setup status: %w", err)
	}

	if setup.IsCompleted {
		return ErrSetupAlreadyCompleted
	}

	// For now, we'll use a simple validation
	// In production, you might want to store tokens in Redis or database with expiration
	if len(token) < 32 {
		return ErrInvalidSetupToken
	}

	return nil
}

// CreateInitialAdmin creates the initial admin user
func (s *SetupService) CreateInitialAdmin(email, password, name string) (*models.User, error) {
	setup, err := s.getOrCreateSetup()
	if err != nil {
		return nil, fmt.Errorf("failed to get setup status: %w", err)
	}

	if setup.IsCompleted {
		return nil, ErrSetupAlreadyCompleted
	}

	// Check if admin already exists
	var adminCount int64
	err = s.db.Model(&models.User{}).Where("role = ?", "Admin").Count(&adminCount).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count admin users: %w", err)
	}

	if adminCount > 0 {
		return nil, ErrAdminAlreadyExists
	}

	// Create admin user
	admin := &models.User{
		Email:  email,
		Name:   name,
		Role:   "Admin",
		Active: true,
	}

	if err := admin.SetPassword(password); err != nil {
		return nil, fmt.Errorf("failed to set password: %w", err)
	}

	if err := s.db.Create(admin).Error; err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	// Update setup record with admin email
	setup.AdminEmail = email
	if err := s.db.Save(setup).Error; err != nil {
		return nil, fmt.Errorf("failed to update setup record: %w", err)
	}

	return admin, nil
}

// CompleteSetup completes the setup process with the provided configuration
func (s *SetupService) CompleteSetup(config SetupConfig) error {
	setup, err := s.getOrCreateSetup()
	if err != nil {
		return fmt.Errorf("failed to get setup status: %w", err)
	}

	if setup.IsCompleted {
		return ErrSetupAlreadyCompleted
	}

	// Ensure admin exists
	var adminCount int64
	err = s.db.Model(&models.User{}).Where("role = ?", "Admin").Count(&adminCount).Error
	if err != nil {
		return fmt.Errorf("failed to count admin users: %w", err)
	}

	if adminCount == 0 {
		return errors.New("admin user must be created before completing setup")
	}

	// Convert config to JSON
	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	configJSON := datatypes.JSON(configBytes)

	// Update setup record
	setup.CompanyName = config.CompanyName
	setup.SystemConfig = configJSON
	setup.MarkCompleted()

	if err := s.db.Save(setup).Error; err != nil {
		return fmt.Errorf("failed to complete setup: %w", err)
	}

	return nil
}

// IsSetupCompleted returns whether the setup is completed
func (s *SetupService) IsSetupCompleted() (bool, error) {
	setup, err := s.getOrCreateSetup()
	if err != nil {
		return false, fmt.Errorf("failed to check setup status: %w", err)
	}

	return setup.IsCompleted, nil
}

// getOrCreateSetup gets the setup record or creates one if it doesn't exist
func (s *SetupService) getOrCreateSetup() (*models.Setup, error) {
	var setup models.Setup
	err := s.db.First(&setup).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create initial setup record
			setup = models.Setup{
				IsCompleted: false,
			}
			if err := s.db.Create(&setup).Error; err != nil {
				return nil, fmt.Errorf("failed to create setup record: %w", err)
			}
			return &setup, nil
		}
		return nil, fmt.Errorf("failed to get setup record: %w", err)
	}

	return &setup, nil
}

// generateSetupToken generates a random setup token
func (s *SetupService) generateSetupToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
