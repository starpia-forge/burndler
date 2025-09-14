package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/burndler/burndler/internal/config"
	"github.com/burndler/burndler/internal/handlers"
	"github.com/burndler/burndler/internal/middleware"
	"github.com/burndler/burndler/internal/models"
	"github.com/burndler/burndler/internal/services"
	"github.com/burndler/burndler/internal/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run migrations
	if err := db.AutoMigrate(&models.User{}, &models.Build{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize storage
	store, err := initStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Initialize services
	merger := services.NewMerger()
	linter := services.NewLinter()
	packager := services.NewPackager(store)

	// Setup Gin router
	router := setupRouter(cfg, db, merger, linter, packager)

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on port %s", cfg.ServerPort)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.DBMaxConnections)
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConnections)
	sqlDB.SetConnMaxLifetime(cfg.DBConnectionLifetime)

	return db, nil
}

func initStorage(cfg *config.Config) (storage.Storage, error) {
	switch cfg.StorageMode {
	case "s3":
		return storage.NewS3Storage(cfg)
	case "local":
		return storage.NewLocalFSStorage(cfg)
	default:
		return nil, fmt.Errorf("unknown storage mode: %s", cfg.StorageMode)
	}
}

func setupRouter(
	cfg *config.Config,
	db *gorm.DB,
	merger *services.Merger,
	linter *services.Linter,
	packager *services.Packager,
) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	composeHandler := handlers.NewComposeHandler(merger, linter)
	packageHandler := handlers.NewPackageHandler(packager, db)

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Public routes
	v1.GET("/health", healthHandler.Health)

	// Protected routes
	protected := v1.Group("/")
	protected.Use(middleware.JWTAuth(cfg))

	// Compose operations
	protected.POST("/compose/merge", composeHandler.Merge)
	protected.POST("/compose/lint", composeHandler.Lint)

	// Package operations (Developer role only)
	protected.POST("/build/package", middleware.RequireRole("Developer"), packageHandler.Create)
	protected.GET("/build/status/:id", packageHandler.Status)

	return router
}