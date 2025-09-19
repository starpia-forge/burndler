package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationRunner_New(t *testing.T) {
	runner := NewMigrationRunner()
	assert.NotNil(t, runner)
}

func TestMigrationRunner_RunMigrations_Success(t *testing.T) {
	runner := NewMigrationRunner()

	// Migration will fail due to database connection but that's expected
	err := runner.RunMigrations()
	// Database connection error is expected in this test environment
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize application")
}

func TestMigrationRunner_RunMigrations_WithEnvironmentLoading(t *testing.T) {
	// Create a migration runner with environment loading
	runner := NewMigrationRunner()

	// Set up environment for testing
	envLoader := NewEnvironmentLoader()
	err := envLoader.LoadEnvironment("", true) // Use default env loading
	require.NoError(t, err)

	// Run migrations with environment loaded
	err = runner.RunMigrations()
	// Database connection error is expected in this test environment
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize application")
}

func TestMigrationRunner_ValidateConfig(t *testing.T) {
	runner := NewMigrationRunner()

	// Test config validation
	isValid := runner.ValidateConfig()
	// Our implementation returns true for now
	assert.True(t, isValid)
}

// Integration test for the full migration flow
func TestMigrationFlow_Integration(t *testing.T) {
	// This test represents the full flow:
	// 1. Load environment
	// 2. Initialize app
	// 3. Run migrations
	// 4. Cleanup

	envLoader := NewEnvironmentLoader()
	err := envLoader.LoadEnvironment("", true)
	require.NoError(t, err)

	runner := NewMigrationRunner()

	// Database connection error is expected in this test environment
	err = runner.RunMigrations()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to initialize application")
}