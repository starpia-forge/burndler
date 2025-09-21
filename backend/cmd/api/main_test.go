package main

import (
	"testing"

	"github.com/burndler/burndler/internal/app"
	"github.com/stretchr/testify/assert"
)

func TestBuildInfoVariables(t *testing.T) {
	// Test that build-time variables are set to expected defaults
	assert.Equal(t, "dev", Version)
	assert.Equal(t, "unknown", BuildTime)
	assert.Equal(t, "unknown", GitCommit)
}

func TestMainFunction(t *testing.T) {
	// Test that we can create a CLI with the build info
	// This tests the integration between main and app packages
	buildInfo := app.BuildInfo{
		Version:   Version,
		BuildTime: BuildTime,
		GitCommit: GitCommit,
	}

	cli := app.NewCLI(buildInfo)
	assert.NotNil(t, cli)

	// Test that we can call the version display without panic
	cli.ShowVersion()
}

func TestMainWithVersionFlag(t *testing.T) {
	// Test main function behavior with version flag
	// This is an integration test showing main delegates properly to CLI
	buildInfo := app.BuildInfo{
		Version:   "v1.0.0",
		BuildTime: "2024-01-01T00:00:00Z",
		GitCommit: "abc123",
	}

	cli := app.NewCLI(buildInfo)

	// Test version flag handling
	err := cli.Run([]string{"app", "--version"})
	assert.NoError(t, err)
}