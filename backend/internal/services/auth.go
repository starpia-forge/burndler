package services

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/burndler/burndler/internal/config"
	"github.com/burndler/burndler/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// Claims represents JWT claims with user role
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"` // Developer, Engineer, or Admin
	jwt.RegisteredClaims
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrInvalidToken       = errors.New("invalid token")
)

// AuthService handles authentication operations
type AuthService struct {
	config *config.Config
	db     *gorm.DB
}

// NewAuthService creates a new authentication service
func NewAuthService(cfg *config.Config, db *gorm.DB) *AuthService {
	return &AuthService{
		config: cfg,
		db:     db,
	}
}

// GenerateToken creates a JWT access token for the user
func (a *AuthService) GenerateToken(user *models.User) (string, error) {
	claims := &Claims{
		UserID: strconv.FormatUint(uint64(user.ID), 10),
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.config.JWTIssuer,
			Audience:  []string{a.config.JWTAudience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.config.JWTExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.config.JWTSecret))
}

// GenerateRefreshToken creates a JWT refresh token for the user
func (a *AuthService) GenerateRefreshToken(user *models.User) (string, error) {
	claims := &Claims{
		UserID: strconv.FormatUint(uint64(user.ID), 10),
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.config.JWTIssuer,
			Audience:  []string{a.config.JWTAudience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.config.JWTRefreshExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.config.JWTSecret))
}

// AuthenticateUser validates user credentials and returns the user if valid
func (a *AuthService) AuthenticateUser(email, password string) (*models.User, error) {
	if email == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	var user models.User
	err := a.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Check if user is active
	if !user.Active {
		return nil, ErrUserInactive
	}

	// Verify password
	if !user.CheckPassword(password) {
		return nil, ErrInvalidCredentials
	}

	return &user, nil
}

// ValidateToken parses and validates a JWT token, returning the claims
func (a *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token parsing error: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Validate issuer and audience
	if claims.Issuer != a.config.JWTIssuer {
		return nil, fmt.Errorf("invalid issuer: %s", claims.Issuer)
	}

	audienceValid := false
	for _, aud := range claims.Audience {
		if aud == a.config.JWTAudience {
			audienceValid = true
			break
		}
	}
	if !audienceValid {
		return nil, fmt.Errorf("invalid audience: %v", claims.Audience)
	}

	return claims, nil
}

// RefreshToken generates new access and refresh tokens from a valid refresh token
func (a *AuthService) RefreshToken(refreshTokenString string) (accessToken, newRefreshToken string, err error) {
	claims, err := a.ValidateToken(refreshTokenString)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Get user from database to ensure they still exist and are active
	userID, err := strconv.ParseUint(claims.UserID, 10, 32)
	if err != nil {
		return "", "", fmt.Errorf("invalid user ID in token: %w", err)
	}

	var user models.User
	err = a.db.First(&user, uint(userID)).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", ErrUserNotFound
		}
		return "", "", fmt.Errorf("database error: %w", err)
	}

	if !user.Active {
		return "", "", ErrUserInactive
	}

	// Generate new tokens
	accessToken, err = a.GenerateToken(&user)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err = a.GenerateRefreshToken(&user)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, newRefreshToken, nil
}
