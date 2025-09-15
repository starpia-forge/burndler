package server

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
	"github.com/burndler/burndler/internal/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Server represents the HTTP server
type Server struct {
	config       *config.Config
	db           *gorm.DB
	merger       *services.Merger
	linter       *services.Linter
	packager     *services.Packager
	authService  *services.AuthService
	setupService *services.SetupService
	router       *gin.Engine
}

// New creates a new server instance
func New(cfg *config.Config, db *gorm.DB, merger *services.Merger, linter *services.Linter, packager *services.Packager) *Server {
	authService := services.NewAuthService(cfg, db)
	setupService := services.NewSetupService(db, cfg)
	s := &Server{
		config:       cfg,
		db:           db,
		merger:       merger,
		linter:       linter,
		packager:     packager,
		authService:  authService,
		setupService: setupService,
	}
	s.setupRouter()
	return s
}

// setupRouter configures all routes and middleware
func (s *Server) setupRouter() {
	s.router = gin.Default()

	// CORS middleware
	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     s.config.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(s.authService, s.db)
	setupHandler := handlers.NewSetupHandler(s.setupService, s.db)
	composeHandler := handlers.NewComposeHandler(s.merger, s.linter)
	packageHandler := handlers.NewPackageHandler(s.packager, s.db)

	// API v1 routes
	v1 := s.router.Group("/api/v1")

	// Setup middleware - protect all routes except setup and health
	v1.Use(middleware.SetupGuard(s.setupService))

	// Public routes (always accessible)
	v1.GET("/health", healthHandler.Health)

	// Setup routes (accessible during setup only)
	setup := v1.Group("/setup")
	setup.Use(middleware.SetupCompleteGuard(s.setupService))
	setup.GET("/status", setupHandler.GetStatus)
	setup.POST("/init", setupHandler.Initialize)
	setup.POST("/admin", setupHandler.CreateAdmin)
	setup.POST("/complete", setupHandler.Complete)

	// Authentication routes (public but blocked during setup)
	auth := v1.Group("/auth")
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.RefreshToken)

	// Protected routes
	protected := v1.Group("/")
	protected.Use(middleware.JWTAuth(s.config))

	// Compose operations
	protected.POST("/compose/merge", composeHandler.Merge)
	protected.POST("/compose/lint", composeHandler.Lint)

	// Package operations (Developer role only)
	protected.POST("/build/package", middleware.RequireRole("Developer"), packageHandler.Create)
	protected.GET("/build/status/:id", packageHandler.Status)
}

// Run starts the server and handles graceful shutdown
func (s *Server) Run() error {
	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", s.config.ServerPort),
		Handler:      s.router,
		ReadTimeout:  s.config.ServerReadTimeout,
		WriteTimeout: s.config.ServerWriteTimeout,
	}

	// Start server asynchronously
	errChan := make(chan error, 1)
	go func() {
		log.Printf("Server starting on port %s", s.config.ServerPort)
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
