package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/burndler/burndler/internal/app"
	"github.com/burndler/burndler/internal/handlers"
	"github.com/burndler/burndler/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	app        *app.App
	router     *gin.Engine
	httpServer *http.Server
}

// New creates a new server instance
func New(application *app.App) *Server {
	s := &Server{
		app: application,
	}
	s.setupRouter()
	return s
}

// setupRouter configures all routes and middleware
func (s *Server) setupRouter() {
	router := gin.Default()

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     s.app.Config.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	composeHandler := handlers.NewComposeHandler(s.app.Merger, s.app.Linter)
	packageHandler := handlers.NewPackageHandler(s.app.Packager, s.app.DB)

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Public routes
	v1.GET("/health", healthHandler.Health)

	// Protected routes
	protected := v1.Group("/")
	protected.Use(middleware.JWTAuth(s.app.Config))

	// Compose operations
	protected.POST("/compose/merge", composeHandler.Merge)
	protected.POST("/compose/lint", composeHandler.Lint)

	// Package operations (Developer role only)
	protected.POST("/build/package", middleware.RequireRole("Developer"), packageHandler.Create)
	protected.GET("/build/status/:id", packageHandler.Status)

	s.router = router
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%s", s.app.Config.ServerPort),
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return s.httpServer.ListenAndServe()
}

// StartAsync starts the HTTP server in a goroutine
func (s *Server) StartAsync() <-chan error {
	errChan := make(chan error, 1)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%s", s.app.Config.ServerPort),
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
		close(errChan)
	}()

	return errChan
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// Router returns the gin router (useful for testing)
func (s *Server) Router() *gin.Engine {
	return s.router
}

// Port returns the server port
func (s *Server) Port() string {
	return s.app.Config.ServerPort
}