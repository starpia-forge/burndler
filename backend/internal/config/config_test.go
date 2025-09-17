package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Clear environment variables to test actual defaults
	_ = os.Unsetenv("DB_NAME")

	// Test default values
	cfg := Load()

	// Database defaults
	if cfg.DBHost != "localhost" {
		t.Errorf("DBHost = %v, want %v", cfg.DBHost, "localhost")
	}
	if cfg.DBPort != "5432" {
		t.Errorf("DBPort = %v, want %v", cfg.DBPort, "5432")
	}
	if cfg.DBName != "burndler" {
		t.Errorf("DBName = %v, want %v", cfg.DBName, "burndler")
	}
	if cfg.DBMaxConnections != 25 {
		t.Errorf("DBMaxConnections = %v, want %v", cfg.DBMaxConnections, 25)
	}

	// Storage defaults
	if cfg.StorageMode != "local" {
		t.Errorf("StorageMode = %v, want %v", cfg.StorageMode, "local")
	}
	if cfg.S3Region != "us-east-1" {
		t.Errorf("S3Region = %v, want %v", cfg.S3Region, "us-east-1")
	}
	if !cfg.S3UseSSL {
		t.Errorf("S3UseSSL = %v, want %v", cfg.S3UseSSL, true)
	}

	// Server defaults
	if cfg.ServerPort != "8080" {
		t.Errorf("ServerPort = %v, want %v", cfg.ServerPort, "8080")
	}
	if cfg.ServerMaxRequestSize != 100*1024*1024 {
		t.Errorf("ServerMaxRequestSize = %v, want %v", cfg.ServerMaxRequestSize, 100*1024*1024)
	}

	// Build worker defaults
	if cfg.BuildWorkerCount != 4 {
		t.Errorf("BuildWorkerCount = %v, want %v", cfg.BuildWorkerCount, 4)
	}
	if cfg.BuildRetentionDays != 7 {
		t.Errorf("BuildRetentionDays = %v, want %v", cfg.BuildRetentionDays, 7)
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	// Set environment variables
	if err := os.Setenv("DB_HOST", "db.example.com"); err != nil {
		t.Fatalf("Failed to set DB_HOST: %v", err)
	}
	if err := os.Setenv("DB_PORT", "5433"); err != nil {
		t.Fatalf("Failed to set DB_PORT: %v", err)
	}
	if err := os.Setenv("DB_MAX_CONNECTIONS", "50"); err != nil {
		t.Fatalf("Failed to set DB_MAX_CONNECTIONS: %v", err)
	}
	if err := os.Setenv("STORAGE_MODE", "s3"); err != nil {
		t.Fatalf("Failed to set STORAGE_MODE: %v", err)
	}
	if err := os.Setenv("S3_BUCKET", "test-bucket"); err != nil {
		t.Fatalf("Failed to set S3_BUCKET: %v", err)
	}
	if err := os.Setenv("S3_USE_SSL", "false"); err != nil {
		t.Fatalf("Failed to set S3_USE_SSL: %v", err)
	}
	if err := os.Setenv("SERVER_PORT", "9090"); err != nil {
		t.Fatalf("Failed to set SERVER_PORT: %v", err)
	}
	if err := os.Setenv("SERVER_MAX_REQUEST_SIZE", "200000000"); err != nil {
		t.Fatalf("Failed to set SERVER_MAX_REQUEST_SIZE: %v", err)
	}
	if err := os.Setenv("BUILD_WORKER_COUNT", "8"); err != nil {
		t.Fatalf("Failed to set BUILD_WORKER_COUNT: %v", err)
	}
	if err := os.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:4000"); err != nil {
		t.Fatalf("Failed to set CORS_ALLOWED_ORIGINS: %v", err)
	}

	defer func() {
		// Clean up environment variables
		if err := os.Unsetenv("DB_HOST"); err != nil {
			t.Logf("Warning: failed to unset DB_HOST: %v", err)
		}
		if err := os.Unsetenv("DB_PORT"); err != nil {
			t.Logf("Warning: failed to unset DB_PORT: %v", err)
		}
		if err := os.Unsetenv("DB_MAX_CONNECTIONS"); err != nil {
			t.Logf("Warning: failed to unset DB_MAX_CONNECTIONS: %v", err)
		}
		if err := os.Unsetenv("STORAGE_MODE"); err != nil {
			t.Logf("Warning: failed to unset STORAGE_MODE: %v", err)
		}
		if err := os.Unsetenv("S3_BUCKET"); err != nil {
			t.Logf("Warning: failed to unset S3_BUCKET: %v", err)
		}
		if err := os.Unsetenv("S3_USE_SSL"); err != nil {
			t.Logf("Warning: failed to unset S3_USE_SSL: %v", err)
		}
		if err := os.Unsetenv("SERVER_PORT"); err != nil {
			t.Logf("Warning: failed to unset SERVER_PORT: %v", err)
		}
		if err := os.Unsetenv("SERVER_MAX_REQUEST_SIZE"); err != nil {
			t.Logf("Warning: failed to unset SERVER_MAX_REQUEST_SIZE: %v", err)
		}
		if err := os.Unsetenv("BUILD_WORKER_COUNT"); err != nil {
			t.Logf("Warning: failed to unset BUILD_WORKER_COUNT: %v", err)
		}
		if err := os.Unsetenv("CORS_ALLOWED_ORIGINS"); err != nil {
			t.Logf("Warning: failed to unset CORS_ALLOWED_ORIGINS: %v", err)
		}
	}()

	cfg := Load()

	// Test overridden values
	if cfg.DBHost != "db.example.com" {
		t.Errorf("DBHost = %v, want %v", cfg.DBHost, "db.example.com")
	}
	if cfg.DBPort != "5433" {
		t.Errorf("DBPort = %v, want %v", cfg.DBPort, "5433")
	}
	if cfg.DBMaxConnections != 50 {
		t.Errorf("DBMaxConnections = %v, want %v", cfg.DBMaxConnections, 50)
	}
	if cfg.StorageMode != "s3" {
		t.Errorf("StorageMode = %v, want %v", cfg.StorageMode, "s3")
	}
	if cfg.S3Bucket != "test-bucket" {
		t.Errorf("S3Bucket = %v, want %v", cfg.S3Bucket, "test-bucket")
	}
	if cfg.S3UseSSL != false {
		t.Errorf("S3UseSSL = %v, want %v", cfg.S3UseSSL, false)
	}
	if cfg.ServerPort != "9090" {
		t.Errorf("ServerPort = %v, want %v", cfg.ServerPort, "9090")
	}
	if cfg.ServerMaxRequestSize != 200000000 {
		t.Errorf("ServerMaxRequestSize = %v, want %v", cfg.ServerMaxRequestSize, 200000000)
	}
	if cfg.BuildWorkerCount != 8 {
		t.Errorf("BuildWorkerCount = %v, want %v", cfg.BuildWorkerCount, 8)
	}

	// Test slice parsing
	if len(cfg.CORSAllowedOrigins) != 2 {
		t.Errorf("CORSAllowedOrigins length = %v, want %v", len(cfg.CORSAllowedOrigins), 2)
	}
	if cfg.CORSAllowedOrigins[0] != "http://localhost:3000" {
		t.Errorf("CORSAllowedOrigins[0] = %v, want %v", cfg.CORSAllowedOrigins[0], "http://localhost:3000")
	}
	if cfg.CORSAllowedOrigins[1] != "http://localhost:4000" {
		t.Errorf("CORSAllowedOrigins[1] = %v, want %v", cfg.CORSAllowedOrigins[1], "http://localhost:4000")
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_ENV_VAR",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "returns env value when set",
			key:          "TEST_ENV_VAR",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				if err := os.Setenv(tt.key, tt.envValue); err != nil {
					t.Fatalf("Failed to set env var %s: %v", tt.key, err)
				}
				defer func() {
					if err := os.Unsetenv(tt.key); err != nil {
						t.Logf("Warning: failed to unset %s: %v", tt.key, err)
					}
				}()
			}

			result := getEnv(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnv() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		expected     int
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_INT_VAR",
			defaultValue: 10,
			envValue:     "",
			expected:     10,
		},
		{
			name:         "returns parsed int when valid",
			key:          "TEST_INT_VAR",
			defaultValue: 10,
			envValue:     "25",
			expected:     25,
		},
		{
			name:         "returns default when invalid int",
			key:          "TEST_INT_VAR",
			defaultValue: 10,
			envValue:     "not-a-number",
			expected:     10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				if err := os.Setenv(tt.key, tt.envValue); err != nil {
					t.Fatalf("Failed to set env var %s: %v", tt.key, err)
				}
				defer func() {
					if err := os.Unsetenv(tt.key); err != nil {
						t.Logf("Warning: failed to unset %s: %v", tt.key, err)
					}
				}()
			}

			result := getEnvAsInt(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvAsInt() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetEnvAsInt64(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int64
		envValue     string
		expected     int64
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_INT64_VAR",
			defaultValue: 1000000,
			envValue:     "",
			expected:     1000000,
		},
		{
			name:         "returns parsed int64 when valid",
			key:          "TEST_INT64_VAR",
			defaultValue: 1000000,
			envValue:     "9999999999",
			expected:     9999999999,
		},
		{
			name:         "returns default when invalid int64",
			key:          "TEST_INT64_VAR",
			defaultValue: 1000000,
			envValue:     "invalid",
			expected:     1000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				if err := os.Setenv(tt.key, tt.envValue); err != nil {
					t.Fatalf("Failed to set env var %s: %v", tt.key, err)
				}
				defer func() {
					if err := os.Unsetenv(tt.key); err != nil {
						t.Logf("Warning: failed to unset %s: %v", tt.key, err)
					}
				}()
			}

			result := getEnvAsInt64(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvAsInt64() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetEnvAsBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		expected     bool
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_BOOL_VAR",
			defaultValue: true,
			envValue:     "",
			expected:     true,
		},
		{
			name:         "returns true for 'true'",
			key:          "TEST_BOOL_VAR",
			defaultValue: false,
			envValue:     "true",
			expected:     true,
		},
		{
			name:         "returns false for 'false'",
			key:          "TEST_BOOL_VAR",
			defaultValue: true,
			envValue:     "false",
			expected:     false,
		},
		{
			name:         "returns true for '1'",
			key:          "TEST_BOOL_VAR",
			defaultValue: false,
			envValue:     "1",
			expected:     true,
		},
		{
			name:         "returns false for '0'",
			key:          "TEST_BOOL_VAR",
			defaultValue: true,
			envValue:     "0",
			expected:     false,
		},
		{
			name:         "returns default for invalid bool",
			key:          "TEST_BOOL_VAR",
			defaultValue: true,
			envValue:     "not-a-bool",
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				if err := os.Setenv(tt.key, tt.envValue); err != nil {
					t.Fatalf("Failed to set env var %s: %v", tt.key, err)
				}
				defer func() {
					if err := os.Unsetenv(tt.key); err != nil {
						t.Logf("Warning: failed to unset %s: %v", tt.key, err)
					}
				}()
			}

			result := getEnvAsBool(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvAsBool() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetEnvAsDuration(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     time.Duration
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_DURATION_VAR",
			defaultValue: "10s",
			envValue:     "",
			expected:     10 * time.Second,
		},
		{
			name:         "returns parsed duration for seconds",
			key:          "TEST_DURATION_VAR",
			defaultValue: "10s",
			envValue:     "30s",
			expected:     30 * time.Second,
		},
		{
			name:         "returns parsed duration for minutes",
			key:          "TEST_DURATION_VAR",
			defaultValue: "10s",
			envValue:     "5m",
			expected:     5 * time.Minute,
		},
		{
			name:         "returns parsed duration for hours",
			key:          "TEST_DURATION_VAR",
			defaultValue: "10s",
			envValue:     "2h",
			expected:     2 * time.Hour,
		},
		{
			name:         "returns default for invalid duration",
			key:          "TEST_DURATION_VAR",
			defaultValue: "10s",
			envValue:     "invalid",
			expected:     10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				if err := os.Setenv(tt.key, tt.envValue); err != nil {
					t.Fatalf("Failed to set env var %s: %v", tt.key, err)
				}
				defer func() {
					if err := os.Unsetenv(tt.key); err != nil {
						t.Logf("Warning: failed to unset %s: %v", tt.key, err)
					}
				}()
			}

			result := getEnvAsDuration(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvAsDuration() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetEnvAsSlice(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue []string
		envValue     string
		expected     []string
	}{
		{
			name:         "returns default when env not set",
			key:          "TEST_SLICE_VAR",
			defaultValue: []string{"a", "b"},
			envValue:     "",
			expected:     []string{"a", "b"},
		},
		{
			name:         "returns single value",
			key:          "TEST_SLICE_VAR",
			defaultValue: []string{"a", "b"},
			envValue:     "single",
			expected:     []string{"single"},
		},
		{
			name:         "returns multiple values",
			key:          "TEST_SLICE_VAR",
			defaultValue: []string{"a", "b"},
			envValue:     "one,two,three",
			expected:     []string{"one", "two", "three"},
		},
		{
			name:         "handles URLs with commas",
			key:          "TEST_SLICE_VAR",
			defaultValue: []string{},
			envValue:     "http://localhost:3000,http://localhost:4000",
			expected:     []string{"http://localhost:3000", "http://localhost:4000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				if err := os.Setenv(tt.key, tt.envValue); err != nil {
					t.Fatalf("Failed to set env var %s: %v", tt.key, err)
				}
				defer func() {
					if err := os.Unsetenv(tt.key); err != nil {
						t.Logf("Warning: failed to unset %s: %v", tt.key, err)
					}
				}()
			}

			result := getEnvAsSlice(tt.key, tt.defaultValue)
			if len(result) != len(tt.expected) {
				t.Errorf("getEnvAsSlice() length = %v, want %v", len(result), len(tt.expected))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("getEnvAsSlice()[%d] = %v, want %v", i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestDurationValues(t *testing.T) {
	// Test duration parsing with actual config
	if err := os.Setenv("JWT_EXPIRATION", "48h"); err != nil {
		t.Fatalf("Failed to set JWT_EXPIRATION: %v", err)
	}
	if err := os.Setenv("SERVER_READ_TIMEOUT", "1m30s"); err != nil {
		t.Fatalf("Failed to set SERVER_READ_TIMEOUT: %v", err)
	}
	if err := os.Setenv("BUILD_TIMEOUT", "45m"); err != nil {
		t.Fatalf("Failed to set BUILD_TIMEOUT: %v", err)
	}

	defer func() {
		if err := os.Unsetenv("JWT_EXPIRATION"); err != nil {
			t.Logf("Warning: failed to unset JWT_EXPIRATION: %v", err)
		}
		if err := os.Unsetenv("SERVER_READ_TIMEOUT"); err != nil {
			t.Logf("Warning: failed to unset SERVER_READ_TIMEOUT: %v", err)
		}
		if err := os.Unsetenv("BUILD_TIMEOUT"); err != nil {
			t.Logf("Warning: failed to unset BUILD_TIMEOUT: %v", err)
		}
	}()

	cfg := Load()

	if cfg.JWTExpiration != 48*time.Hour {
		t.Errorf("JWTExpiration = %v, want %v", cfg.JWTExpiration, 48*time.Hour)
	}
	if cfg.ServerReadTimeout != 90*time.Second {
		t.Errorf("ServerReadTimeout = %v, want %v", cfg.ServerReadTimeout, 90*time.Second)
	}
	if cfg.BuildTimeout != 45*time.Minute {
		t.Errorf("BuildTimeout = %v, want %v", cfg.BuildTimeout, 45*time.Minute)
	}
}

func TestInvalidTypeConversions(t *testing.T) {
	// Test that invalid values fall back to defaults
	if err := os.Setenv("DB_MAX_CONNECTIONS", "not-a-number"); err != nil {
		t.Fatalf("Failed to set DB_MAX_CONNECTIONS: %v", err)
	}
	if err := os.Setenv("S3_USE_SSL", "maybe"); err != nil {
		t.Fatalf("Failed to set S3_USE_SSL: %v", err)
	}
	if err := os.Setenv("SERVER_MAX_REQUEST_SIZE", "way-too-big"); err != nil {
		t.Fatalf("Failed to set SERVER_MAX_REQUEST_SIZE: %v", err)
	}
	if err := os.Setenv("BUILD_TIMEOUT", "forever"); err != nil {
		t.Fatalf("Failed to set BUILD_TIMEOUT: %v", err)
	}

	defer func() {
		if err := os.Unsetenv("DB_MAX_CONNECTIONS"); err != nil {
			t.Logf("Warning: failed to unset DB_MAX_CONNECTIONS: %v", err)
		}
		if err := os.Unsetenv("S3_USE_SSL"); err != nil {
			t.Logf("Warning: failed to unset S3_USE_SSL: %v", err)
		}
		if err := os.Unsetenv("SERVER_MAX_REQUEST_SIZE"); err != nil {
			t.Logf("Warning: failed to unset SERVER_MAX_REQUEST_SIZE: %v", err)
		}
		if err := os.Unsetenv("BUILD_TIMEOUT"); err != nil {
			t.Logf("Warning: failed to unset BUILD_TIMEOUT: %v", err)
		}
	}()

	cfg := Load()

	// Should fall back to defaults when parsing fails
	if cfg.DBMaxConnections != 25 {
		t.Errorf("DBMaxConnections = %v, want %v (default)", cfg.DBMaxConnections, 25)
	}
	if cfg.S3UseSSL != true {
		t.Errorf("S3UseSSL = %v, want %v (default)", cfg.S3UseSSL, true)
	}
	if cfg.ServerMaxRequestSize != 100*1024*1024 {
		t.Errorf("ServerMaxRequestSize = %v, want %v (default)", cfg.ServerMaxRequestSize, 100*1024*1024)
	}
	if cfg.BuildTimeout != 30*time.Minute {
		t.Errorf("BuildTimeout = %v, want %v (default)", cfg.BuildTimeout, 30*time.Minute)
	}
}
