package middleware

import (
	"net/http"
	"strings"

	"github.com/burndler/burndler/internal/services"
	"github.com/gin-gonic/gin"
)

// SetupGuard middleware ensures setup is completed before allowing access to protected endpoints
func SetupGuard(setupService *services.SetupService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip setup guard for setup endpoints
		if isSetupEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Skip setup guard for health endpoint
		if c.Request.URL.Path == "/api/v1/health" {
			c.Next()
			return
		}

		// Check if setup is completed
		isCompleted, err := setupService.IsSetupCompleted()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "SETUP_CHECK_FAILED",
				"message": "Failed to check setup status",
			})
			c.Abort()
			return
		}

		// If setup is not completed, block access to all other endpoints
		if !isCompleted {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":          "SETUP_REQUIRED",
				"message":        "System setup is required before accessing this resource",
				"requires_setup": true,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SetupCompleteGuard middleware blocks access to setup endpoints after setup is completed
func SetupCompleteGuard(setupService *services.SetupService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to setup endpoints
		if !isSetupEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Allow GET /setup/status even after setup is completed (for status checks)
		if c.Request.Method == http.MethodGet && c.Request.URL.Path == "/api/v1/setup/status" {
			c.Next()
			return
		}

		// Check if setup is completed
		isCompleted, err := setupService.IsSetupCompleted()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "SETUP_CHECK_FAILED",
				"message": "Failed to check setup status",
			})
			c.Abort()
			return
		}

		// If setup is completed, block access to setup modification endpoints
		if isCompleted {
			// Special case: Allow admin creation if no admin exists (inconsistent state recovery)
			if c.Request.Method == http.MethodPost && c.Request.URL.Path == "/api/v1/setup/admin" {
				adminExists, err := setupService.CheckAdminExists()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "ADMIN_CHECK_FAILED",
						"message": "Failed to check admin existence",
					})
					c.Abort()
					return
				}

				// If no admin exists despite setup being marked complete, allow admin creation
				// This handles database inconsistency where setup was marked complete but admin creation failed
				if !adminExists {
					c.Next()
					return
				}
			}

			c.JSON(http.StatusForbidden, gin.H{
				"error":        "SETUP_ALREADY_COMPLETED",
				"message":      "Setup has already been completed and cannot be modified",
				"is_completed": true,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// isSetupEndpoint checks if the given path is a setup endpoint
func isSetupEndpoint(path string) bool {
	setupPaths := []string{
		"/api/v1/setup/",
	}

	for _, setupPath := range setupPaths {
		if strings.HasPrefix(path, setupPath) {
			return true
		}
	}

	return false
}
