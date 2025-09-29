package handlers

import (
	"errors"
	"net/http"

	"github.com/burndler/burndler/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupHandler handles setup endpoints
type SetupHandler struct {
	setupService *services.SetupService
	db           *gorm.DB
}

// NewSetupHandler creates a new setup handler
func NewSetupHandler(setupService *services.SetupService, db *gorm.DB) *SetupHandler {
	return &SetupHandler{
		setupService: setupService,
		db:           db,
	}
}

// InitializeRequest represents the setup initialization request
type InitializeRequest struct {
	SetupToken string `json:"setup_token" binding:"required,min=32"`
}

// CreateAdminRequest represents the create admin request
type CreateAdminRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required,min=1"`
}

// CompleteSetupRequest represents the complete setup request
type CompleteSetupRequest struct {
	CompanyName    string            `json:"company_name" binding:"required,min=1"`
	SystemSettings map[string]string `json:"system_settings"`
}

// SetupStatusResponse represents the setup status response
type SetupStatusResponse struct {
	IsCompleted   bool   `json:"is_completed"`
	RequiresSetup bool   `json:"requires_setup"`
	AdminExists   bool   `json:"admin_exists"`
	SetupToken    string `json:"setup_token,omitempty"`
}

// CreateAdminResponse represents the create admin response
type CreateAdminResponse struct {
	User interface{} `json:"user"`
}

// CompleteSetupResponse represents the complete setup response
type CompleteSetupResponse struct {
	Message   string `json:"message"`
	Completed bool   `json:"completed"`
}

// GetStatus returns the current setup status
func (h *SetupHandler) GetStatus(c *gin.Context) {
	status, err := h.setupService.CheckSetupStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "SETUP_STATUS_ERROR",
			Message: "Failed to check setup status",
		})
		return
	}

	c.JSON(http.StatusOK, SetupStatusResponse{
		IsCompleted:   status.IsCompleted,
		RequiresSetup: status.RequiresSetup,
		AdminExists:   status.AdminExists,
		SetupToken:    status.SetupToken,
	})
}

// Initialize initializes the setup process with a valid token
func (h *SetupHandler) Initialize(c *gin.Context) {
	var req InitializeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request format or missing required fields",
		})
		return
	}

	// Validate setup token
	if err := h.setupService.ValidateSetupToken(req.SetupToken); err != nil {
		if errors.Is(err, services.ErrSetupAlreadyCompleted) {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "SETUP_ALREADY_COMPLETED",
				Message: "Setup has already been completed",
			})
			return
		}
		if errors.Is(err, services.ErrInvalidSetupToken) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "INVALID_SETUP_TOKEN",
				Message: "Invalid or expired setup token",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "SETUP_VALIDATION_ERROR",
			Message: "Failed to validate setup token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Setup initialized successfully",
		"initialized": true,
		"next_step":   "create_admin",
	})
}

// CreateAdmin creates the initial administrator account
func (h *SetupHandler) CreateAdmin(c *gin.Context) {
	var req CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request format or missing required fields",
		})
		return
	}

	// Create initial admin user
	admin, err := h.setupService.CreateInitialAdmin(req.Email, req.Password, req.Name)
	if err != nil {
		if errors.Is(err, services.ErrSetupAlreadyCompleted) {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "SETUP_ALREADY_COMPLETED",
				Message: "Setup has already been completed",
			})
			return
		}
		if errors.Is(err, services.ErrAdminAlreadyExists) {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "ADMIN_ALREADY_EXISTS",
				Message: "Administrator account already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "ADMIN_CREATION_FAILED",
			Message: "Failed to create administrator account",
		})
		return
	}

	c.JSON(http.StatusCreated, CreateAdminResponse{
		User: admin,
	})
}

// Complete completes the setup process
func (h *SetupHandler) Complete(c *gin.Context) {
	var req CompleteSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request format or missing required fields",
		})
		return
	}

	// Complete setup
	config := services.SetupConfig{
		CompanyName:    req.CompanyName,
		SystemSettings: req.SystemSettings,
	}

	if err := h.setupService.CompleteSetup(config); err != nil {
		if errors.Is(err, services.ErrSetupAlreadyCompleted) {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "SETUP_ALREADY_COMPLETED",
				Message: "Setup has already been completed",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "SETUP_COMPLETION_FAILED",
			Message: "Failed to complete setup",
		})
		return
	}

	c.JSON(http.StatusOK, CompleteSetupResponse{
		Message:   "Setup completed successfully",
		Completed: true,
	})
}
