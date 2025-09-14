package main

import (
	"log"

	"github.com/burndler/burndler/internal/app"
)

func main() {
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