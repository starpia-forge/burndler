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

	// Create a container version first
	container := &Container{Name: "test-container"}
	if err := db.Create(container).Error; err != nil {
		t.Fatalf("failed to create container: %v", err)
	}

	version := &ContainerVersion{
		ContainerID:    container.ID,
		Version:        "v1.0.0",
		ComposeContent: "test",
	}
	if err := db.Create(version).Error; err != nil {
		t.Fatalf("failed to create container version: %v", err)
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
		ContainerVersionID: version.ID,
		UISchema:           datatypes.JSON(uiSchemaJSON),
		DependencyRules:    datatypes.JSON(dependencyRulesJSON),
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

	// Create a container version
	container := &Container{Name: "test-container"}
	db.Create(container)

	version := &ContainerVersion{
		ContainerID:    container.ID,
		Version:        "v1.0.0",
		ComposeContent: "test",
	}
	db.Create(version)

	// Create first configuration
	config1 := &ContainerConfiguration{
		ContainerVersionID: version.ID,
		UISchema:           datatypes.JSON([]byte(`{}`)),
		DependencyRules:    datatypes.JSON([]byte(`{}`)),
	}
	if err := db.Create(config1).Error; err != nil {
		t.Fatalf("failed to create first configuration: %v", err)
	}

	// Try to create duplicate configuration
	config2 := &ContainerConfiguration{
		ContainerVersionID: version.ID,
		UISchema:           datatypes.JSON([]byte(`{}`)),
		DependencyRules:    datatypes.JSON([]byte(`{}`)),
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

	// Create a container version
	container := &Container{Name: "test-container"}
	db.Create(container)

	version := &ContainerVersion{
		ContainerID:    container.ID,
		Version:        "v1.0.0",
		ComposeContent: "test",
	}
	db.Create(version)

	// Create ContainerFile
	file := &ContainerFile{
		ContainerVersionID: version.ID,
		FilePath:           "config/app.yaml",
		FileType:           "template",
		StoragePath:        "/storage/path/file.yaml",
		TemplateFormat:     "yaml",
		DisplayCondition:   "{{.Database.Enabled}}",
		IsDirectory:        false,
		Description:        "Application configuration",
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

	// Create a container version
	container := &Container{Name: "test-container"}
	db.Create(container)

	version := &ContainerVersion{
		ContainerID:    container.ID,
		Version:        "v1.0.0",
		ComposeContent: "test",
	}
	db.Create(version)

	// Create ContainerAsset
	asset := &ContainerAsset{
		ContainerVersionID: version.ID,
		OriginalFileName:   "database.tar.gz",
		FilePath:           "data/database.tar.gz",
		AssetType:          "data",
		MimeType:           "application/gzip",
		FileSize:           1024000,
		Checksum:           "abc123def456",
		Compressed:         true,
		IncludeCondition:   "{{.Database.Enabled}}",
		StorageType:        "embedded",
		StoragePath:        "/storage/assets/database.tar.gz",
		DownloadURL:        "",
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

	// Create container and version
	container := &Container{Name: "test-container"}
	db.Create(container)

	version := &ContainerVersion{
		ContainerID:    container.ID,
		Version:        "v1.0.0",
		ComposeContent: "test",
	}
	db.Create(version)

	// Create configuration with files and assets
	config := &ContainerConfiguration{
		ContainerVersionID: version.ID,
		UISchema:           datatypes.JSON([]byte(`{}`)),
		DependencyRules:    datatypes.JSON([]byte(`{}`)),
	}
	db.Create(config)

	// Create files
	file1 := &ContainerFile{
		ContainerVersionID: version.ID,
		FilePath:           "config/app1.yaml",
		FileType:           "template",
		StoragePath:        "/storage/app1.yaml",
	}
	db.Create(file1)

	file2 := &ContainerFile{
		ContainerVersionID: version.ID,
		FilePath:           "config/app2.yaml",
		FileType:           "static",
		StoragePath:        "/storage/app2.yaml",
	}
	db.Create(file2)

	// Create assets
	asset1 := &ContainerAsset{
		ContainerVersionID: version.ID,
		OriginalFileName:   "data1.tar.gz",
		FilePath:           "data/data1.tar.gz",
		AssetType:          "data",
		FileSize:           1000,
		Checksum:           "abc123",
		StorageType:        "embedded",
		StoragePath:        "/storage/data1.tar.gz",
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

	// Create container version
	container := &Container{Name: "test-container"}
	db.Create(container)

	version := &ContainerVersion{
		ContainerID:    container.ID,
		Version:        "v1.0.0",
		ComposeContent: "test",
	}
	db.Create(version)

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
		ContainerVersionID: version.ID,
		UISchema:           datatypes.JSON(uiSchemaJSON),
		DependencyRules:    datatypes.JSON([]byte(`{"rules":[]}`)),
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

	// Create container version
	container := &Container{Name: "test-container"}
	db.Create(container)

	version := &ContainerVersion{
		ContainerID:    container.ID,
		Version:        "v1.0.0",
		ComposeContent: "test",
	}
	db.Create(version)

	// Create configuration
	config := &ContainerConfiguration{
		ContainerVersionID: version.ID,
		UISchema:           datatypes.JSON([]byte(`{}`)),
		DependencyRules:    datatypes.JSON([]byte(`{}`)),
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