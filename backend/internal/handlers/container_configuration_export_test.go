package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/burndler/burndler/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestExportServiceConfiguration(t *testing.T) {
	db := setupConfigTestDB(t)
	handler := NewContainerConfigurationHandler(db)

	// Setup test data
	service := &models.Service{
		Name: "test-service",
	}
	require.NoError(t, db.Create(service).Error)

	container1 := &models.Container{
		Name: "nginx",
	}
	container2 := &models.Container{
		Name: "postgres",
	}
	require.NoError(t, db.Create(container1).Error)
	require.NoError(t, db.Create(container2).Error)

	config1Values := map[string]interface{}{
		"port":   8080,
		"domain": "example.com",
	}
	config1JSON, _ := json.Marshal(config1Values)

	config2Values := map[string]interface{}{
		"host": "localhost",
		"port": 5432,
	}
	config2JSON, _ := json.Marshal(config2Values)

	config1 := &models.ServiceConfiguration{
		ServiceID:           service.ID,
		ContainerID:         container1.ID,
		ConfigurationValues: config1JSON,
	}
	config2 := &models.ServiceConfiguration{
		ServiceID:           service.ID,
		ContainerID:         container2.ID,
		ConfigurationValues: config2JSON,
	}
	require.NoError(t, db.Create(config1).Error)
	require.NoError(t, db.Create(config2).Error)

	// Test export
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{
		{Key: "id", Value: "1"},
	}

	handler.ExportServiceConfiguration(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "1.0", response["version"])
	assert.Equal(t, float64(1), response["service_id"])

	containers, ok := response["containers"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, containers, "nginx")
	assert.Contains(t, containers, "postgres")

	nginxConfig, ok := containers["nginx"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(8080), nginxConfig["port"])
	assert.Equal(t, "example.com", nginxConfig["domain"])
}

func TestExportServiceConfiguration_EmptyService(t *testing.T) {
	db := setupConfigTestDB(t)
	handler := NewContainerConfigurationHandler(db)

	// Create service with no configurations
	service := &models.Service{
		Name: "empty-service",
	}
	require.NoError(t, db.Create(service).Error)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{
		{Key: "id", Value: "1"},
	}

	handler.ExportServiceConfiguration(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	containers, ok := response["containers"].(map[string]interface{})
	require.True(t, ok)
	assert.Empty(t, containers)
}

func TestImportServiceConfiguration(t *testing.T) {
	db := setupConfigTestDB(t)
	handler := NewContainerConfigurationHandler(db)

	// Setup test data
	service := &models.Service{
		Name: "test-service",
	}
	require.NoError(t, db.Create(service).Error)

	container1 := &models.Container{
		Name: "nginx",
	}
	container2 := &models.Container{
		Name: "postgres",
	}
	require.NoError(t, db.Create(container1).Error)
	require.NoError(t, db.Create(container2).Error)

	// Import data
	importData := map[string]interface{}{
		"version": "1.0",
		"containers": map[string]interface{}{
			"nginx": map[string]interface{}{
				"port":   8080,
				"domain": "example.com",
			},
			"postgres": map[string]interface{}{
				"host": "localhost",
				"port": 5432,
			},
		},
	}

	importJSON, _ := json.Marshal(importData)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{
		{Key: "id", Value: "1"},
	}
	c.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(importJSON))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ImportServiceConfiguration(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Configuration imported successfully", response["message"])
	assert.Equal(t, float64(2), response["imported"])

	// Verify configurations were created
	var configs []models.ServiceConfiguration
	err = db.Where("service_id = ?", service.ID).Find(&configs).Error
	require.NoError(t, err)
	assert.Len(t, configs, 2)
}

func TestImportServiceConfiguration_UnsupportedVersion(t *testing.T) {
	db := setupConfigTestDB(t)
	handler := NewContainerConfigurationHandler(db)

	service := &models.Service{
		Name: "test-service",
	}
	require.NoError(t, db.Create(service).Error)

	importData := map[string]interface{}{
		"version":    "2.0",
		"containers": map[string]interface{}{},
	}

	importJSON, _ := json.Marshal(importData)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{
		{Key: "id", Value: "1"},
	}
	c.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(importJSON))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ImportServiceConfiguration(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Unsupported export version", response["error"])
}

func TestImportServiceConfiguration_UnknownContainers(t *testing.T) {
	db := setupConfigTestDB(t)
	handler := NewContainerConfigurationHandler(db)

	service := &models.Service{
		Name: "test-service",
	}
	require.NoError(t, db.Create(service).Error)

	container := &models.Container{
		Name: "nginx",
	}
	require.NoError(t, db.Create(container).Error)

	// Import with one known and one unknown container
	importData := map[string]interface{}{
		"version": "1.0",
		"containers": map[string]interface{}{
			"nginx": map[string]interface{}{
				"port": 8080,
			},
			"unknown-container": map[string]interface{}{
				"some": "value",
			},
		},
	}

	importJSON, _ := json.Marshal(importData)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{
		{Key: "id", Value: "1"},
	}
	c.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(importJSON))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ImportServiceConfiguration(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Configuration imported successfully", response["message"])
	assert.Equal(t, float64(1), response["imported"])

	skipped, ok := response["skipped"].([]interface{})
	require.True(t, ok)
	assert.Contains(t, skipped, "unknown-container")
}

func TestImportServiceConfiguration_UpdateExisting(t *testing.T) {
	db := setupConfigTestDB(t)
	handler := NewContainerConfigurationHandler(db)

	service := &models.Service{
		Name: "test-service",
	}
	require.NoError(t, db.Create(service).Error)

	container := &models.Container{
		Name: "nginx",
	}
	require.NoError(t, db.Create(container).Error)

	// Create existing configuration
	existingValues := map[string]interface{}{
		"port": 80,
	}
	existingJSON, _ := json.Marshal(existingValues)

	existingConfig := &models.ServiceConfiguration{
		ServiceID:           service.ID,
		ContainerID:         container.ID,
		ConfigurationValues: datatypes.JSON(existingJSON),
	}
	require.NoError(t, db.Create(existingConfig).Error)

	// Import updated configuration
	importData := map[string]interface{}{
		"version": "1.0",
		"containers": map[string]interface{}{
			"nginx": map[string]interface{}{
				"port":   8080,
				"domain": "example.com",
			},
		},
	}

	importJSON, _ := json.Marshal(importData)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{
		{Key: "id", Value: "1"},
	}
	c.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(importJSON))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ImportServiceConfiguration(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify configuration was updated
	var updatedConfig models.ServiceConfiguration
	err := db.Where("service_id = ? AND container_id = ?", service.ID, container.ID).First(&updatedConfig).Error
	require.NoError(t, err)

	var values map[string]interface{}
	err = json.Unmarshal(updatedConfig.ConfigurationValues, &values)
	require.NoError(t, err)

	assert.Equal(t, float64(8080), values["port"])
	assert.Equal(t, "example.com", values["domain"])
}
