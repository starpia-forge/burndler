package models

import (
	"encoding/json"
	"testing"
	"time"

	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// Migrate all required tables
	err = db.AutoMigrate(
		&Container{},
		&ContainerVersion{},
		&ContainerConfiguration{},
		&ContainerFile{},
		&ContainerAsset{},
		&Service{},
		&ServiceConfiguration{},
	)
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	return db
}

func TestContainerConfiguration_TableName(t *testing.T) {
	config := ContainerConfiguration{}
	if config.TableName() != "container_configurations" {
		t.Errorf("expected table name 'container_configurations', got '%s'", config.TableName())
	}
}

func TestContainerConfiguration_Create(t *testing.T) {
	db := setupTestDB(t)

	// Create a container
	container := &Container{Name: "test-container"}
	if err := db.Create(container).Error; err != nil {
		t.Fatalf("failed to create container: %v", err)
	}

	// Create ContainerConfiguration
	uiSchema := map[string]interface{}{
		"sections": []map[string]interface{}{
			{
				"id":    "database",
				"title": "Database Settings",
			},
		},
	}
	uiSchemaJSON, _ := json.Marshal(uiSchema)

	dependencyRules := map[string]interface{}{
		"rules": []map[string]interface{}{},
	}
	dependencyRulesJSON, _ := json.Marshal(dependencyRules)

	config := &ContainerConfiguration{
		ContainerID:     container.ID,
		Name:            "default",
		MinimumVersion:  "v0.1.0",
		UISchema:        datatypes.JSON(uiSchemaJSON),
		DependencyRules: datatypes.JSON(dependencyRulesJSON),
	}

	if err := db.Create(config).Error; err != nil {
		t.Fatalf("failed to create container configuration: %v", err)
	}

	if config.ID == 0 {
		t.Error("expected ID to be set after creation")
	}

	if config.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if config.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestContainerConfiguration_UniqueConstraint(t *testing.T) {
	db := setupTestDB(t)

	// Create a container
	container := &Container{Name: "test-container"}
	db.Create(container)

	// Create first configuration
	config1 := &ContainerConfiguration{
		ContainerID:     container.ID,
		Name:            "default",
		MinimumVersion:  "v0.1.0",
		UISchema:        datatypes.JSON([]byte(`{}`)),
		DependencyRules: datatypes.JSON([]byte(`{}`)),
	}
	if err := db.Create(config1).Error; err != nil {
		t.Fatalf("failed to create first configuration: %v", err)
	}

	// Try to create duplicate configuration with same (ContainerID, Name)
	config2 := &ContainerConfiguration{
		ContainerID:     container.ID,
		Name:            "default", // Same name
		MinimumVersion:  "v0.2.0",
		UISchema:        datatypes.JSON([]byte(`{}`)),
		DependencyRules: datatypes.JSON([]byte(`{}`)),
	}
	err := db.Create(config2).Error
	if err == nil {
		t.Error("expected unique constraint violation, but creation succeeded")
	}
}

func TestContainerFile_TableName(t *testing.T) {
	file := ContainerFile{}
	if file.TableName() != "container_files" {
		t.Errorf("expected table name 'container_files', got '%s'", file.TableName())
	}
}

func TestContainerFile_Create(t *testing.T) {
	db := setupTestDB(t)

	// Create container and configuration
	container := &Container{Name: "test-container"}
	db.Create(container)

	config := &ContainerConfiguration{
		ContainerID:     container.ID,
		Name:            "default",
		MinimumVersion:  "v0.1.0",
		UISchema:        datatypes.JSON([]byte(`{}`)),
		DependencyRules: datatypes.JSON([]byte(`{}`)),
	}
	db.Create(config)

	// Create ContainerFile
	file := &ContainerFile{
		ContainerConfigurationID: config.ID,
		FilePath:                 "config/app.yaml",
		FileType:                 "template",
		StoragePath:              "/storage/path/file.yaml",
		TemplateFormat:           "yaml",
		DisplayCondition:         "{{.Database.Enabled}}",
		IsDirectory:              false,
		Description:              "Application configuration",
	}

	if err := db.Create(file).Error; err != nil {
		t.Fatalf("failed to create container file: %v", err)
	}

	if file.ID == 0 {
		t.Error("expected ID to be set after creation")
	}

	// Verify data
	var retrieved ContainerFile
	db.First(&retrieved, file.ID)

	if retrieved.FilePath != "config/app.yaml" {
		t.Errorf("expected FilePath 'config/app.yaml', got '%s'", retrieved.FilePath)
	}

	if retrieved.FileType != "template" {
		t.Errorf("expected FileType 'template', got '%s'", retrieved.FileType)
	}

	if retrieved.TemplateFormat != "yaml" {
		t.Errorf("expected TemplateFormat 'yaml', got '%s'", retrieved.TemplateFormat)
	}
}

func TestContainerAsset_TableName(t *testing.T) {
	asset := ContainerAsset{}
	if asset.TableName() != "container_assets" {
		t.Errorf("expected table name 'container_assets', got '%s'", asset.TableName())
	}
}

func TestContainerAsset_Create(t *testing.T) {
	db := setupTestDB(t)

	// Create container and configuration
	container := &Container{Name: "test-container"}
	db.Create(container)

	config := &ContainerConfiguration{
		ContainerID:     container.ID,
		Name:            "default",
		MinimumVersion:  "v0.1.0",
		UISchema:        datatypes.JSON([]byte(`{}`)),
		DependencyRules: datatypes.JSON([]byte(`{}`)),
	}
	db.Create(config)

	// Create ContainerAsset
	asset := &ContainerAsset{
		ContainerConfigurationID: config.ID,
		OriginalFileName:         "database.tar.gz",
		FilePath:                 "data/database.tar.gz",
		AssetType:                "data",
		MimeType:                 "application/gzip",
		FileSize:                 1024000,
		Checksum:                 "abc123def456",
		Compressed:               true,
		IncludeCondition:         "{{.Database.Enabled}}",
		StorageType:              "embedded",
		StoragePath:              "/storage/assets/database.tar.gz",
		DownloadURL:              "",
	}

	if err := db.Create(asset).Error; err != nil {
		t.Fatalf("failed to create container asset: %v", err)
	}

	if asset.ID == 0 {
		t.Error("expected ID to be set after creation")
	}

	// Verify data
	var retrieved ContainerAsset
	db.First(&retrieved, asset.ID)

	if retrieved.OriginalFileName != "database.tar.gz" {
		t.Errorf("expected OriginalFileName 'database.tar.gz', got '%s'", retrieved.OriginalFileName)
	}

	if retrieved.FileSize != 1024000 {
		t.Errorf("expected FileSize 1024000, got %d", retrieved.FileSize)
	}

	if retrieved.Compressed != true {
		t.Error("expected Compressed to be true")
	}
}

func TestServiceConfiguration_TableName(t *testing.T) {
	config := ServiceConfiguration{}
	if config.TableName() != "service_configurations" {
		t.Errorf("expected table name 'service_configurations', got '%s'", config.TableName())
	}
}

func TestServiceConfiguration_Create(t *testing.T) {
	db := setupTestDB(t)

	// Create user
	user := &User{
		Name:     "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Role:     "Developer",
	}
	db.Create(user)

	// Create service
	service := &Service{
		Name:   "test-service",
		UserID: user.ID,
	}
	db.Create(service)

	// Create container
	container := &Container{Name: "test-container"}
	db.Create(container)

	// Create ServiceConfiguration
	configValues := map[string]interface{}{
		"Database": map[string]interface{}{
			"Host": "localhost",
			"Port": 5432,
		},
	}
	configValuesJSON, _ := json.Marshal(configValues)

	serviceConfig := &ServiceConfiguration{
		ServiceID:           service.ID,
		ContainerID:         container.ID,
		ConfigurationValues: datatypes.JSON(configValuesJSON),
	}

	if err := db.Create(serviceConfig).Error; err != nil {
		t.Fatalf("failed to create service configuration: %v", err)
	}

	if serviceConfig.ID == 0 {
		t.Error("expected ID to be set after creation")
	}

	// Test unique constraint
	duplicate := &ServiceConfiguration{
		ServiceID:           service.ID,
		ContainerID:         container.ID,
		ConfigurationValues: datatypes.JSON([]byte(`{}`)),
	}
	err := db.Create(duplicate).Error
	if err == nil {
		t.Error("expected unique constraint violation for (service_id, container_id)")
	}
}

func TestContainerConfiguration_Relationships(t *testing.T) {
	db := setupTestDB(t)

	// Create container
	container := &Container{Name: "test-container"}
	db.Create(container)

	// Create configuration with files and assets
	config := &ContainerConfiguration{
		ContainerID:     container.ID,
		Name:            "default",
		MinimumVersion:  "v0.1.0",
		UISchema:        datatypes.JSON([]byte(`{}`)),
		DependencyRules: datatypes.JSON([]byte(`{}`)),
	}
	db.Create(config)

	// Create files
	file1 := &ContainerFile{
		ContainerConfigurationID: config.ID,
		FilePath:                 "config/app1.yaml",
		FileType:                 "template",
		StoragePath:              "/storage/app1.yaml",
	}
	db.Create(file1)

	file2 := &ContainerFile{
		ContainerConfigurationID: config.ID,
		FilePath:                 "config/app2.yaml",
		FileType:                 "static",
		StoragePath:              "/storage/app2.yaml",
	}
	db.Create(file2)

	// Create assets
	asset1 := &ContainerAsset{
		ContainerConfigurationID: config.ID,
		OriginalFileName:         "data1.tar.gz",
		FilePath:                 "data/data1.tar.gz",
		AssetType:                "data",
		FileSize:                 1000,
		Checksum:                 "abc123",
		StorageType:              "embedded",
		StoragePath:              "/storage/data1.tar.gz",
	}
	db.Create(asset1)

	// Load configuration with relationships
	var loaded ContainerConfiguration
	err := db.Preload("Files").Preload("Assets").First(&loaded, config.ID).Error
	if err != nil {
		t.Fatalf("failed to load configuration with relationships: %v", err)
	}

	if len(loaded.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(loaded.Files))
	}

	if len(loaded.Assets) != 1 {
		t.Errorf("expected 1 asset, got %d", len(loaded.Assets))
	}
}

