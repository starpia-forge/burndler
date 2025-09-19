package app

import (
	"log"

	"github.com/joho/godotenv"
)

// EnvironmentLoader interface for loading environment files
type EnvironmentLoader interface {
	LoadEnvironment(envFile string, isDev bool) error
}

// envLoader implements EnvironmentLoader
type envLoader struct{}

// NewEnvironmentLoader creates a new environment loader
func NewEnvironmentLoader() EnvironmentLoader {
	return &envLoader{}
}

// LoadEnvironment loads environment variables from files in development mode
func (e *envLoader) LoadEnvironment(envFile string, isDev bool) error {
	// Only load environment files in development mode
	if !isDev {
		return nil
	}

	if envFile != "" {
		// Use specified environment file
		if err := godotenv.Load(envFile); err != nil {
			log.Printf("Warning: Failed to load specified env file %s: %v", envFile, err)
		}
		return nil
	}

	// Try to load .env.development first, then .env as fallback
	if err := godotenv.Load(".env.development"); err != nil {
		// If .env.development doesn't exist, try .env
		if err := godotenv.Load(".env"); err != nil {
			log.Printf("Warning: No .env file found: %v", err)
		}
	}

	return nil
}