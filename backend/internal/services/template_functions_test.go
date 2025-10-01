package services

import (
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetExtendedTemplateFuncMap(t *testing.T) {
	funcMap := GetExtendedTemplateFuncMap()
	assert.NotNil(t, funcMap)

	// Verify all expected functions are present
	expectedFunctions := []string{
		// String functions
		"upper", "lower", "trim", "replace", "contains", "hasPrefix", "hasSuffix", "split", "join",
		// Math functions
		"add", "sub", "mul", "div", "mod",
		// Conditional functions
		"default", "eq", "ne",
		// Utility functions
		"env", "uuid", "timestamp", "now",
		// Security functions
		"generatePassword", "hash", "base64encode", "base64decode",
		// Network functions
		"randomPort", "localIP",
	}

	for _, funcName := range expectedFunctions {
		assert.Contains(t, funcMap, funcName, "Function %s should be present", funcName)
	}
}

// Test Math Functions
func TestSafeDiv(t *testing.T) {
	tests := []struct {
		name        string
		a, b        int
		expected    int
		expectError bool
	}{
		{"normal division", 10, 2, 5, false},
		{"division by zero", 10, 0, 0, true},
		{"negative numbers", -10, 2, -5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := safeDiv(tt.a, tt.b)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSafeMod(t *testing.T) {
	tests := []struct {
		name        string
		a, b        int
		expected    int
		expectError bool
	}{
		{"normal modulo", 10, 3, 1, false},
		{"modulo by zero", 10, 0, 0, true},
		{"negative numbers", -10, 3, -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := safeMod(tt.a, tt.b)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Test Security Functions
func TestGenerateSecurePassword(t *testing.T) {
	tests := []struct {
		name        string
		length      int
		expectError bool
	}{
		{"valid length 8", 8, false},
		{"valid length 16", 16, false},
		{"valid length 32", 32, false},
		{"minimum length 1", 1, false},
		{"maximum length 128", 128, false},
		{"zero length", 0, true},
		{"negative length", -1, true},
		{"too long", 129, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := generateSecurePassword(tt.length)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.length, len(password))
				// Verify password contains only allowed characters
				const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?"
				for _, c := range password {
					assert.Contains(t, charset, string(c))
				}
			}
		})
	}
}

func TestGenerateSecurePassword_Randomness(t *testing.T) {
	// Generate multiple passwords and ensure they're different
	passwords := make(map[string]bool)
	for i := 0; i < 10; i++ {
		password, err := generateSecurePassword(16)
		require.NoError(t, err)
		passwords[password] = true
	}
	// All passwords should be unique
	assert.Equal(t, 10, len(passwords), "Generated passwords should be unique")
}

func TestHashString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"empty string",
			"",
			"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			"hello world",
			"hello world",
			"b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hashString(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, 64, len(result), "SHA256 hash should be 64 hex characters")
		})
	}
}

