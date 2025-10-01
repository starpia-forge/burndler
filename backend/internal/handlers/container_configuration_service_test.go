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

func TestGetServiceContainerConfiguration(t *testing.T) {
	db := setupConfigTestDB(t)
	handler := NewContainerConfigurationHandler(db)

	// Setup test data
	service := &models.Service{Name: "test-service"}
	require.NoError(t, db.Create(service).Error)

	container := &models.Container{Name: "nginx"}
	require.NoError(t, db.Create(container).Error)

	version := &models.ContainerVersion{
		ContainerID: container.ID,
		Version:     "1.0.0",
	}
	require.NoError(t, db.Create(version).Error)

	// Create UI schema
	uiSchema := map[string]interface{}{
		"sections": []map[string]interface{}{
			{
				"title": "Basic Settings",
				"fields": []map[string]interface{}{
					{"name": "port", "type": "number", "label": "Port"},
				},
			},
		},
	}
	uiSchemaJSON, _ := json.Marshal(uiSchema)

	// Create dependency rules
	depRules := []map[string]interface{}{
		{
			"type":      "requires",
			"field":     "port",
			"condition": "ne 0",
		},
	}
	depRulesJSON, _ := json.Marshal(depRules)

	// Create container configuration
	containerConfig := &models.ContainerConfiguration{
		ContainerVersionID: version.ID,
		UISchema:           datatypes.JSON(uiSchemaJSON),
		DependencyRules:    datatypes.JSON(depRulesJSON),
	}
	require.NoError(t, db.Create(containerConfig).Error)

	// Create service container
	serviceContainer := &models.ServiceContainer{
		ServiceID:          service.ID,
		ContainerID:        container.ID,
		ContainerVersionID: version.ID,
	}
	require.NoError(t, db.Create(serviceContainer).Error)

	// Create service configuration (current values)
	configValues := map[string]interface{}{
		"port": 8080,
	}
	configValuesJSON, _ := json.Marshal(configValues)
	serviceConfig := &models.ServiceConfiguration{
		ServiceID:           service.ID,
		ContainerID:         container.ID,
		ConfigurationValues: configValuesJSON,
	}
	require.NoError(t, db.Create(serviceConfig).Error)

	// Test GET request
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{
		{Key: "id", Value: "1"},
		{Key: "container_id", Value: "1"},
	}

	handler.GetServiceContainerConfiguration(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response["ui_schema"])
	assert.NotNil(t, response["dependency_rules"])
	assert.NotNil(t, response["current_values"])

	currentValues := response["current_values"].(map[string]interface{})
	assert.Equal(t, float64(8080), currentValues["port"])
}

func TestGetServiceContainerConfiguration_NotFound(t *testing.T) {
	db := setupConfigTestDB(t)
	handler := NewContainerConfigurationHandler(db)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{
		{Key: "id", Value: "999"},
		{Key: "container_id", Value: "999"},
	}

	handler.GetServiceContainerConfiguration(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "NOT_FOUND", response.Error)
	assert.Equal(t, "Service container not found", response.Message)
}

