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
	"github.com/burndler/burndler/internal/static"
	"github.com/burndler/burndler/internal/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Server represents the HTTP server
type Server struct {
	config        *config.Config
	db            *gorm.DB
	storage       storage.Storage
	merger        *services.Merger
	linter        *services.Linter
	packager      *services.Packager
	authService      *services.AuthService
	setupService     *services.SetupService
	containerService *services.ContainerService
	serviceService   *services.ServiceService
	router           *gin.Engine
}

// New creates a new server instance
func New(cfg *config.Config, db *gorm.DB, storage storage.Storage, merger *services.Merger, linter *services.Linter, packager *services.Packager) *Server {
	authService := services.NewAuthService(cfg, db)
	setupService := services.NewSetupService(db, cfg)
	containerService := services.NewContainerService(db, storage, linter)
	serviceService := services.NewServiceService(db, storage)
	s := &Server{
		config:        cfg,
		db:            db,
		storage:       storage,
		merger:        merger,
		linter:        linter,
		packager:      packager,
		authService:      authService,
		setupService:     setupService,
		containerService: containerService,
		serviceService:   serviceService,
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
	containerHandler := handlers.NewContainerHandler(s.containerService, s.db)
	serviceHandler := handlers.NewServiceHandler(s.serviceService, s.db)

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

	// Protected auth routes
	authProtected := auth.Group("/")
	authProtected.Use(middleware.JWTAuth(s.config))
	authProtected.GET("/me", authHandler.GetCurrentUser)

	// Protected routes
	protected := v1.Group("/")
	protected.Use(middleware.JWTAuth(s.config))

	// Compose operations
	protected.POST("/compose/merge", composeHandler.Merge)
	protected.POST("/compose/lint", composeHandler.Lint)

	// Package operations (Developer role only)
	protected.POST("/build/package", middleware.RequireRole("Developer"), packageHandler.Create)
	protected.GET("/build/status/:id", packageHandler.Status)

	// Container management
	containers := protected.Group("/containers")
	containers.GET("", containerHandler.ListContainers)
	containers.POST("", middleware.RequireRole("Developer"), containerHandler.CreateContainer)
	containers.GET("/:id", containerHandler.GetContainer)
	containers.PUT("/:id", middleware.RequireRole("Developer"), containerHandler.UpdateContainer)
	containers.DELETE("/:id", middleware.RequireRole("Developer"), containerHandler.DeleteContainer)

	// Container version management
	containers.GET("/:id/versions", containerHandler.ListVersions)
	containers.POST("/:id/versions", middleware.RequireRole("Developer"), containerHandler.CreateVersion)
	containers.GET("/:id/versions/:version", containerHandler.GetVersion)
	containers.PUT("/:id/versions/:version", middleware.RequireRole("Developer"), containerHandler.UpdateVersion)
	containers.POST("/:id/versions/:version/publish", middleware.RequireRole("Developer"), containerHandler.PublishVersion)

	// Service management
	serviceRoutes := protected.Group("/services")
	serviceRoutes.GET("", serviceHandler.ListServices)
	serviceRoutes.POST("", middleware.RequireRole("Developer"), serviceHandler.CreateService)
	serviceRoutes.GET("/:id", serviceHandler.GetService)
	serviceRoutes.PUT("/:id", middleware.RequireRole("Developer"), serviceHandler.UpdateService)
	serviceRoutes.DELETE("/:id", middleware.RequireRole("Developer"), serviceHandler.DeleteService)

	// Service container management
	serviceRoutes.GET("/:id/containers", serviceHandler.GetServiceContainers)
	serviceRoutes.POST("/:id/containers", middleware.RequireRole("Developer"), serviceHandler.AddContainerToService)
	serviceRoutes.PUT("/:id/containers/:container_id", middleware.RequireRole("Developer"), serviceHandler.UpdateServiceContainer)
	serviceRoutes.DELETE("/:id/containers/:container_id", middleware.RequireRole("Developer"), serviceHandler.RemoveContainerFromService)

	// Service operations
	serviceRoutes.POST("/:id/validate", serviceHandler.ValidateService)
	serviceRoutes.POST("/:id/build", middleware.RequireRole("Developer"), serviceHandler.BuildService)

	// Serve static files if enabled
	if s.config.ServeStaticFiles {
		s.setupStaticFileServing()
	}
}

// setupStaticFileServing configures static file serving with SPA routing fallback
func (s *Server) setupStaticFileServing() {
	// Try to get the SPA handler from embedded files first
	spaHandler, err := static.SPAHandler()
	if err != nil {
		log.Printf("Warning: Could not setup embedded file serving: %v", err)
		log.Printf("Falling back to filesystem serving from: %s", s.config.StaticFilesPath)

		// Fallback to filesystem serving
		s.router.Static("/static", s.config.StaticFilesPath)
		s.router.StaticFile("/", s.config.StaticFilesPath+"/index.html")
		return
	}

	// Use embedded files - serve all static assets
	staticFileHandler, err := static.StaticFileHandler()
	if err != nil {
		log.Printf("Warning: Could not create static file handler: %v", err)
		return
	}

	// Serve static assets (JS, CSS, images, etc.)
	s.router.GET("/static/*filepath", gin.WrapH(staticFileHandler))
	s.router.GET("/assets/*filepath", gin.WrapH(staticFileHandler))

	// Handle favicon and other root assets
	s.router.GET("/favicon.ico", gin.WrapH(staticFileHandler))
	s.router.GET("/vite.svg", gin.WrapH(staticFileHandler))

	// SPA routing fallback - serve index.html for all non-API routes
	s.router.NoRoute(func(c *gin.Context) {
		// Skip API routes
		path := c.Request.URL.Path
		if path == "/api" || (len(path) >= 5 && path[:5] == "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}

		// Serve index.html for SPA routes
		spaHandler.ServeHTTP(c.Writer, c.Request)
	})
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
