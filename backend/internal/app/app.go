package app

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

// App contains all application dependencies
type App struct {
	Config   *config.Config
	DB       *gorm.DB
	Storage  storage.Storage
	Merger   *services.Merger
	Linter   *services.Linter
	Packager *services.Packager
}

// New creates and initializes a new App instance
func New() (*App, error) {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := initDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	if err := db.AutoMigrate(&models.User{}, &models.Build{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Initialize storage
	store, err := initStorage(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize services
	merger := services.NewMerger()
	linter := services.NewLinter()
	packager := services.NewPackager(store)

	return &App{
		Config:   cfg,
		DB:       db,
		Storage:  store,
		Merger:   merger,
		Linter:   linter,
		Packager: packager,
	}, nil
}

// NewWithConfig creates a new App instance with a provided config (useful for testing)
func NewWithConfig(cfg *config.Config) (*App, error) {
	// Initialize database
	db, err := initDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	if err := db.AutoMigrate(&models.User{}, &models.Build{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Initialize storage
	store, err := initStorage(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize services
	merger := services.NewMerger()
	linter := services.NewLinter()
	packager := services.NewPackager(store)

	return &App{
		Config:   cfg,
		DB:       db,
		Storage:  store,
		Merger:   merger,
		Linter:   linter,
		Packager: packager,
	}, nil
}

// Close gracefully closes all app resources
func (a *App) Close() error {
	if a.DB != nil {
		sqlDB, err := a.DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// Run starts the application and handles graceful shutdown
func (a *App) Run() error {
	// Setup router
	router := a.setupRouter()

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", a.Config.ServerPort),
		Handler:      router,
		ReadTimeout:  a.Config.ServerReadTimeout,
		WriteTimeout: a.Config.ServerWriteTimeout,
	}

	// Start server asynchronously
	errChan := make(chan error, 1)
	go func() {
		log.Printf("Server starting on port %s", a.Config.ServerPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
		close(errChan)
	}()

	// Wait for interrupt signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	case <-quit:
		log.Println("Shutting down server...")
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server exited")
	return nil
}

// setupRouter configures all routes and middleware
func (a *App) setupRouter() *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     a.Config.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	composeHandler := handlers.NewComposeHandler(a.Merger, a.Linter)
	packageHandler := handlers.NewPackageHandler(a.Packager, a.DB)

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Public routes
	v1.GET("/health", healthHandler.Health)

	// Protected routes
	protected := v1.Group("/")
	protected.Use(middleware.JWTAuth(a.Config))

	// Compose operations
	protected.POST("/compose/merge", composeHandler.Merge)
	protected.POST("/compose/lint", composeHandler.Lint)

	// Package operations (Developer role only)
	protected.POST("/build/package", middleware.RequireRole("Developer"), packageHandler.Create)
	protected.GET("/build/status/:id", packageHandler.Status)

	return router
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