func TestGetServiceContainerConfiguration_NoCurrentValues(t *testing.T) {
	db := setupConfigTestDB(t)
	handler := NewContainerConfigurationHandler(db)

	// Setup test data without service configuration
	service := &models.Service{Name: "test-service"}
	require.NoError(t, db.Create(service).Error)

	container := &models.Container{Name: "nginx"}
	require.NoError(t, db.Create(container).Error)

	version := &models.ContainerVersion{
		ContainerID: container.ID,
		Version:     "1.0.0",
	}
	require.NoError(t, db.Create(version).Error)

	uiSchema := map[string]interface{}{"sections": []interface{}{}}
	uiSchemaJSON, _ := json.Marshal(uiSchema)

	containerConfig := &models.ContainerConfiguration{
		ContainerVersionID: version.ID,
		UISchema:           datatypes.JSON(uiSchemaJSON),
	}
	require.NoError(t, db.Create(containerConfig).Error)

	serviceContainer := &models.ServiceContainer{
		ServiceID:          service.ID,
		ContainerID:        container.ID,
		ContainerVersionID: version.ID,
	}
	require.NoError(t, db.Create(serviceContainer).Error)

	// Test GET request
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{
		{Key: "id", Value: "1"},
		{Key: "container_id", Value: "1"},
	}

	handler.GetServiceContainerConfiguration(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	currentValues := response["current_values"].(map[string]interface{})
	assert.Empty(t, currentValues)
}

func TestSaveServiceContainerConfiguration_Create(t *testing.T) {
	db := setupConfigTestDB(t)
	handler := NewContainerConfigurationHandler(db)

	// Setup test data
	service := &models.Service{Name: "test-service"}
	require.NoError(t, db.Create(service).Error)

	container := &models.Container{Name: "nginx"}
	require.NoError(t, db.Create(container).Error)

	version := &models.ContainerVersion{
		ContainerID: container.ID,
		Version:     "1.0.0",
	}
	require.NoError(t, db.Create(version).Error)

	// Create container configuration without dependency rules
	containerConfig := &models.ContainerConfiguration{
		ContainerVersionID: version.ID,
		UISchema:           datatypes.JSON([]byte(`{"sections":[]}`)),
		DependencyRules:    datatypes.JSON([]byte(`[]`)),
	}
	require.NoError(t, db.Create(containerConfig).Error)

	serviceContainer := &models.ServiceContainer{
		ServiceID:          service.ID,
		ContainerID:        container.ID,
		ContainerVersionID: version.ID,
	}
	require.NoError(t, db.Create(serviceContainer).Error)

	// Prepare request
	requestData := map[string]interface{}{
		"configuration_values": map[string]interface{}{
			"port":   8080,
			"domain": "example.com",
		},
	}
	requestJSON, _ := json.Marshal(requestData)

	// Test PUT request
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{
		{Key: "id", Value: "1"},
		{Key: "container_id", Value: "1"},
	}
	c.Request = httptest.NewRequest("PUT", "/", bytes.NewBuffer(requestJSON))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.SaveServiceContainerConfiguration(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Configuration saved successfully", response["message"])

	// Verify database
	var savedConfig models.ServiceConfiguration
	err = db.Where("service_id = ? AND container_id = ?", service.ID, container.ID).
		First(&savedConfig).Error
	require.NoError(t, err)

	var savedValues map[string]interface{}
	err = json.Unmarshal(savedConfig.ConfigurationValues, &savedValues)
	require.NoError(t, err)

	assert.Equal(t, float64(8080), savedValues["port"])
	assert.Equal(t, "example.com", savedValues["domain"])
}

func TestSaveServiceContainerConfiguration_Update(t *testing.T) {
	db := setupConfigTestDB(t)
	handler := NewContainerConfigurationHandler(db)

	// Setup test data
	service := &models.Service{Name: "test-service"}
	require.NoError(t, db.Create(service).Error)

	container := &models.Container{Name: "nginx"}
	require.NoError(t, db.Create(container).Error)

	version := &models.ContainerVersion{
		ContainerID: container.ID,
		Version:     "1.0.0",
	}
	require.NoError(t, db.Create(version).Error)

	containerConfig := &models.ContainerConfiguration{
		ContainerVersionID: version.ID,
		UISchema:           datatypes.JSON([]byte(`{"sections":[]}`)),
		DependencyRules:    datatypes.JSON([]byte(`[]`)),
	}
	require.NoError(t, db.Create(containerConfig).Error)

	serviceContainer := &models.ServiceContainer{
		ServiceID:          service.ID,
		ContainerID:        container.ID,
		ContainerVersionID: version.ID,
	}
	require.NoError(t, db.Create(serviceContainer).Error)

	// Create existing configuration
	existingValues := map[string]interface{}{"port": 80}
	existingValuesJSON, _ := json.Marshal(existingValues)
	existingConfig := &models.ServiceConfiguration{
		ServiceID:           service.ID,
		ContainerID:         container.ID,
		ConfigurationValues: existingValuesJSON,
	}
	require.NoError(t, db.Create(existingConfig).Error)

	// Prepare update request
	requestData := map[string]interface{}{
		"configuration_values": map[string]interface{}{
			"port":   8080,
			"domain": "updated.com",
		},
	}
	requestJSON, _ := json.Marshal(requestData)

	// Test PUT request
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{
		{Key: "id", Value: "1"},
		{Key: "container_id", Value: "1"},
	}
	c.Request = httptest.NewRequest("PUT", "/", bytes.NewBuffer(requestJSON))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.SaveServiceContainerConfiguration(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Configuration updated successfully", response["message"])

	// Verify database update
	var updatedConfig models.ServiceConfiguration
	err = db.Where("service_id = ? AND container_id = ?", service.ID, container.ID).
		First(&updatedConfig).Error
	require.NoError(t, err)

	var updatedValues map[string]interface{}
	err = json.Unmarshal(updatedConfig.ConfigurationValues, &updatedValues)
	require.NoError(t, err)

	assert.Equal(t, float64(8080), updatedValues["port"])
	assert.Equal(t, "updated.com", updatedValues["domain"])
}

func TestSaveServiceContainerConfiguration_ValidationError(t *testing.T) {
	db := setupConfigTestDB(t)
	handler := NewContainerConfigurationHandler(db)

	// Setup test data
	service := &models.Service{Name: "test-service"}
	require.NoError(t, db.Create(service).Error)

	container := &models.Container{Name: "nginx"}
	require.NoError(t, db.Create(container).Error)

	version := &models.ContainerVersion{
		ContainerID: container.ID,
		Version:     "1.0.0",
	}
	require.NoError(t, db.Create(version).Error)

	// Create container configuration with dependency rules that will fail
	// Rule: if ssl_enabled is true, ssl_cert must be set
	depRules := []map[string]interface{}{
		{
			"type":      "requires",
			"field":     "ssl_enabled",
			"condition": "eq true",
			"target":    "ssl_cert",
			"message":   "SSL certificate is required when SSL is enabled",
		},
	}
	depRulesJSON, _ := json.Marshal(depRules)

	containerConfig := &models.ContainerConfiguration{
		ContainerVersionID: version.ID,
		UISchema:           datatypes.JSON([]byte(`{"sections":[]}`)),
		DependencyRules:    datatypes.JSON(depRulesJSON),
	}
	require.NoError(t, db.Create(containerConfig).Error)

	serviceContainer := &models.ServiceContainer{
		ServiceID:          service.ID,
		ContainerID:        container.ID,
		ContainerVersionID: version.ID,
	}
	require.NoError(t, db.Create(serviceContainer).Error)

	// Prepare request with invalid data (ssl_enabled=true but no ssl_cert)
	requestData := map[string]interface{}{
		"configuration_values": map[string]interface{}{
			"ssl_enabled": true,
			// Missing ssl_cert - should trigger validation error
		},
	}
	requestJSON, _ := json.Marshal(requestData)

	// Test PUT request
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{
		{Key: "id", Value: "1"},
		{Key: "container_id", Value: "1"},
	}
	c.Request = httptest.NewRequest("PUT", "/", bytes.NewBuffer(requestJSON))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.SaveServiceContainerConfiguration(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "VALIDATION_ERROR", response.Error)
	assert.Equal(t, "Validation failed", response.Message)
	assert.NotNil(t, response.Details)
	assert.Contains(t, response.Details, "ssl_enabled")
}

func TestSaveServiceContainerConfiguration_ServiceNotFound(t *testing.T) {
	db := setupConfigTestDB(t)
	handler := NewContainerConfigurationHandler(db)

	requestData := map[string]interface{}{
		"configuration_values": map[string]interface{}{
			"port": 8080,
		},
	}
	requestJSON, _ := json.Marshal(requestData)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{
		{Key: "id", Value: "999"},
		{Key: "container_id", Value: "1"},
	}
	c.Request = httptest.NewRequest("PUT", "/", bytes.NewBuffer(requestJSON))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.SaveServiceContainerConfiguration(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "NOT_FOUND", response.Error)
	assert.Equal(t, "Service not found", response.Message)
}
