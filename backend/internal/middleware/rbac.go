package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RBACRoles defines the available roles
type RBACRoles string

const (
	RoleDeveloper RBACRoles = "Developer" // Read/Write access
	RoleEngineer  RBACRoles = "Engineer"  // Read-only access
)

// Permission defines what operations are allowed
type Permission string

const (
	PermissionRead   Permission = "read"
	PermissionWrite  Permission = "write"
	PermissionDelete Permission = "delete"
	PermissionAdmin  Permission = "admin"
)

// RolePermissions maps roles to their permissions
var RolePermissions = map[RBACRoles][]Permission{
	RoleDeveloper: {
		PermissionRead,
		PermissionWrite,
		PermissionDelete,
		PermissionAdmin,
	},
	RoleEngineer: {
		PermissionRead,
	},
}

// RequirePermission checks if the user has the required permission
func RequirePermission(permission Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleInterface, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "NO_ROLE",
				"message": "User role not found",
			})
			c.Abort()
			return
		}

		roleStr, ok := roleInterface.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "INVALID_ROLE_TYPE",
				"message": "Invalid role type in context",
			})
			c.Abort()
			return
		}

		role := RBACRoles(roleStr)
		permissions, exists := RolePermissions[role]
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "UNKNOWN_ROLE",
				"message": "Unknown user role",
			})
			c.Abort()
			return
		}

		// Check if role has required permission
		hasPermission := false
		for _, p := range permissions {
			if p == permission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "INSUFFICIENT_PERMISSIONS",
				"message": "You don't have permission to perform this action",
				"details": gin.H{
					"required_permission": permission,
					"user_role":           role,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserRole returns the current user's role from context
func GetUserRole(c *gin.Context) (RBACRoles, bool) {
	roleInterface, exists := c.Get("role")
	if !exists {
		return "", false
	}

	roleStr, ok := roleInterface.(string)
	if !ok {
		return "", false
	}

	return RBACRoles(roleStr), true
}

// HasPermission checks if a role has a specific permission
func HasPermission(role RBACRoles, permission Permission) bool {
	permissions, exists := RolePermissions[role]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}

	return false
}

// EnforceReadOnly ensures Engineers can only perform read operations
func EnforceReadOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := GetUserRole(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "NO_ROLE",
				"message": "User role not found",
			})
			c.Abort()
			return
		}

		// Engineers can only perform GET requests
		if role == RoleEngineer && c.Request.Method != http.MethodGet {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "READ_ONLY_ACCESS",
				"message": "Engineer role has read-only access",
				"details": gin.H{
					"method": c.Request.Method,
					"role":   role,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
