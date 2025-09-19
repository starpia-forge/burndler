package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/burndler/burndler/internal/app"
	"github.com/joho/godotenv"
)

// Build-time variables injected via ldflags
var (
	Version   = "dev"     // Version is set during build
	BuildTime = "unknown" // BuildTime is set during build
	GitCommit = "unknown" // GitCommit is set during build
)

func main() {
	// Parse command line flags
	var showVersion bool
	var envFile string
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Show version information (shorthand)")
	flag.StringVar(&envFile, "env", "", "Path to environment file (default: .env.development then .env)")
	flag.Parse()

	// Handle version flag
	if showVersion {
		fmt.Printf("Burndler v%s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		return
	}

	// Load environment files in development mode
	loadEnvFiles(envFile)

	// Check for migrate command
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		// Run migrations only
		if err := runMigrations(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Database migrations completed successfully")
		return
	}

	// Log version information on startup
	log.Printf("Starting Burndler v%s (built %s, commit %s)", Version, BuildTime, GitCommit)

	// Initialize application
	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer func() {
		if closeErr := application.Close(); closeErr != nil {
			log.Printf("Error closing application: %v", closeErr)
		}
	}()

	// Run the application
	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func runMigrations() error {
	// Load environment files in development mode for migrations
	loadEnvFiles("")

	// Initialize application for migrations only
	application, err := app.New()
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := application.Close(); closeErr != nil {
			log.Printf("Error closing application during migration: %v", closeErr)
		}
	}()

	log.Println("Database migrations completed - application initialized successfully")
	return nil
}

// loadEnvFiles loads environment files in development mode
func loadEnvFiles(envFile string) {
	if Version == "dev" {
		if envFile != "" {
			// Use specified environment file
			if err := godotenv.Load(envFile); err != nil {
				log.Printf("Warning: Failed to load specified env file %s: %v", envFile, err)
			}
		} else {
			// Try to load .env.development first, then .env as fallback
			if err := godotenv.Load(".env.development"); err != nil {
				// If .env.development doesn't exist, try .env
				if err := godotenv.Load(".env"); err != nil {
					log.Printf("Warning: No .env file found: %v", err)
				}
			}
		}
	}
}
