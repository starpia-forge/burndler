package main

import (
	"log"
	"os"

	"github.com/burndler/burndler/internal/app"
)

func main() {
	// Check for migrate command
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		// Run migrations only
		if err := runMigrations(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Database migrations completed successfully")
		return
	}

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
