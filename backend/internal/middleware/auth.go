package middleware

import (
	"net/http"
	"strings"

	"github.com/burndler/burndler/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims with user role
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"` // Developer, Engineer, or Admin
	jwt.RegisteredClaims
}

// JWTAuth middleware validates JWT tokens
func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "UNAUTHORIZED",
				"message": "Missing authorization header",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "INVALID_TOKEN_FORMAT",
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "INVALID_TOKEN",
				"message": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*Claims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "INVALID_CLAIMS",
				"message": "Invalid token claims",
			})
			c.Abort()
			return
		}

		// Validate issuer and audience
		if claims.Issuer != cfg.JWTIssuer || !contains(claims.Audience, cfg.JWTAudience) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "INVALID_TOKEN_SCOPE",
				"message": "Token not valid for this service",
			})
			c.Abort()
			return
		}

		// Validate role
		if claims.Role != "Developer" && claims.Role != "Engineer" && claims.Role != "Admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "INVALID_ROLE",
				"message": "Invalid user role",
			})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRole checks if the user has the required role
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "NO_ROLE",
				"message": "User role not found",
			})
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "INVALID_ROLE_TYPE",
				"message": "Invalid role type in context",
			})
			c.Abort()
			return
		}

		// Check role hierarchy
		// Admin has full access to everything
		// Developer has full access (read/write)
		// Engineer has read-only access
		if requiredRole == "Developer" && userRole != "Developer" && userRole != "Admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "INSUFFICIENT_PERMISSIONS",
				"message": "This operation requires Developer role",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
