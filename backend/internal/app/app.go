package app

import (
	"fmt"

	"github.com/burndler/burndler/internal/config"
	"github.com/burndler/burndler/internal/models"
	"github.com/burndler/burndler/internal/server"
	"github.com/burndler/burndler/internal/services"
	"github.com/burndler/burndler/internal/storage"
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
	if err := db.AutoMigrate(
		&models.User{},
		&models.Container{},
		&models.ContainerVersion{},
		&models.Service{},
		&models.ServiceContainer{},
		&models.Build{},
		&models.Setup{},
	); err != nil {
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
	if err := db.AutoMigrate(
		&models.User{},
		&models.Container{},
		&models.ContainerVersion{},
		&models.Service{},
		&models.ServiceContainer{},
		&models.Build{},
		&models.Setup{},
	); err != nil {
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
	// Create and run server
	srv := server.New(a.Config, a.DB, a.Storage, a.Merger, a.Linter, a.Packager)
	return srv.Run()
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