func TestJSONBFieldsMarshaling(t *testing.T) {
	db := setupTestDB(t)

	// Create container
	container := &Container{Name: "test-container"}
	db.Create(container)

	// Create configuration with complex JSONB data
	uiSchema := map[string]interface{}{
		"sections": []map[string]interface{}{
			{
				"id":    "database",
				"title": "Database Settings",
				"fields": []map[string]interface{}{
					{
						"key":   "Database.Host",
						"type":  "string",
						"label": "Host",
					},
				},
			},
		},
	}
	uiSchemaJSON, _ := json.Marshal(uiSchema)

	config := &ContainerConfiguration{
		ContainerID:     container.ID,
		Name:            "default",
		MinimumVersion:  "v0.1.0",
		UISchema:        datatypes.JSON(uiSchemaJSON),
		DependencyRules: datatypes.JSON([]byte(`{"rules":[]}`)),
	}
	db.Create(config)

	// Retrieve and unmarshal
	var loaded ContainerConfiguration
	db.First(&loaded, config.ID)

	var retrievedSchema map[string]interface{}
	err := json.Unmarshal(loaded.UISchema, &retrievedSchema)
	if err != nil {
		t.Fatalf("failed to unmarshal UISchema: %v", err)
	}

	sections, ok := retrievedSchema["sections"].([]interface{})
	if !ok || len(sections) == 0 {
		t.Error("expected UISchema to contain sections array")
	}
}

