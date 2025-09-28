package services

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/burndler/burndler/internal/models"
	"github.com/burndler/burndler/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupModuleServiceTest(t *testing.T) (*ModuleService, *gorm.DB) {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Drop and recreate tables to ensure fresh schema
	err = db.Migrator().DropTable(&models.Module{}, &models.ModuleVersion{})
	require.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(&models.Module{}, &models.ModuleVersion{})
	require.NoError(t, err)

	// Create storage and linter
	store := &mockStorage{} // Mock storage
	linter := NewLinter()

	service := NewModuleService(db, store, linter)
	return service, db
}

// mockStorage implements storage.Storage interface for testing
type mockStorage struct{}

func (m *mockStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64) (string, error) {
	return "mock://" + key, nil
}

func (m *mockStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, nil
}

func (m *mockStorage) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *mockStorage) Exists(ctx context.Context, key string) (bool, error) {
	return true, nil
}

func (m *mockStorage) List(ctx context.Context, prefix string) ([]storage.FileInfo, error) {
	return nil, nil
}

func (m *mockStorage) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return "mock://" + key, nil
}

func TestModuleService_CreateModule(t *testing.T) {
	service, _ := setupModuleServiceTest(t)

	tests := []struct {
		name        string
		req         CreateModuleRequest
		expectError bool
	}{
		{
			name: "successful module creation",
			req: CreateModuleRequest{
				Name:        "webapp",
				Description: "Web application module",
				Author:      "test-author",
				Repository:  "https://github.com/test/webapp",
			},
			expectError: false,
		},
		{
			name: "duplicate module name",
			req: CreateModuleRequest{
				Name:        "webapp", // Same name as above
				Description: "Another web app",
				Author:      "different-author",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module, err := service.CreateModule(tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, module)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, module)
				assert.Equal(t, tt.req.Name, module.Name)
				assert.Equal(t, tt.req.Description, module.Description)
				assert.Equal(t, tt.req.Author, module.Author)
				assert.Equal(t, tt.req.Repository, module.Repository)
				assert.True(t, module.Active)
			}
		})
	}
}

func TestModuleService_GetModule(t *testing.T) {
	service, db := setupModuleServiceTest(t)

	// Create test module
	module := &models.Module{
		Name:        "test-module",
		Description: "Test module",
		Author:      "test-author",
		Active:      true,
	}
	require.NoError(t, db.Create(module).Error)

	tests := []struct {
		name            string
		moduleID        uint
		includeVersions bool
		expectError     bool
	}{
		{
			name:            "get existing module",
			moduleID:        module.ID,
			includeVersions: false,
			expectError:     false,
		},
		{
			name:            "get non-existent module",
			moduleID:        999,
			includeVersions: false,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetModule(tt.moduleID, tt.includeVersions)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, module.Name, result.Name)
			}
		})
	}
}

func TestModuleService_ListModules(t *testing.T) {
	service, db := setupModuleServiceTest(t)

	// Create test modules
	module1 := &models.Module{Name: "module1", Author: "author1", Active: true}
	module2 := &models.Module{Name: "module2", Author: "author2", Active: true}
	module3 := &models.Module{Name: "module3", Author: "author1", Active: false}

	require.NoError(t, db.Create(module1).Error)
	require.NoError(t, db.Create(module2).Error)
	require.NoError(t, db.Create(module3).Error)

	// Explicitly update module3 to be inactive to override the default
	require.NoError(t, db.Model(module3).Update("active", false).Error)

	tests := []struct {
		name           string
		filters        ModuleFilters
		expectedCount  int
		expectedTotal  int64
		expectedPages  int
	}{
		{
			name:           "list all modules",
			filters:        ModuleFilters{Page: 1, PageSize: 10},
			expectedCount:  3,
			expectedTotal:  3,
			expectedPages:  1,
		},
		{
			name:           "list active modules only",
			filters:        ModuleFilters{Active: &[]bool{true}[0], Page: 1, PageSize: 10},
			expectedCount:  2,
			expectedTotal:  2,
			expectedPages:  1,
		},
		{
			name:           "filter by author",
			filters:        ModuleFilters{Author: "author1", Page: 1, PageSize: 10},
			expectedCount:  2,
			expectedTotal:  2,
			expectedPages:  1,
		},
		{
			name:           "pagination test",
			filters:        ModuleFilters{Page: 1, PageSize: 2},
			expectedCount:  2,
			expectedTotal:  3,
			expectedPages:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ListModules(tt.filters)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Len(t, result.Data, tt.expectedCount)
			assert.Equal(t, tt.expectedTotal, result.Total)
			assert.Equal(t, tt.expectedPages, result.TotalPages)
		})
	}
}