func TestBase64Encode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "aGVsbG8="},
		{"", ""},
		{"test123", "dGVzdDEyMw=="},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := base64Encode(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBase64Decode(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{"valid base64", "aGVsbG8=", "hello", false},
		{"empty string", "", "", false},
		{"valid base64 2", "dGVzdDEyMw==", "test123", false},
		{"invalid base64", "not-base64!!!", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := base64Decode(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Test Network Functions
func TestGenerateRandomPort(t *testing.T) {
	tests := []struct {
		name        string
		min, max    int
		expectError bool
	}{
		{"valid range", 8000, 9000, false},
		{"single port", 8080, 8080, false},
		{"full range", 1, 65535, false},
		{"invalid min", 0, 1000, true},
		{"invalid max", 1000, 70000, true},
		{"min > max", 9000, 8000, true},
		{"negative min", -1, 1000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port, err := generateRandomPort(tt.min, tt.max)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, port, tt.min)
				assert.LessOrEqual(t, port, tt.max)
			}
		})
	}
}

func TestGenerateRandomPort_Distribution(t *testing.T) {
	// Generate multiple ports and ensure they vary
	ports := make(map[int]bool)
	min, max := 8000, 8100
	for i := 0; i < 50; i++ {
		port, err := generateRandomPort(min, max)
		require.NoError(t, err)
		ports[port] = true
	}
	// Should have generated at least 10 different ports
	assert.GreaterOrEqual(t, len(ports), 10, "Ports should vary")
}

func TestGetLocalIP(t *testing.T) {
	ip := getLocalIP()
	assert.NotEmpty(t, ip)
	// Should return a valid IP address format
	parts := strings.Split(ip, ".")
	assert.Equal(t, 4, len(parts), "Should be IPv4 format")
}

// Test Utility Functions
func TestGenerateUUID(t *testing.T) {
	uuid1 := generateUUID()
	uuid2 := generateUUID()

	// UUIDs should be 36 characters (with hyphens)
	assert.Equal(t, 36, len(uuid1))
	assert.Equal(t, 36, len(uuid2))

	// UUIDs should be different
	assert.NotEqual(t, uuid1, uuid2)

	// UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	parts := strings.Split(uuid1, "-")
	assert.Equal(t, 5, len(parts))
	assert.Equal(t, 8, len(parts[0]))
	assert.Equal(t, 4, len(parts[1]))
	assert.Equal(t, 4, len(parts[2]))
	assert.Equal(t, 4, len(parts[3]))
	assert.Equal(t, 12, len(parts[4]))
}

func TestGetTimestamp(t *testing.T) {
	ts := getTimestamp()
	assert.Greater(t, ts, int64(0))
	// Timestamp should be reasonable (after 2020)
	assert.Greater(t, ts, int64(1577836800)) // Jan 1, 2020
}

func TestGetNow(t *testing.T) {
	now := getNow()
	assert.NotEmpty(t, now)
	// Should be RFC3339 format (e.g., "2006-01-02T15:04:05Z07:00")
	assert.Contains(t, now, "T")
	assert.True(t, len(now) >= 20)
}

func TestSafeGetEnv(t *testing.T) {
	// Test whitelisted variable
	result := safeGetEnv("HOME")
	// HOME might be set or not, but shouldn't error
	assert.NotNil(t, result)

	// Test non-whitelisted variable
	result = safeGetEnv("SECRET_KEY")
	assert.Equal(t, "", result, "Non-whitelisted env var should return empty string")

	// Test blacklisted variable
	result = safeGetEnv("PATH")
	assert.Equal(t, "", result, "PATH should be blocked for security")
}

// Test Template Integration
func TestTemplateFunctionsInTemplate(t *testing.T) {
	funcMap := GetExtendedTemplateFuncMap()

	t.Run("string functions", func(t *testing.T) {
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(
			`{{ .name | upper }} - {{ .name | lower }}`,
		))

		var buf strings.Builder
		err := tmpl.Execute(&buf, map[string]interface{}{"name": "Test"})
		require.NoError(t, err)
		assert.Equal(t, "TEST - test", buf.String())
	})

	t.Run("contains function", func(t *testing.T) {
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(
			`{{ if contains .text "hello" }}found{{ else }}not found{{ end }}`,
		))

		var buf strings.Builder
		err := tmpl.Execute(&buf, map[string]interface{}{"text": "hello world"})
		require.NoError(t, err)
		assert.Equal(t, "found", buf.String())
	})

	t.Run("math functions", func(t *testing.T) {
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(
			`{{ add .a .b }} - {{ sub .a .b }} - {{ mul .a .b }}`,
		))

		var buf strings.Builder
		err := tmpl.Execute(&buf, map[string]interface{}{"a": 10, "b": 5})
		require.NoError(t, err)
		assert.Equal(t, "15 - 5 - 50", buf.String())
	})

	t.Run("uuid function", func(t *testing.T) {
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(
			`{{ uuid }}`,
		))

		var buf strings.Builder
		err := tmpl.Execute(&buf, nil)
		require.NoError(t, err)
		assert.Equal(t, 36, len(buf.String()))
	})

	t.Run("hash function", func(t *testing.T) {
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(
			`{{ hash "test" }}`,
		))

		var buf strings.Builder
		err := tmpl.Execute(&buf, nil)
		require.NoError(t, err)
		assert.Equal(t, 64, len(buf.String()))
	})

	t.Run("base64 functions", func(t *testing.T) {
		tmpl := template.Must(template.New("test").Funcs(funcMap).Parse(
			`{{ base64encode "hello" }}`,
		))

		var buf strings.Builder
		err := tmpl.Execute(&buf, nil)
		require.NoError(t, err)
		assert.Equal(t, "aGVsbG8=", buf.String())
	})
}
