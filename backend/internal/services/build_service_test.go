package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/burndler/burndler/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupBuildServiceTest(t *testing.T) (*gorm.DB, *BuildService) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate schema
	err = db.AutoMigrate(
		&models.User{},
		&models.Build{},
		&models.Service{},
		&models.ServiceContainer{},
		&models.Container{},
		&models.ContainerVersion{},
		&models.ContainerConfiguration{},
		&models.ContainerFile{},
		&models.ContainerAsset{},
	)
	assert.NoError(t, err)

	// Create build service with nil storage for unit tests
	buildService := NewBuildService(db, nil)

	return db, buildService
}

func TestNewBuildService(t *testing.T) {
	db, buildService := setupBuildServiceTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	assert.NotNil(t, buildService)
	assert.NotNil(t, buildService.db)
	assert.NotNil(t, buildService.templateEngine)
	assert.NotNil(t, buildService.dependencyChecker)
	assert.NotNil(t, buildService.merger)
	assert.NotNil(t, buildService.linter)
	assert.NotNil(t, buildService.packager)
}

func TestBuildService_ResolveVariables(t *testing.T) {
	db, buildService := setupBuildServiceTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Create test data
	service := &models.Service{
		ID:   1,
		Name: "test-service",
	}

	// Service-level variables
	serviceVarsJSON, _ := json.Marshal(map[string]interface{}{
		"DB_HOST": "localhost",
		"DB_PORT": 5432,
	})
	service.Variables = datatypes.JSON(serviceVarsJSON)

	// Container version variables
	containerVarsJSON, _ := json.Marshal(map[string]interface{}{
		"DB_PORT": 3306, // This should be overridden
		"DB_NAME": "mydb",
	})

	containerVersion := &models.ContainerVersion{
		Variables: datatypes.JSON(containerVarsJSON),
	}

	// Service container with overrides
	overrideVarsJSON, _ := json.Marshal(map[string]interface{}{
		"DB_PORT": 5433, // Highest precedence
	})

	serviceContainer := &models.ServiceContainer{
		ContainerVersion: *containerVersion,
		OverrideVars:     datatypes.JSON(overrideVarsJSON),
	}

	config := &models.ContainerConfiguration{}

	// Test variable resolution
	variables := buildService.resolveVariables(service, serviceContainer, config)

	// Verify global variables
	assert.Equal(t, "test-service", variables["SERVICE_NAME"])
	assert.Equal(t, uint(1), variables["SERVICE_ID"])

	// Verify service variables
	assert.Equal(t, "localhost", variables["DB_HOST"])

	// Verify precedence: override > service > container
	assert.Equal(t, float64(5433), variables["DB_PORT"]) // Override wins

	// Verify container default (not overridden)
	assert.Equal(t, "mydb", variables["DB_NAME"])
}

func TestBuildService_ApplyNamespace(t *testing.T) {
	db, buildService := setupBuildServiceTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Create test data
	service := &models.Service{
		ID:   123,
		Name: "my-service",
	}

	container := &models.Container{
		ID:   456,
		Name: "my-container",
	}
	db.Create(container)

	// Test namespace application
	filePath := "config/app.yaml"
	namespacedPath := buildService.applyNamespace(filePath, container.ID, service)

	expected := "my-service_123/my-container/config/app.yaml"
	assert.Equal(t, expected, namespacedPath)
}

func TestBuildService_ValidateConfiguration(t *testing.T) {
	db, buildService := setupBuildServiceTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Create test user
	user := &models.User{
		Email: "test@example.com",
		Name:  "Test User",
		Role:  "Developer",
	}
	db.Create(user)

	// Create active service with enabled containers
	service := &models.Service{
		UserID: user.ID,
		Name:   "test-service",
		Active: true,
	}
	db.Create(service)

	// Create container and version
	container := &models.Container{
		Name: "test-container",
	}
	db.Create(container)

	containerVersion := &models.ContainerVersion{
		ContainerID: container.ID,
		Version:     "1.0.0",
	}
	db.Create(containerVersion)

	// Create enabled service container
	serviceContainer := &models.ServiceContainer{
		ServiceID:          service.ID,
		ContainerID:        container.ID,
		ContainerVersionID: containerVersion.ID,
		Enabled:            true,
	}
	db.Create(serviceContainer)

	// Reload service with relationships
	db.Preload("ServiceContainers").First(service, service.ID)

	// Create build
	build := &models.Build{
		UserID:    user.ID,
		ServiceID: &service.ID,
		Status:    "queued",
		Name:      "test-build",
	}
	db.Create(build)

	buildCtx := &BuildContext{
		Build:   build,
		Service: service,
	}

	// Test validation - should pass
	err := buildService.validateConfiguration(context.Background(), buildCtx)
	assert.NoError(t, err)
}

func TestBuildService_ValidateConfiguration_InactiveService(t *testing.T) {
	db, buildService := setupBuildServiceTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Create inactive service
	service := &models.Service{
		Name:   "inactive-service",
		Active: false,
	}

	buildCtx := &BuildContext{
		Service: service,
	}

	// Test validation - should fail
	err := buildService.validateConfiguration(context.Background(), buildCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not active")
}

func TestBuildService_ValidateConfiguration_NoContainers(t *testing.T) {
	db, buildService := setupBuildServiceTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Create active service with no containers
	service := &models.Service{
		Name:              "empty-service",
		Active:            true,
		ServiceContainers: []models.ServiceContainer{},
	}

	buildCtx := &BuildContext{
		Service: service,
	}

	// Test validation - should fail
	err := buildService.validateConfiguration(context.Background(), buildCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no enabled containers")
}

func TestBuildService_UpdateBuildStatus(t *testing.T) {
	db, buildService := setupBuildServiceTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Create test user
	user := &models.User{
		Email: "test@example.com",
		Name:  "Test User",
		Role:  "Developer",
	}
	db.Create(user)

	// Create build
	build := &models.Build{
		UserID:   user.ID,
		Status:   "queued",
		Name:     "test-build",
		Progress: 0,
	}
	db.Create(build)

	// Update status
	build.Status = "building"
	build.Progress = 50
	err := buildService.updateBuildStatus(build)
	assert.NoError(t, err)

	// Verify update
	var updatedBuild models.Build
	db.First(&updatedBuild, build.ID)
	assert.Equal(t, "building", updatedBuild.Status)
	assert.Equal(t, 50, updatedBuild.Progress)
}

func TestBuildService_EvaluateCondition(t *testing.T) {
	_, buildService := setupBuildServiceTest(t)

	variables := map[string]interface{}{
		"SSL_ENABLED": true,
		"PORT":        8080,
	}

	tests := []struct {
		name      string
		condition string
		want      bool
		wantErr   bool
	}{
		{
			name:      "simple true condition",
			condition: "true",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "simple false condition",
			condition: "false",
			want:      false,
			wantErr:   false,
		},
		{
			name:      "boolean equality",
			condition: "{{.SSL_ENABLED}} == true",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "number comparison",
			condition: "{{.PORT}} == 8080",
			want:      true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildService.EvaluateCondition(tt.condition, variables)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