func TestModuleService_CreateVersion(t *testing.T) {
	service, db := setupModuleServiceTest(t)

	// Create test module
	module := &models.Module{
		Name:   "test-module",
		Active: true,
	}
	require.NoError(t, db.Create(module).Error)

	validCompose := `
version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
`

	tests := []struct {
		name        string
		moduleID    uint
		req         CreateVersionRequest
		expectError bool
	}{
		{
			name:     "successful version creation",
			moduleID: module.ID,
			req: CreateVersionRequest{
				Version: "v1.0.0",
				Compose: validCompose,
				Variables: map[string]interface{}{
					"port": 8080,
					"env":  "production",
				},
			},
			expectError: false,
		},
		{
			name:     "duplicate version",
			moduleID: module.ID,
			req: CreateVersionRequest{
				Version: "v1.0.0", // Same version as above
				Compose: validCompose,
			},
			expectError: true,
		},
		{
			name:     "invalid module ID",
			moduleID: 999,
			req: CreateVersionRequest{
				Version: "v1.1.0",
				Compose: validCompose,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := service.CreateVersion(tt.moduleID, tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, version)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, version)
				assert.Equal(t, tt.req.Version, version.Version)
				assert.Equal(t, tt.req.Compose, version.ComposeContent)
				assert.False(t, version.Published)
			}
		})
	}
}

func TestModuleService_PublishVersion(t *testing.T) {
	service, db := setupModuleServiceTest(t)

	// Create test module and version
	module := &models.Module{
		Name:   "test-module",
		Active: true,
	}
	require.NoError(t, db.Create(module).Error)

	tests := []struct {
		name        string
		setupFunc   func() *models.ModuleVersion
		moduleID    uint
		version     string
		expectError bool
	}{
		{
			name: "successful publish",
			setupFunc: func() *models.ModuleVersion {
				version := &models.ModuleVersion{
					ModuleID:       module.ID,
					Version:        "v1.0.0",
					ComposeContent: "version: '3.8'\nservices:\n  web:\n    image: nginx@sha256:abcd1234",
					Published:      false,
				}
				require.NoError(t, db.Create(version).Error)
				return version
			},
			moduleID:    module.ID,
			version:     "v1.0.0",
			expectError: false,
		},
		{
			name: "publish already published version",
			setupFunc: func() *models.ModuleVersion {
				version := &models.ModuleVersion{
					ModuleID:       module.ID,
					Version:        "v1.1.0",
					ComposeContent: "version: '3.8'\nservices:\n  web:\n    image: nginx@sha256:abcd1234",
					Published:      true,
				}
				require.NoError(t, db.Create(version).Error)
				return version
			},
			moduleID:    module.ID,
			version:     "v1.1.0",
			expectError: true,
		},
		{
			name: "publish non-existent version",
			setupFunc: func() *models.ModuleVersion {
				return nil // No setup needed
			},
			moduleID:    module.ID,
			version:     "v2.0.0",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupFunc()
			result, err := service.PublishVersion(tt.moduleID, tt.version)

			if tt.expectError {
				assert.Error(t, err)
				if tt.name != "publish non-existent version" {
					// For already published, we get a result but with error
					if result != nil {
						assert.True(t, result.Published)
					}
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.Published)
				assert.NotNil(t, result.PublishedAt)
			}
		})
	}
}

func TestModuleService_DeleteModule(t *testing.T) {
	service, db := setupModuleServiceTest(t)

	// Create test module without published versions
	module1 := &models.Module{
		Name:   "test-module-1",
		Active: true,
	}
	require.NoError(t, db.Create(module1).Error)

	// Create test module with published version
	module2 := &models.Module{
		Name:   "test-module-2",
		Active: true,
	}
	require.NoError(t, db.Create(module2).Error)

	version := &models.ModuleVersion{
		ModuleID:       module2.ID,
		Version:        "v1.0.0",
		ComposeContent: "version: '3.8'\nservices:\n  web:\n    image: nginx:latest",
		Published:      true,
	}
	require.NoError(t, db.Create(version).Error)

	tests := []struct {
		name        string
		moduleID    uint
		expectError bool
	}{
		{
			name:        "delete module without published versions",
			moduleID:    module1.ID,
			expectError: false,
		},
		{
			name:        "delete module with published versions",
			moduleID:    module2.ID,
			expectError: true,
		},
		{
			name:        "delete non-existent module",
			moduleID:    999,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteModule(tt.moduleID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}