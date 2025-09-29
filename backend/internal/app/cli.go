package app

import (
	"flag"
	"fmt"
	"log"
)

// BuildInfo contains build-time information
type BuildInfo struct {
	Version   string
	BuildTime string
	GitCommit string
}

// CLIConfig contains parsed command-line configuration
type CLIConfig struct {
	ShowVersion    bool
	EnvFile        string
	ShouldMigrate  bool
}

// CLI handles command-line interface operations
type CLI struct {
	buildInfo BuildInfo
	envLoader EnvironmentLoader
}

// NewCLI creates a new CLI instance
func NewCLI(buildInfo BuildInfo) *CLI {
	return &CLI{
		buildInfo: buildInfo,
		envLoader: NewEnvironmentLoader(),
	}
}

// ShowVersion displays version information
func (c *CLI) ShowVersion() {
	fmt.Printf("Burndler v%s\n", c.buildInfo.Version)
	fmt.Printf("Build Time: %s\n", c.buildInfo.BuildTime)
	fmt.Printf("Git Commit: %s\n", c.buildInfo.GitCommit)
}

// ParseFlags parses command-line arguments and returns configuration
func (c *CLI) ParseFlags(args []string) (*CLIConfig, error) {
	// Create a new flag set for testing purposes
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)

	config := &CLIConfig{}

	fs.BoolVar(&config.ShowVersion, "version", false, "Show version information")
	fs.BoolVar(&config.ShowVersion, "v", false, "Show version information (shorthand)")
	fs.StringVar(&config.EnvFile, "env", "", "Path to environment file (default: .env.development then .env)")

	// Parse flags
	err := fs.Parse(args[1:])
	if err != nil {
		return nil, err
	}

	// Check for migrate command
	remainingArgs := fs.Args()
	if len(remainingArgs) > 0 && remainingArgs[0] == "migrate" {
		config.ShouldMigrate = true
	}

	return config, nil
}

// Run executes the CLI with given arguments
func (c *CLI) Run(args []string) error {
	config, err := c.ParseFlags(args)
	if err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Handle version flag
	if config.ShowVersion {
		c.ShowVersion()
		return nil
	}

	// Load environment files in development mode
	isDev := c.buildInfo.Version == "dev"
	if err := c.envLoader.LoadEnvironment(config.EnvFile, isDev); err != nil {
		log.Printf("Warning: Failed to load environment: %v", err)
	}

	// Handle migrate command
	if config.ShouldMigrate {
		runner := NewMigrationRunner()
		return runner.RunMigrations()
	}

	// Normal application startup
	application, err := New()
	if err != nil {
		return fmt.Errorf("failed to initialize application: %w", err)
	}
	defer func() {
		if closeErr := application.Close(); closeErr != nil {
			log.Printf("Error closing application: %v", closeErr)
		}
	}()

	// Run the application
	if err := application.Run(); err != nil {
		return fmt.Errorf("application error: %w", err)
	}

	return nil
}