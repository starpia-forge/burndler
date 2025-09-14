package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequirePermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		permission     Permission
		contextRole    interface{}
		hasRole        bool
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing role in context",
			permission:     PermissionRead,
			hasRole:        false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "NO_ROLE",
		},
		{
			name:           "invalid role type in context",
			permission:     PermissionRead,
			contextRole:    123, // Invalid type
			hasRole:        true,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "INVALID_ROLE_TYPE",
		},
		{
			name:           "unknown role",
			permission:     PermissionRead,
			contextRole:    "Admin", // Unknown role
			hasRole:        true,
			expectedStatus: http.StatusForbidden,
			expectedError:  "UNKNOWN_ROLE",
		},
		{
			name:           "Developer with read permission",
			permission:     PermissionRead,
			contextRole:    "Developer",
			hasRole:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Developer with write permission",
			permission:     PermissionWrite,
			contextRole:    "Developer",
			hasRole:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Developer with delete permission",
			permission:     PermissionDelete,
			contextRole:    "Developer",
			hasRole:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Developer with admin permission",
			permission:     PermissionAdmin,
			contextRole:    "Developer",
			hasRole:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Engineer with read permission",
			permission:     PermissionRead,
			contextRole:    "Engineer",
			hasRole:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Engineer without write permission",
			permission:     PermissionWrite,
			contextRole:    "Engineer",
			hasRole:        true,
			expectedStatus: http.StatusForbidden,
			expectedError:  "INSUFFICIENT_PERMISSIONS",
		},
		{
			name:           "Engineer without delete permission",
			permission:     PermissionDelete,
			contextRole:    "Engineer",
			hasRole:        true,
			expectedStatus: http.StatusForbidden,
			expectedError:  "INSUFFICIENT_PERMISSIONS",
		},
		{
			name:           "Engineer without admin permission",
			permission:     PermissionAdmin,
			contextRole:    "Engineer",
			hasRole:        true,
			expectedStatus: http.StatusForbidden,
			expectedError:  "INSUFFICIENT_PERMISSIONS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			// Set up context
			router.Use(func(c *gin.Context) {
				if tt.hasRole {
					c.Set("role", tt.contextRole)
				}
				c.Next()
			})

			router.Use(RequirePermission(tt.permission))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err == nil {
					if errorCode, ok := response["error"].(string); !ok || errorCode != tt.expectedError {
						t.Errorf("error = %v, want %v", response["error"], tt.expectedError)
					}

					// Check details for INSUFFICIENT_PERMISSIONS
					if tt.expectedError == "INSUFFICIENT_PERMISSIONS" {
						if details, ok := response["details"].(map[string]interface{}); ok {
							if perm, ok := details["required_permission"].(string); !ok || perm != string(tt.permission) {
								t.Errorf("required_permission = %v, want %v", perm, tt.permission)
							}
						}
					}
				}
			}
		})
	}
}

func TestGetUserRole(t *testing.T) {
	tests := []struct {
		name         string
		contextRole  interface{}
		hasRole      bool
		expectedRole RBACRoles
		expectedOk   bool
	}{
		{
			name:         "valid Developer role",
			contextRole:  "Developer",
			hasRole:      true,
			expectedRole: RoleDeveloper,
			expectedOk:   true,
		},
		{
			name:         "valid Engineer role",
			contextRole:  "Engineer",
			hasRole:      true,
			expectedRole: RoleEngineer,
			expectedOk:   true,
		},
		{
			name:         "missing role",
			hasRole:      false,
			expectedRole: "",
			expectedOk:   false,
		},
		{
			name:         "invalid role type",
			contextRole:  123,
			hasRole:      true,
			expectedRole: "",
			expectedOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.hasRole {
				c.Set("role", tt.contextRole)
			}

			role, ok := GetUserRole(c)
			if ok != tt.expectedOk {
				t.Errorf("GetUserRole() ok = %v, want %v", ok, tt.expectedOk)
			}
			if role != tt.expectedRole {
				t.Errorf("GetUserRole() role = %v, want %v", role, tt.expectedRole)
			}
		})
	}
}

