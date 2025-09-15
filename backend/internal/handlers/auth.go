package handlers

import (
	"errors"
	"net/http"

	"github.com/burndler/burndler/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *services.AuthService
	db          *gorm.DB
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *services.AuthService, db *gorm.DB) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		db:          db,
	}
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=1"`
}

// RefreshTokenRequest represents the refresh token request body
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required,min=1"`
}

// LoginResponse represents the successful login response
type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         interface{} `json:"user"`
}

// RefreshTokenResponse represents the successful refresh token response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Login handles user authentication
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request format or missing required fields",
		})
		return
	}

	// Authenticate user
	user, err := h.authService.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) || errors.Is(err, services.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "AUTHENTICATION_FAILED",
				Message: "Invalid email or password",
			})
			return
		}
		if errors.Is(err, services.ErrUserInactive) {
			c.JSON(http.StatusForbidden, ErrorResponse{
				Error:   "ACCOUNT_INACTIVE",
				Message: "Your account is inactive. Please contact an administrator",
			})
			return
		}
		// Internal server error for other types of errors
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "An internal error occurred",
		})
		return
	}

	// Generate tokens
	accessToken, err := h.authService.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "TOKEN_GENERATION_FAILED",
			Message: "Failed to generate access token",
		})
		return
	}

	refreshToken, err := h.authService.GenerateRefreshToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "TOKEN_GENERATION_FAILED",
			Message: "Failed to generate refresh token",
		})
		return
	}

	// Return successful response with user data (password excluded by model's json tag)
	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request format or missing refresh token",
		})
		return
	}

	// Generate new tokens using the refresh token
	newAccessToken, newRefreshToken, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		// Check for specific error types and handle them appropriately
		if errors.Is(err, services.ErrInvalidToken) ||
		   errors.Is(err, services.ErrUserNotFound) ||
		   errors.Is(err, services.ErrUserInactive) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "INVALID_REFRESH_TOKEN",
				Message: "Invalid or expired refresh token",
			})
			return
		}
		// Check if error contains "invalid" (for token parsing errors)
		errorStr := err.Error()
		if len(errorStr) > 0 && (
			errors.Is(err, services.ErrInvalidToken) ||
			errorStr == "invalid token" ||
			errorStr == "invalid refresh token: invalid token" ||
			errorStr[:7] == "invalid" ||
			errorStr[:14] == "token parsing error") {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "INVALID_REFRESH_TOKEN",
				Message: "Invalid or expired refresh token",
			})
			return
		}
		// Internal server error for other types of errors
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "An internal error occurred",
		})
		return
	}

	// Return new tokens
	c.JSON(http.StatusOK, RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	})
}