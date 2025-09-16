package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/burndler/burndler/internal/app"
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
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Show version information (shorthand)")
	flag.Parse()

	// Handle version flag
	if showVersion {
		fmt.Printf("Burndler v%s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		fmt.Printf("Git Commit: %s\n", GitCommit)
		return
	}

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
	defer application.Close()

	// Run the application
	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

func runMigrations() error {
	// Initialize application for migrations only
	application, err := app.New()
	if err != nil {
		return err
	}
	defer application.Close()

	log.Println("Database migrations completed - application initialized successfully")
	return nil
}
