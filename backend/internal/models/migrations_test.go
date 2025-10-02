package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupMigrationTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create tables with OLD structure manually for migration testing
	// This simulates the database state before migration

	// Containers table
	err = db.Exec(`
		CREATE TABLE containers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			author TEXT,
			repository TEXT,
			active INTEGER DEFAULT 1,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	// Container versions table
	err = db.Exec(`
		CREATE TABLE container_versions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			container_id INTEGER NOT NULL,
			version TEXT NOT NULL,
			compose_content TEXT NOT NULL,
			variables TEXT,
			resource_paths TEXT,
			dependencies TEXT,
			configuration_id INTEGER,
			published INTEGER DEFAULT 0,
			published_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	// Container configurations table with BOTH old and new columns
	// This simulates the state after AutoMigrate but before data migration
	err = db.Exec(`
		CREATE TABLE container_configurations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			container_version_id INTEGER,
			container_id INTEGER,
			name TEXT,
			description TEXT,
			minimum_version TEXT,
			ui_schema TEXT,
			dependency_rules TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	// Container files table with BOTH old and new columns
	err = db.Exec(`
		CREATE TABLE container_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			container_version_id INTEGER,
			container_configuration_id INTEGER,
			file_path TEXT NOT NULL,
			file_type TEXT NOT NULL,
			storage_path TEXT,
			template_format TEXT,
			display_condition TEXT,
			is_directory INTEGER DEFAULT 0,
			description TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	// Container assets table with BOTH old and new columns
	err = db.Exec(`
		CREATE TABLE container_assets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			container_version_id INTEGER,
			container_configuration_id INTEGER,
			original_file_name TEXT NOT NULL,
			file_path TEXT NOT NULL,
			asset_type TEXT NOT NULL,
			mime_type TEXT,
			file_size INTEGER NOT NULL,
			checksum TEXT NOT NULL,
			compressed INTEGER DEFAULT 0,
			include_condition TEXT,
			storage_type TEXT NOT NULL,
			storage_path TEXT NOT NULL,
			download_url TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	return db
}

func TestMigrateContainerConfigurationToContainerLevel(t *testing.T) {
	db := setupMigrationTestDB(t)

	// Create test data in OLD structure (simulating pre-migration state)
	container := &Container{Name: "nginx"}
	require.NoError(t, db.Create(container).Error)

	version1 := &ContainerVersion{
		ContainerID:    container.ID,
		Version:        "v1.0.0",
		ComposeContent: "test",
	}
	require.NoError(t, db.Create(version1).Error)

	version2 := &ContainerVersion{
		ContainerID:    container.ID,
		Version:        "v1.1.0",
		ComposeContent: "test",
	}
	require.NoError(t, db.Create(version2).Error)

	// Insert old-style configurations (without Name and MinimumVersion)
	// Simulate old structure by directly inserting
	err := db.Exec(`
		INSERT INTO container_configurations (container_version_id, ui_schema, dependency_rules)
		VALUES (?, '{}', '[]')
	`, version1.ID).Error
	require.NoError(t, err)

	var config1ID uint
	err = db.Raw(`SELECT id FROM container_configurations WHERE container_version_id = ?`, version1.ID).
		Scan(&config1ID).Error
	require.NoError(t, err)

	err = db.Exec(`
		INSERT INTO container_configurations (container_version_id, ui_schema, dependency_rules)
		VALUES (?, '{}', '[]')
	`, version2.ID).Error
	require.NoError(t, err)

	var config2ID uint
	err = db.Raw(`SELECT id FROM container_configurations WHERE container_version_id = ?`, version2.ID).
		Scan(&config2ID).Error
	require.NoError(t, err)

	// Create files and assets with old structure
	err = db.Exec(`
		INSERT INTO container_files (container_version_id, file_path, file_type, storage_path)
		VALUES (?, 'config.yaml', 'template', '/storage/config.yaml')
	`, version1.ID).Error
	require.NoError(t, err)

	err = db.Exec(`
		INSERT INTO container_assets (container_version_id, original_file_name, file_path, asset_type, mime_type, file_size, checksum, storage_type, storage_path)
		VALUES (?, 'data.tar.gz', 'data/data.tar.gz', 'data', 'application/gzip', 1000, 'abc123', 'embedded', '/storage/data.tar.gz')
	`, version1.ID).Error
	require.NoError(t, err)

	// Run migration
	err = MigrateContainerConfigurationToContainerLevel(db)
	require.NoError(t, err)

	// Verify migration results
	// 1. Check configurations have new structure
	var migratedConfig1 ContainerConfiguration
	err = db.First(&migratedConfig1, config1ID).Error
	require.NoError(t, err)
	assert.Equal(t, container.ID, migratedConfig1.ContainerID)
	assert.Equal(t, "default", migratedConfig1.Name)
	assert.Equal(t, "v1.0.0", migratedConfig1.MinimumVersion)

	var migratedConfig2 ContainerConfiguration
	err = db.First(&migratedConfig2, config2ID).Error
	require.NoError(t, err)
	assert.Equal(t, container.ID, migratedConfig2.ContainerID)
	assert.Equal(t, "default", migratedConfig2.Name)
	assert.Equal(t, "v1.1.0", migratedConfig2.MinimumVersion)

	// 2. Check versions reference configurations
	var migratedVersion1 ContainerVersion
	err = db.First(&migratedVersion1, version1.ID).Error
	require.NoError(t, err)
	require.NotNil(t, migratedVersion1.ConfigurationID)
	assert.Equal(t, config1ID, *migratedVersion1.ConfigurationID)

	var migratedVersion2 ContainerVersion
	err = db.First(&migratedVersion2, version2.ID).Error
	require.NoError(t, err)
	require.NotNil(t, migratedVersion2.ConfigurationID)
	assert.Equal(t, config2ID, *migratedVersion2.ConfigurationID)

	// 3. Check files migrated
	var migratedFile ContainerFile
	err = db.Where("file_path = ?", "config.yaml").First(&migratedFile).Error
	require.NoError(t, err)
	assert.Equal(t, config1ID, migratedFile.ContainerConfigurationID)

	// 4. Check assets migrated
	var migratedAsset ContainerAsset
	err = db.Where("original_file_name = ?", "data.tar.gz").First(&migratedAsset).Error
	require.NoError(t, err)
	assert.Equal(t, config1ID, migratedAsset.ContainerConfigurationID)
}

func TestMigrateContainerConfigurationToContainerLevel_AlreadyMigrated(t *testing.T) {
	db := setupMigrationTestDB(t)

	// Create test data in NEW structure (already migrated)
	container := &Container{Name: "nginx"}
	require.NoError(t, db.Create(container).Error)

	version := &ContainerVersion{
		ContainerID:    container.ID,
		Version:        "v1.0.0",
		ComposeContent: "test",
	}
	require.NoError(t, db.Create(version).Error)

	config := &ContainerConfiguration{
		ContainerID:     container.ID,
		Name:            "default",
		MinimumVersion:  "v1.0.0",
		UISchema:        datatypes.JSON([]byte(`{}`)),
		DependencyRules: datatypes.JSON([]byte(`[]`)),
	}
	require.NoError(t, db.Create(config).Error)

	version.ConfigurationID = &config.ID
	require.NoError(t, db.Save(version).Error)

	// Run migration - should be no-op
	err := MigrateContainerConfigurationToContainerLevel(db)
	require.NoError(t, err)

	// Verify nothing changed
	var checkConfig ContainerConfiguration
	err = db.First(&checkConfig, config.ID).Error
	require.NoError(t, err)
	assert.Equal(t, "default", checkConfig.Name)
	assert.Equal(t, "v1.0.0", checkConfig.MinimumVersion)
	assert.Equal(t, container.ID, checkConfig.ContainerID)
}

func TestMigrateContainerConfigurationToContainerLevel_EmptyDatabase(t *testing.T) {
	db := setupMigrationTestDB(t)

	// Run migration on empty database - should be no-op
	err := MigrateContainerConfigurationToContainerLevel(db)
	require.NoError(t, err)

	// Verify database is still empty
	var count int64
	db.Model(&ContainerConfiguration{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestMigrateContainerConfigurationToContainerLevel_MultipleContainers(t *testing.T) {
	db := setupMigrationTestDB(t)

	// Create multiple containers with configurations
	container1 := &Container{Name: "nginx"}
	require.NoError(t, db.Create(container1).Error)

	container2 := &Container{Name: "postgres"}
	require.NoError(t, db.Create(container2).Error)

	version1 := &ContainerVersion{
		ContainerID:    container1.ID,
		Version:        "v1.0.0",
		ComposeContent: "test",
	}
	require.NoError(t, db.Create(version1).Error)

	version2 := &ContainerVersion{
		ContainerID:    container2.ID,
		Version:        "v14.0",
		ComposeContent: "test",
	}
	require.NoError(t, db.Create(version2).Error)

	// Insert old-style configurations
	err := db.Exec(`
		INSERT INTO container_configurations (container_version_id, ui_schema, dependency_rules)
		VALUES (?, '{}', '[]'), (?, '{}', '[]')
	`, version1.ID, version2.ID).Error
	require.NoError(t, err)

	// Run migration
	err = MigrateContainerConfigurationToContainerLevel(db)
	require.NoError(t, err)

	// Verify both configurations migrated correctly
	var configs []ContainerConfiguration
	err = db.Find(&configs).Error
	require.NoError(t, err)
	require.Len(t, configs, 2)

	// Check each configuration
	for _, config := range configs {
		assert.Equal(t, "default", config.Name)
		assert.NotEmpty(t, config.MinimumVersion)
		assert.NotZero(t, config.ContainerID)

		switch config.ContainerID {
		case container1.ID:
			assert.Equal(t, "v1.0.0", config.MinimumVersion)
		case container2.ID:
			assert.Equal(t, "v14.0", config.MinimumVersion)
		default:
			t.Errorf("Unexpected container ID: %d", config.ContainerID)
		}
	}
}

func TestRollbackContainerConfigurationMigration(t *testing.T) {
	db := setupMigrationTestDB(t)

	// Create migrated data
	container := &Container{Name: "nginx"}
	require.NoError(t, db.Create(container).Error)

	version := &ContainerVersion{
		ContainerID:    container.ID,
		Version:        "v1.0.0",
		ComposeContent: "test",
	}
	require.NoError(t, db.Create(version).Error)

	config := &ContainerConfiguration{
		ContainerID:     container.ID,
		Name:            "default",
		MinimumVersion:  "v1.0.0",
		UISchema:        datatypes.JSON([]byte(`{}`)),
		DependencyRules: datatypes.JSON([]byte(`[]`)),
	}
	require.NoError(t, db.Create(config).Error)

	version.ConfigurationID = &config.ID
	require.NoError(t, db.Save(version).Error)

	// Rollback
	err := RollbackContainerConfigurationMigration(db)
	require.NoError(t, err)

	// Verify rollback
	var rolledBackConfig ContainerConfiguration
	err = db.First(&rolledBackConfig, config.ID).Error
	require.NoError(t, err)
	assert.Empty(t, rolledBackConfig.Name)
	assert.Empty(t, rolledBackConfig.MinimumVersion)

	var rolledBackVersion ContainerVersion
	err = db.First(&rolledBackVersion, version.ID).Error
	require.NoError(t, err)
	assert.Nil(t, rolledBackVersion.ConfigurationID)
}