func TestHasPermission(t *testing.T) {
	tests := []struct {
		name       string
		role       RBACRoles
		permission Permission
		expected   bool
	}{
		{
			name:       "Developer has read permission",
			role:       RoleDeveloper,
			permission: PermissionRead,
			expected:   true,
		},
		{
			name:       "Developer has write permission",
			role:       RoleDeveloper,
			permission: PermissionWrite,
			expected:   true,
		},
		{
			name:       "Developer has delete permission",
			role:       RoleDeveloper,
			permission: PermissionDelete,
			expected:   true,
		},
		{
			name:       "Developer has admin permission",
			role:       RoleDeveloper,
			permission: PermissionAdmin,
			expected:   true,
		},
		{
			name:       "Engineer has read permission",
			role:       RoleEngineer,
			permission: PermissionRead,
			expected:   true,
		},
		{
			name:       "Engineer does not have write permission",
			role:       RoleEngineer,
			permission: PermissionWrite,
			expected:   false,
		},
		{
			name:       "Engineer does not have delete permission",
			role:       RoleEngineer,
			permission: PermissionDelete,
			expected:   false,
		},
		{
			name:       "Engineer does not have admin permission",
			role:       RoleEngineer,
			permission: PermissionAdmin,
			expected:   false,
		},
		{
			name:       "unknown role has no permissions",
			role:       "Unknown",
			permission: PermissionRead,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasPermission(tt.role, tt.permission)
			if result != tt.expected {
				t.Errorf("HasPermission() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEnforceReadOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		contextRole    interface{}
		hasRole        bool
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing role",
			method:         http.MethodGet,
			hasRole:        false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "NO_ROLE",
		},
		{
			name:           "Developer can GET",
			method:         http.MethodGet,
			contextRole:    "Developer",
			hasRole:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Developer can POST",
			method:         http.MethodPost,
			contextRole:    "Developer",
			hasRole:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Developer can PUT",
			method:         http.MethodPut,
			contextRole:    "Developer",
			hasRole:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Developer can DELETE",
			method:         http.MethodDelete,
			contextRole:    "Developer",
			hasRole:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Engineer can GET",
			method:         http.MethodGet,
			contextRole:    "Engineer",
			hasRole:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Engineer cannot POST",
			method:         http.MethodPost,
			contextRole:    "Engineer",
			hasRole:        true,
			expectedStatus: http.StatusForbidden,
			expectedError:  "READ_ONLY_ACCESS",
		},
		{
			name:           "Engineer cannot PUT",
			method:         http.MethodPut,
			contextRole:    "Engineer",
			hasRole:        true,
			expectedStatus: http.StatusForbidden,
			expectedError:  "READ_ONLY_ACCESS",
		},
		{
			name:           "Engineer cannot DELETE",
			method:         http.MethodDelete,
			contextRole:    "Engineer",
			hasRole:        true,
			expectedStatus: http.StatusForbidden,
			expectedError:  "READ_ONLY_ACCESS",
		},
		{
			name:           "Engineer cannot PATCH",
			method:         http.MethodPatch,
			contextRole:    "Engineer",
			hasRole:        true,
			expectedStatus: http.StatusForbidden,
			expectedError:  "READ_ONLY_ACCESS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			// Set up context
			router.Use(func(c *gin.Context) {
				if tt.hasRole {
					c.Set("role", tt.contextRole)
				}
				c.Next()
			})

			router.Use(EnforceReadOnly())

			// Add handlers for all methods
			handler := func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			}
			router.GET("/test", handler)
			router.POST("/test", handler)
			router.PUT("/test", handler)
			router.DELETE("/test", handler)
			router.PATCH("/test", handler)

			req, _ := http.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err == nil {
					if errorCode, ok := response["error"].(string); !ok || errorCode != tt.expectedError {
						t.Errorf("error = %v, want %v", response["error"], tt.expectedError)
					}

					// Check details for READ_ONLY_ACCESS
					if tt.expectedError == "READ_ONLY_ACCESS" {
						if details, ok := response["details"].(map[string]interface{}); ok {
							if method, ok := details["method"].(string); !ok || method != tt.method {
								t.Errorf("method in details = %v, want %v", method, tt.method)
							}
							if role, ok := details["role"].(string); !ok || role != string(tt.contextRole.(string)) {
								t.Errorf("role in details = %v, want %v", role, tt.contextRole)
							}
						}
					}
				}
			}
		})
	}
}

func TestRolePermissions(t *testing.T) {
	// Test that RolePermissions map is correctly configured
	tests := []struct {
		name               string
		role               RBACRoles
		expectedPermCount  int
		shouldHaveRead     bool
		shouldHaveWrite    bool
		shouldHaveDelete   bool
		shouldHaveAdmin    bool
	}{
		{
			name:              "Developer permissions",
			role:              RoleDeveloper,
			expectedPermCount: 4,
			shouldHaveRead:    true,
			shouldHaveWrite:   true,
			shouldHaveDelete:  true,
			shouldHaveAdmin:   true,
		},
		{
			name:              "Engineer permissions",
			role:              RoleEngineer,
			expectedPermCount: 1,
			shouldHaveRead:    true,
			shouldHaveWrite:   false,
			shouldHaveDelete:  false,
			shouldHaveAdmin:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			perms, exists := RolePermissions[tt.role]
			if !exists {
				t.Fatalf("Role %v not found in RolePermissions", tt.role)
			}

			if len(perms) != tt.expectedPermCount {
				t.Errorf("Permission count = %v, want %v", len(perms), tt.expectedPermCount)
			}

			hasRead := HasPermission(tt.role, PermissionRead)
			if hasRead != tt.shouldHaveRead {
				t.Errorf("HasPermission(Read) = %v, want %v", hasRead, tt.shouldHaveRead)
			}

			hasWrite := HasPermission(tt.role, PermissionWrite)
			if hasWrite != tt.shouldHaveWrite {
				t.Errorf("HasPermission(Write) = %v, want %v", hasWrite, tt.shouldHaveWrite)
			}

			hasDelete := HasPermission(tt.role, PermissionDelete)
			if hasDelete != tt.shouldHaveDelete {
				t.Errorf("HasPermission(Delete) = %v, want %v", hasDelete, tt.shouldHaveDelete)
			}

			hasAdmin := HasPermission(tt.role, PermissionAdmin)
			if hasAdmin != tt.shouldHaveAdmin {
				t.Errorf("HasPermission(Admin) = %v, want %v", hasAdmin, tt.shouldHaveAdmin)
			}
		})
	}
}