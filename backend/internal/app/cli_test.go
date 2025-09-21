package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCLI(t *testing.T) {
	buildInfo := BuildInfo{
		Version:   "v1.0.0",
		BuildTime: "2024-01-01T00:00:00Z",
		GitCommit: "abc123",
	}

	cli := NewCLI(buildInfo)

	assert.NotNil(t, cli)
	assert.Equal(t, buildInfo, cli.buildInfo)
}

func TestCLI_ShowVersion(t *testing.T) {
	buildInfo := BuildInfo{
		Version:   "v1.2.3",
		BuildTime: "2024-01-01T12:00:00Z",
		GitCommit: "def456",
	}

	cli := NewCLI(buildInfo)

	// ShowVersion uses fmt.Printf which prints to stdout, not log
	// For testing, we'll just verify the method doesn't panic
	// In a more sophisticated test, we could capture stdout
	cli.ShowVersion()

	// Just verify the BuildInfo is set correctly
	assert.Equal(t, "v1.2.3", cli.buildInfo.Version)
	assert.Equal(t, "2024-01-01T12:00:00Z", cli.buildInfo.BuildTime)
	assert.Equal(t, "def456", cli.buildInfo.GitCommit)
}

func TestCLI_ParseFlags_Version(t *testing.T) {
	buildInfo := BuildInfo{
		Version:   "v1.0.0",
		BuildTime: "2024-01-01T00:00:00Z",
		GitCommit: "abc123",
	}

	cli := NewCLI(buildInfo)

	// Test version flag
	config, err := cli.ParseFlags([]string{"app", "--version"})
	require.NoError(t, err)
	assert.True(t, config.ShowVersion)
}

func TestCLI_ParseFlags_VersionShort(t *testing.T) {
	buildInfo := BuildInfo{
		Version:   "v1.0.0",
		BuildTime: "2024-01-01T00:00:00Z",
		GitCommit: "abc123",
	}

	cli := NewCLI(buildInfo)

	// Test version flag short form
	config, err := cli.ParseFlags([]string{"app", "-v"})
	require.NoError(t, err)
	assert.True(t, config.ShowVersion)
}

func TestCLI_ParseFlags_EnvFile(t *testing.T) {
	buildInfo := BuildInfo{
		Version:   "v1.0.0",
		BuildTime: "2024-01-01T00:00:00Z",
		GitCommit: "abc123",
	}

	cli := NewCLI(buildInfo)

	// Test env file flag
	config, err := cli.ParseFlags([]string{"app", "--env", ".env.test"})
	require.NoError(t, err)
	assert.Equal(t, ".env.test", config.EnvFile)
}

func TestCLI_ParseFlags_MigrateCommand(t *testing.T) {
	buildInfo := BuildInfo{
		Version:   "v1.0.0",
		BuildTime: "2024-01-01T00:00:00Z",
		GitCommit: "abc123",
	}

	cli := NewCLI(buildInfo)

	// Test migrate command
	config, err := cli.ParseFlags([]string{"app", "migrate"})
	require.NoError(t, err)
	assert.True(t, config.ShouldMigrate)
}

func TestCLI_Run_ShowVersion(t *testing.T) {
	buildInfo := BuildInfo{
		Version:   "v1.0.0",
		BuildTime: "2024-01-01T00:00:00Z",
		GitCommit: "abc123",
	}

	cli := NewCLI(buildInfo)

	err := cli.Run([]string{"app", "--version"})
	require.NoError(t, err)

	// Just verify no error occurred - version was displayed
	// In a more sophisticated test, we could capture stdout
}

func TestCLI_Run_Migration(t *testing.T) {
	buildInfo := BuildInfo{
		Version:   "dev",
		BuildTime: "2024-01-01T00:00:00Z",
		GitCommit: "abc123",
	}

	cli := NewCLI(buildInfo)

	err := cli.Run([]string{"app", "migrate"})

	if os.Getenv("CI") == "true" {
		// In CI environment, database is available so migration should succeed
		assert.NoError(t, err)
	} else {
		// In local environment without database, expect failure
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "failed to initialize application")
		}
	}
}

func TestCLI_Run_NormalStartup(t *testing.T) {
	buildInfo := BuildInfo{
		Version:   "dev",
		BuildTime: "2024-01-01T00:00:00Z",
		GitCommit: "abc123",
	}

	cli := NewCLI(buildInfo)

	err := cli.Run([]string{"app"})

	if os.Getenv("CI") == "true" {
		// In CI environment, database is available so startup should succeed
		assert.NoError(t, err)
	} else {
		// In local environment without database, expect failure
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "failed to initialize application")
		}
	}
}