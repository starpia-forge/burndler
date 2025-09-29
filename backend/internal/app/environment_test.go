package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEnvironmentLoader(t *testing.T) {
	loader := NewEnvironmentLoader()
	assert.NotNil(t, loader)
}

func TestEnvironmentLoader_LoadEnvironment_DevMode(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()

	// Create test .env file
	envFile := filepath.Join(tempDir, ".env.test")
	envContent := "TEST_VAR=test_value\nANOTHER_VAR=another_value\n"
	err := os.WriteFile(envFile, []byte(envContent), 0644)
	require.NoError(t, err)

	loader := NewEnvironmentLoader()

	// Test loading in development mode
	err = loader.LoadEnvironment(envFile, true)
	assert.NoError(t, err)

	// Verify environment variables were set
	assert.Equal(t, "test_value", os.Getenv("TEST_VAR"))
	assert.Equal(t, "another_value", os.Getenv("ANOTHER_VAR"))

	// Cleanup
	_ = os.Unsetenv("TEST_VAR")
	_ = os.Unsetenv("ANOTHER_VAR")
}

func TestEnvironmentLoader_LoadEnvironment_ProductionMode(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()

	// Create test .env file
	envFile := filepath.Join(tempDir, ".env.test")
	envContent := "PROD_VAR=prod_value\n"
	err := os.WriteFile(envFile, []byte(envContent), 0644)
	require.NoError(t, err)

	loader := NewEnvironmentLoader()

	// Test loading in production mode (should not load file)
	err = loader.LoadEnvironment(envFile, false)
	assert.NoError(t, err)

	// Verify environment variable was NOT set (production mode ignores env files)
	assert.Empty(t, os.Getenv("PROD_VAR"))
}

func TestEnvironmentLoader_LoadEnvironment_FileNotFound(t *testing.T) {
	loader := NewEnvironmentLoader()

	// Test loading non-existent file in dev mode
	err := loader.LoadEnvironment("/nonexistent/file.env", true)
	// Should not return error but should log warning
	assert.NoError(t, err)
}

func TestEnvironmentLoader_LoadEnvironment_EmptyEnvFile(t *testing.T) {
	loader := NewEnvironmentLoader()

	// Test with empty env file parameter in dev mode
	err := loader.LoadEnvironment("", true)
	assert.NoError(t, err)
}

func TestEnvironmentLoader_LoadEnvironment_DefaultFallback(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()

	// Change to temp directory
	_ = os.Chdir(tempDir)
	defer func() { _ = os.Chdir(originalDir) }()

	// Create .env.development file
	envDevContent := "DEV_VAR=dev_value\n"
	err := os.WriteFile(".env.development", []byte(envDevContent), 0644)
	require.NoError(t, err)

	// Create .env file as fallback
	envContent := "FALLBACK_VAR=fallback_value\n"
	err = os.WriteFile(".env", []byte(envContent), 0644)
	require.NoError(t, err)

	loader := NewEnvironmentLoader()

	// Test loading with empty env file (should use default logic)
	err = loader.LoadEnvironment("", true)
	assert.NoError(t, err)

	// Should load .env.development first
	assert.Equal(t, "dev_value", os.Getenv("DEV_VAR"))

	// Cleanup
	_ = os.Unsetenv("DEV_VAR")
	_ = os.Unsetenv("FALLBACK_VAR")
}

func TestEnvironmentLoader_LoadEnvironment_FallbackToEnv(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()

	// Change to temp directory
	_ = os.Chdir(tempDir)
	defer func() { _ = os.Chdir(originalDir) }()

	// Create only .env file (no .env.development)
	envContent := "FALLBACK_VAR=fallback_value\n"
	err := os.WriteFile(".env", []byte(envContent), 0644)
	require.NoError(t, err)

	loader := NewEnvironmentLoader()

	// Test loading with empty env file (should fallback to .env)
	err = loader.LoadEnvironment("", true)
	assert.NoError(t, err)

	// Should load .env as fallback
	assert.Equal(t, "fallback_value", os.Getenv("FALLBACK_VAR"))

	// Cleanup
	_ = os.Unsetenv("FALLBACK_VAR")
}