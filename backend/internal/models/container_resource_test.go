package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupResourceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&Container{},
		&ContainerVersion{},
		&ContainerResource{},
	)
	require.NoError(t, err)

	return db
}

func TestContainerResource_TableName(t *testing.T) {
	resource := ContainerResource{}
	assert.Equal(t, "container_resources", resource.TableName())
}

func TestContainerResource_BeforeCreate_Validation(t *testing.T) {
	tests := []struct {
		name        string
		resource    ContainerResource
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid resource",
			resource: ContainerResource{
				ContainerVersionID: 1,
				Path:               "config/app.yaml",
				StorageKey:         "storage/key",
			},
			expectError: false,
		},
		{
			name: "missing container_version_id",
			resource: ContainerResource{
				Path:       "config/app.yaml",
				StorageKey: "storage/key",
			},
			expectError: true,
			errorMsg:    "container_version_id is required",
		},
		{
			name: "missing path",
			resource: ContainerResource{
				ContainerVersionID: 1,
				StorageKey:         "storage/key",
			},
			expectError: true,
			errorMsg:    "path is required",
		},
		{
			name: "missing storage_key",
			resource: ContainerResource{
				ContainerVersionID: 1,
				Path:               "config/app.yaml",
			},
			expectError: true,
			errorMsg:    "storage_key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.resource.BeforeCreate(nil)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContainerResource_Create(t *testing.T) {
	db := setupResourceTestDB(t)

	// Create container and version
	container := &Container{Name: "nginx"}
	require.NoError(t, db.Create(container).Error)

	version := &ContainerVersion{
		ContainerID:    container.ID,
		Version:        "v1.0.0",
		ComposeContent: "version: '3'",
	}
	require.NoError(t, db.Create(version).Error)

	// Create resource
	resource := &ContainerResource{
		ContainerVersionID: version.ID,
		Path:               "config/nginx.conf",
		StorageKey:         "containers/nginx/v1.0.0/config/nginx.conf",
		FileSize:           1024,
		Checksum:           "abc123def456",
		IsDirectory:        false,
		ContentType:        "text/plain",
	}

	err := db.Create(resource).Error
	require.NoError(t, err)
	assert.NotZero(t, resource.ID)
	assert.NotZero(t, resource.CreatedAt)
	assert.NotZero(t, resource.UpdatedAt)
}

func TestContainerResource_Relationships(t *testing.T) {
	db := setupResourceTestDB(t)

	// Create container and version
	container := &Container{Name: "postgres"}
	require.NoError(t, db.Create(container).Error)

	version := &ContainerVersion{
		ContainerID:    container.ID,
		Version:        "v14.0",
		ComposeContent: "version: '3'",
	}
	require.NoError(t, db.Create(version).Error)

	// Create resources
	resources := []*ContainerResource{
		{
			ContainerVersionID: version.ID,
			Path:               "data/init.sql",
			StorageKey:         "containers/postgres/v14.0/data/init.sql",
			FileSize:           2048,
			Checksum:           "checksum1",
		},
		{
			ContainerVersionID: version.ID,
			Path:               "config/postgresql.conf",
			StorageKey:         "containers/postgres/v14.0/config/postgresql.conf",
			FileSize:           4096,
			Checksum:           "checksum2",
		},
	}

	for _, r := range resources {
		require.NoError(t, db.Create(r).Error)
	}

	// Load version with resources
	var loadedVersion ContainerVersion
	err := db.Preload("Resources").First(&loadedVersion, version.ID).Error
	require.NoError(t, err)
	assert.Len(t, loadedVersion.Resources, 2)
	assert.Equal(t, "data/init.sql", loadedVersion.Resources[0].Path)
	assert.Equal(t, "config/postgresql.conf", loadedVersion.Resources[1].Path)
}

func TestContainerResource_RequiredFields(t *testing.T) {
	db := setupResourceTestDB(t)

	// Create container and version
	container := &Container{Name: "redis"}
	require.NoError(t, db.Create(container).Error)

	version := &ContainerVersion{
		ContainerID:    container.ID,
		Version:        "v6.0",
		ComposeContent: "version: '3'",
	}
	require.NoError(t, db.Create(version).Error)

	tests := []struct {
		name     string
		resource *ContainerResource
		wantErr  bool
	}{
		{
			name: "valid resource",
			resource: &ContainerResource{
				ContainerVersionID: version.ID,
				Path:               "config.yaml",
				StorageKey:         "key",
				FileSize:           100,
			},
			wantErr: false,
		},
		{
			name: "missing container_version_id",
			resource: &ContainerResource{
				Path:       "config.yaml",
				StorageKey: "key",
				FileSize:   100,
			},
			wantErr: true,
		},
		{
			name: "missing path",
			resource: &ContainerResource{
				ContainerVersionID: version.ID,
				StorageKey:         "key",
				FileSize:           100,
			},
			wantErr: true,
		},
		{
			name: "missing storage_key",
			resource: &ContainerResource{
				ContainerVersionID: version.ID,
				Path:               "config.yaml",
				FileSize:           100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.Create(tt.resource).Error
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestContainerResource_CascadeDelete(t *testing.T) {
	db := setupResourceTestDB(t)

	// Create container and version
	container := &Container{Name: "mysql"}
	require.NoError(t, db.Create(container).Error)

	version := &ContainerVersion{
		ContainerID:    container.ID,
		Version:        "v8.0",
		ComposeContent: "version: '3'",
	}
	require.NoError(t, db.Create(version).Error)

	// Create resource
	resource := &ContainerResource{
		ContainerVersionID: version.ID,
		Path:               "init.sql",
		StorageKey:         "key",
		FileSize:           1000,
	}
	require.NoError(t, db.Create(resource).Error)

	// Delete resources first (manual cascade for SQLite testing)
	err := db.Where("container_version_id = ?", version.ID).Delete(&ContainerResource{}).Error
	require.NoError(t, err)

	// Delete version
	err = db.Delete(version).Error
	require.NoError(t, err)

	// Verify resources are deleted
	var count int64
	db.Model(&ContainerResource{}).Where("container_version_id = ?", version.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestContainerResource_DirectoryFlag(t *testing.T) {
	db := setupResourceTestDB(t)

	container := &Container{Name: "app"}
	require.NoError(t, db.Create(container).Error)

	version := &ContainerVersion{
		ContainerID:    container.ID,
		Version:        "v1.0.0",
		ComposeContent: "version: '3'",
	}
	require.NoError(t, db.Create(version).Error)

	tests := []struct {
		name        string
		path        string
		isDirectory bool
	}{
		{
			name:        "file resource",
			path:        "config/app.yaml",
			isDirectory: false,
		},
		{
			name:        "directory resource",
			path:        "config/",
			isDirectory: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := &ContainerResource{
				ContainerVersionID: version.ID,
				Path:               tt.path,
				StorageKey:         "key-" + tt.name,
				FileSize:           0,
				IsDirectory:        tt.isDirectory,
			}

			err := db.Create(resource).Error
			require.NoError(t, err)

			var loaded ContainerResource
			err = db.First(&loaded, resource.ID).Error
			require.NoError(t, err)
			assert.Equal(t, tt.isDirectory, loaded.IsDirectory)
		})
	}
}
