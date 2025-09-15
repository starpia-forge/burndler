package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/burndler/burndler/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestJWTAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		JWTSecret:   "test-secret-key",
		JWTIssuer:   "burndler",
		JWTAudience: "burndler-api",
	}

	tests := []struct {
		name           string
		authHeader     string
		setupToken     func() string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "UNAUTHORIZED",
		},
		{
			name:           "invalid header format - no Bearer",
			authHeader:     "token-without-bearer",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_TOKEN_FORMAT",
		},
		{
			name:           "invalid header format - wrong prefix",
			authHeader:     "Basic abc123",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_TOKEN_FORMAT",
		},
		{
			name: "valid token with Developer role",
			setupToken: func() string {
				claims := &Claims{
					UserID: "123",
					Email:  "dev@example.com",
					Role:   "Developer",
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer:    cfg.JWTIssuer,
						Audience:  []string{cfg.JWTAudience},
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(cfg.JWTSecret))
				return tokenString
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "valid token with Engineer role",
			setupToken: func() string {
				claims := &Claims{
					UserID: "456",
					Email:  "eng@example.com",
					Role:   "Engineer",
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer:    cfg.JWTIssuer,
						Audience:  []string{cfg.JWTAudience},
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(cfg.JWTSecret))
				return tokenString
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "expired token",
			setupToken: func() string {
				claims := &Claims{
					UserID: "123",
					Email:  "dev@example.com",
					Role:   "Developer",
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer:    cfg.JWTIssuer,
						Audience:  []string{cfg.JWTAudience},
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // Expired
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(cfg.JWTSecret))
				return tokenString
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_TOKEN",
		},
		{
			name: "wrong issuer",
			setupToken: func() string {
				claims := &Claims{
					UserID: "123",
					Email:  "dev@example.com",
					Role:   "Developer",
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer:    "wrong-issuer",
						Audience:  []string{cfg.JWTAudience},
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(cfg.JWTSecret))
				return tokenString
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_TOKEN_SCOPE",
		},
		{
			name: "wrong audience",
			setupToken: func() string {
				claims := &Claims{
					UserID: "123",
					Email:  "dev@example.com",
					Role:   "Developer",
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer:    cfg.JWTIssuer,
						Audience:  []string{"wrong-audience"},
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(cfg.JWTSecret))
				return tokenString
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_TOKEN_SCOPE",
		},
		{
			name: "invalid role",
			setupToken: func() string {
				claims := &Claims{
					UserID: "123",
					Email:  "user@example.com",
					Role:   "User", // Invalid role
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer:    cfg.JWTIssuer,
						Audience:  []string{cfg.JWTAudience},
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte(cfg.JWTSecret))
				return tokenString
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  "INVALID_ROLE",
		},
		{
			name:           "invalid token - wrong signature",
			authHeader:     "Bearer invalid.token.signature",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_TOKEN",
		},
		{
			name: "token signed with wrong secret",
			setupToken: func() string {
				claims := &Claims{
					UserID: "123",
					Email:  "dev@example.com",
					Role:   "Developer",
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer:    cfg.JWTIssuer,
						Audience:  []string{cfg.JWTAudience},
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, _ := token.SignedString([]byte("wrong-secret"))
				return tokenString
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_TOKEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(JWTAuth(cfg))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest(http.MethodGet, "/test", nil)

			// Set up authorization header
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			} else if tt.setupToken != nil {
				token := tt.setupToken()
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("status = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.expectedError != "" && w.Code != http.StatusOK {
				// Check that the error response contains the expected error code
				bodyStr := w.Body.String()
				if !strings.Contains(bodyStr, tt.expectedError) {
					t.Errorf("error response does not contain %v, got: %v", tt.expectedError, bodyStr)
				}
			}
		})
	}
}

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requiredRole   string
		contextRole    interface{}
		hasRole        bool
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing role in context",
			requiredRole:   "Developer",
			hasRole:        false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "NO_ROLE",
		},
		{
			name:           "invalid role type in context",
			requiredRole:   "Developer",
			contextRole:    123, // Invalid type
			hasRole:        true,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "INVALID_ROLE_TYPE",
		},
		{
			name:           "Developer accessing Developer-only endpoint",
			requiredRole:   "Developer",
			contextRole:    "Developer",
			hasRole:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Engineer accessing Developer-only endpoint",
			requiredRole:   "Developer",
			contextRole:    "Engineer",
			hasRole:        true,
			expectedStatus: http.StatusForbidden,
			expectedError:  "INSUFFICIENT_PERMISSIONS",
		},
		{
			name:           "Developer accessing Engineer endpoint",
			requiredRole:   "Engineer",
			contextRole:    "Developer",
			hasRole:        true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Engineer accessing Engineer endpoint",
			requiredRole:   "Engineer",
			contextRole:    "Engineer",
			hasRole:        true,
			expectedStatus: http.StatusOK,
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

			router.Use(RequireRole(tt.requiredRole))
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
				bodyStr := w.Body.String()
				if !strings.Contains(bodyStr, tt.expectedError) {
					t.Errorf("error response does not contain %v, got: %v", tt.expectedError, bodyStr)
				}
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "item exists in slice",
			slice:    []string{"a", "b", "c"},
			item:     "b",
			expected: true,
		},
		{
			name:     "item does not exist in slice",
			slice:    []string{"a", "b", "c"},
			item:     "d",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			item:     "a",
			expected: false,
		},
		{
			name:     "nil slice",
			slice:    nil,
			item:     "a",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			if result != tt.expected {
				t.Errorf("contains() = %v, want %v", result, tt.expected)
			}
		})
	}
}
