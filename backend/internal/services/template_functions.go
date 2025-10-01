package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"net"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
)

// GetExtendedTemplateFuncMap returns extended template functions
func GetExtendedTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		// String functions
		"upper":     strings.ToUpper,
		"lower":     strings.ToLower,
		"trim":      strings.TrimSpace,
		"replace":   strings.ReplaceAll,
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"split":     strings.Split,
		"join":      strings.Join,

		// Math functions
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"mul": func(a, b int) int { return a * b },
		"div": safeDiv,
		"mod": safeMod,

		// Conditional functions
		"default": func(defaultVal, val interface{}) interface{} {
			if val == nil || val == "" {
				return defaultVal
			}
			return val
		},
		"eq": func(a, b interface{}) bool { return a == b },
		"ne": func(a, b interface{}) bool { return a != b },

		// Utility functions
		"env":       safeGetEnv,
		"uuid":      generateUUID,
		"timestamp": getTimestamp,
		"now":       getNow,

		// Security functions
		"generatePassword": generateSecurePassword,
		"hash":             hashString,
		"base64encode":     base64Encode,
		"base64decode":     base64Decode,

		// Network functions
		"randomPort": generateRandomPort,
		"localIP":    getLocalIP,
	}
}

// safeDiv performs division with zero-check
func safeDiv(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return a / b, nil
}

// safeMod performs modulo with zero-check
func safeMod(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("modulo by zero")
	}
	return a % b, nil
}

// safeGetEnv gets environment variable with whitelist check
func safeGetEnv(key string) string {
	// Whitelist of allowed environment variables
	allowedKeys := map[string]bool{
		"HOME":     true,
		"USER":     true,
		"HOSTNAME": true,
		"PWD":      true,
		"PATH":     false, // Potentially sensitive
	}

	if allowed, exists := allowedKeys[key]; exists && allowed {
		return os.Getenv(key)
	}

	// Return empty string for non-whitelisted keys
	return ""
}

// generateUUID generates a new UUID
func generateUUID() string {
	return uuid.New().String()
}

// getTimestamp returns current Unix timestamp
func getTimestamp() int64 {
	return time.Now().Unix()
}

// getNow returns current time in RFC3339 format
func getNow() string {
	return time.Now().Format(time.RFC3339)
}

// generateSecurePassword generates a cryptographically secure random password
func generateSecurePassword(length int) (string, error) {
	if length < 1 {
		return "", fmt.Errorf("password length must be at least 1")
	}
	if length > 128 {
		return "", fmt.Errorf("password length must be at most 128")
	}

	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?"
	password := make([]byte, length)

	for i := range password {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		password[i] = charset[n.Int64()]
	}

	return string(password), nil
}

// hashString returns SHA256 hash of a string
func hashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

// base64Encode encodes a string to base64
func base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// base64Decode decodes a base64 string
func base64Decode(s string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", fmt.Errorf("base64 decode error: %w", err)
	}
	return string(decoded), nil
}

// generateRandomPort generates a random port in a given range
func generateRandomPort(min, max int) (int, error) {
	if min < 1 || min > 65535 {
		return 0, fmt.Errorf("min port must be between 1 and 65535")
	}
	if max < 1 || max > 65535 {
		return 0, fmt.Errorf("max port must be between 1 and 65535")
	}
	if min > max {
		return 0, fmt.Errorf("min port must be less than or equal to max port")
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return 0, fmt.Errorf("failed to generate random number: %w", err)
	}

	return min + int(n.Int64()), nil
}

// getLocalIP returns the local non-loopback IP address
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "127.0.0.1"
}