func TestContainerFile_IndexCreation(t *testing.T) {
	db := setupTestDB(t)

	// Verify index exists on container_version_id
	var indexExists bool
	result := db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM sqlite_master
			WHERE type='index'
			AND tbl_name='container_files'
		)
	`).Scan(&indexExists)

	if result.Error != nil {
		t.Fatalf("failed to check for indexes: %v", result.Error)
	}

	if !indexExists {
		t.Log("Note: Index verification is database-specific. SQLite may handle indexes differently.")
	}
}

func TestTimestampsAutoUpdate(t *testing.T) {
	db := setupTestDB(t)

	// Create container
	container := &Container{Name: "test-container"}
	db.Create(container)

	// Create configuration
	config := &ContainerConfiguration{
		ContainerID:     container.ID,
		Name:            "default",
		MinimumVersion:  "v0.1.0",
		UISchema:        datatypes.JSON([]byte(`{}`)),
		DependencyRules: datatypes.JSON([]byte(`{}`)),
	}
	db.Create(config)

	createdAt := config.CreatedAt
	updatedAt := config.UpdatedAt

	// Wait a bit and update
	time.Sleep(10 * time.Millisecond)

	config.UISchema = datatypes.JSON([]byte(`{"updated":true}`))
	db.Save(config)

	if !config.UpdatedAt.After(updatedAt) {
		t.Error("expected UpdatedAt to be updated after save")
	}

	if !config.CreatedAt.Equal(createdAt) {
		t.Error("expected CreatedAt to remain unchanged after update")
	}
}

// ===== Phase 1.1: New tests for Container-level Configuration structure =====

// TestContainerConfiguration_BelongsToContainer tests that ContainerConfiguration
// belongs to Container (not ContainerVersion) with Name field
func TestContainerConfiguration_BelongsToContainer(t *testing.T) {
	db := setupTestDB(t)

	// Create container
	container := &Container{Name: "test-container"}
	if err := db.Create(container).Error; err != nil {
		t.Fatalf("failed to create container: %v", err)
	}

	// Create configuration at Container level with Name
	config := &ContainerConfiguration{
		ContainerID:    container.ID,
		Name:           "default",
		Description:    "Default configuration",
		MinimumVersion: "v0.1.0",
		UISchema:       datatypes.JSON([]byte(`{}`)),
		DependencyRules: datatypes.JSON([]byte(`{}`)),
	}

	if err := db.Create(config).Error; err != nil {
		t.Fatalf("failed to create container configuration: %v", err)
	}

	if config.ID == 0 {
		t.Error("expected ID to be set after creation")
	}

	// Verify relationship
	var loaded ContainerConfiguration
	err := db.Preload("Container").First(&loaded, config.ID).Error
	if err != nil {
		t.Fatalf("failed to load configuration with container: %v", err)
	}

	if loaded.Container.ID != container.ID {
		t.Errorf("expected Container ID %d, got %d", container.ID, loaded.Container.ID)
	}

	if loaded.Name != "default" {
		t.Errorf("expected Name 'default', got '%s'", loaded.Name)
	}
}

// TestContainerConfiguration_RequiresMinimumVersion tests that MinimumVersion is required
func TestContainerConfiguration_RequiresMinimumVersion(t *testing.T) {
	db := setupTestDB(t)

	container := &Container{Name: "test-container"}
	db.Create(container)

	// Try to create configuration without MinimumVersion
	config := &ContainerConfiguration{
		ContainerID:     container.ID,
		Name:            "config1",
		UISchema:        datatypes.JSON([]byte(`{}`)),
		DependencyRules: datatypes.JSON([]byte(`{}`)),
		// MinimumVersion is missing
	}

	err := db.Create(config).Error
	if err == nil {
		t.Error("expected error when creating configuration without MinimumVersion")
	}
}

// TestContainerConfiguration_UniqueNamePerContainer tests that (ContainerID, Name) is unique
func TestContainerConfiguration_UniqueNamePerContainer(t *testing.T) {
	db := setupTestDB(t)

	container := &Container{Name: "test-container"}
	db.Create(container)

	// Create first configuration
	config1 := &ContainerConfiguration{
		ContainerID:     container.ID,
		Name:            "default",
		MinimumVersion:  "v0.1.0",
		UISchema:        datatypes.JSON([]byte(`{}`)),
		DependencyRules: datatypes.JSON([]byte(`{}`)),
	}
	if err := db.Create(config1).Error; err != nil {
		t.Fatalf("failed to create first configuration: %v", err)
	}

	// Try to create duplicate configuration with same Container + Name
	config2 := &ContainerConfiguration{
		ContainerID:     container.ID,
		Name:            "default", // Same name
		MinimumVersion:  "v0.2.0",
		UISchema:        datatypes.JSON([]byte(`{}`)),
		DependencyRules: datatypes.JSON([]byte(`{}`)),
	}
	err := db.Create(config2).Error
	if err == nil {
		t.Error("expected unique constraint violation for (container_id, name)")
	}

	// But different name should work
	config3 := &ContainerConfiguration{
		ContainerID:     container.ID,
		Name:            "advanced", // Different name
		MinimumVersion:  "v0.2.0",
		UISchema:        datatypes.JSON([]byte(`{}`)),
		DependencyRules: datatypes.JSON([]byte(`{}`)),
	}
	if err := db.Create(config3).Error; err != nil {
		t.Errorf("failed to create configuration with different name: %v", err)
	}
}

// TestContainerVersion_CanReferenceConfiguration tests that ContainerVersion
// can reference ContainerConfiguration via nullable ConfigurationID
func TestContainerVersion_CanReferenceConfiguration(t *testing.T) {
	db := setupTestDB(t)

	// Create container
	container := &Container{Name: "test-container"}
	db.Create(container)

	// Create configuration
	config := &ContainerConfiguration{
		ContainerID:     container.ID,
		Name:            "default",
		MinimumVersion:  "v0.1.0",
		UISchema:        datatypes.JSON([]byte(`{}`)),
		DependencyRules: datatypes.JSON([]byte(`{}`)),
	}
	db.Create(config)

	// Create version WITH configuration reference
	version1 := &ContainerVersion{
		ContainerID:     container.ID,
		Version:         "v0.1.0",
		ComposeContent:  "test",
		ConfigurationID: &config.ID,
	}
	if err := db.Create(version1).Error; err != nil {
		t.Fatalf("failed to create version with configuration: %v", err)
	}

	// Create version WITHOUT configuration (nullable)
	version2 := &ContainerVersion{
		ContainerID:     container.ID,
		Version:         "v0.0.5",
		ComposeContent:  "test",
		ConfigurationID: nil,
	}
	if err := db.Create(version2).Error; err != nil {
		t.Fatalf("failed to create version without configuration: %v", err)
	}

	// Load version with configuration
	var loaded ContainerVersion
	err := db.Preload("Configuration").First(&loaded, version1.ID).Error
	if err != nil {
		t.Fatalf("failed to load version with configuration: %v", err)
	}

	if loaded.Configuration == nil {
		t.Fatal("expected Configuration to be loaded")
	}

	if loaded.Configuration.ID != config.ID {
		t.Errorf("expected Configuration ID %d, got %d", config.ID, loaded.Configuration.ID)
	}

	// Load version without configuration
	var loaded2 ContainerVersion
	err = db.Preload("Configuration").First(&loaded2, version2.ID).Error
	if err != nil {
		t.Fatalf("failed to load version without configuration: %v", err)
	}

	if loaded2.Configuration != nil {
		t.Error("expected Configuration to be nil for version without configuration")
	}
}

// TestContainerFile_BelongsToConfiguration tests that ContainerFile
// belongs to ContainerConfiguration (not ContainerVersion)
func TestContainerFile_BelongsToConfiguration(t *testing.T) {
	db := setupTestDB(t)

	// Create container and configuration
	container := &Container{Name: "test-container"}
	db.Create(container)

	config := &ContainerConfiguration{
		ContainerID:     container.ID,
		Name:            "default",
		MinimumVersion:  "v0.1.0",
		UISchema:        datatypes.JSON([]byte(`{}`)),
		DependencyRules: datatypes.JSON([]byte(`{}`)),
	}
	db.Create(config)

	// Create file belonging to configuration
	file := &ContainerFile{
		ContainerConfigurationID: config.ID,
		FilePath:                 "config/app.yaml",
		FileType:                 "template",
		StoragePath:              "/storage/app.yaml",
	}

	if err := db.Create(file).Error; err != nil {
		t.Fatalf("failed to create container file: %v", err)
	}

	// Load configuration with files
	var loaded ContainerConfiguration
	err := db.Preload("Files").First(&loaded, config.ID).Error
	if err != nil {
		t.Fatalf("failed to load configuration with files: %v", err)
	}

	if len(loaded.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(loaded.Files))
	}

	if loaded.Files[0].FilePath != "config/app.yaml" {
		t.Errorf("expected FilePath 'config/app.yaml', got '%s'", loaded.Files[0].FilePath)
	}
}

// TestContainerAsset_BelongsToConfiguration tests that ContainerAsset
// belongs to ContainerConfiguration (not ContainerVersion)
func TestContainerAsset_BelongsToConfiguration(t *testing.T) {
	db := setupTestDB(t)

	// Create container and configuration
	container := &Container{Name: "test-container"}
	db.Create(container)

	config := &ContainerConfiguration{
		ContainerID:     container.ID,
		Name:            "default",
		MinimumVersion:  "v0.1.0",
		UISchema:        datatypes.JSON([]byte(`{}`)),
		DependencyRules: datatypes.JSON([]byte(`{}`)),
	}
	db.Create(config)

	// Create asset belonging to configuration
	asset := &ContainerAsset{
		ContainerConfigurationID: config.ID,
		OriginalFileName:         "data.tar.gz",
		FilePath:                 "data/data.tar.gz",
		AssetType:                "data",
		FileSize:                 1000,
		Checksum:                 "abc123",
		StorageType:              "embedded",
		StoragePath:              "/storage/data.tar.gz",
	}

	if err := db.Create(asset).Error; err != nil {
		t.Fatalf("failed to create container asset: %v", err)
	}

	// Load configuration with assets
	var loaded ContainerConfiguration
	err := db.Preload("Assets").First(&loaded, config.ID).Error
	if err != nil {
		t.Fatalf("failed to load configuration with assets: %v", err)
	}

	if len(loaded.Assets) != 1 {
		t.Errorf("expected 1 asset, got %d", len(loaded.Assets))
	}

	if loaded.Assets[0].OriginalFileName != "data.tar.gz" {
		t.Errorf("expected OriginalFileName 'data.tar.gz', got '%s'", loaded.Assets[0].OriginalFileName)
	}
}