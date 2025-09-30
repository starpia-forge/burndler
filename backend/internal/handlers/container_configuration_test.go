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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupConfigTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.Container{},
		&models.ContainerVersion{},
		&models.ContainerConfiguration{},
		&models.ContainerFile{},
		&models.ContainerAsset{},
	)
	require.NoError(t, err)

	return db
}

func TestContainerConfigurationHandler_CreateConfiguration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid configuration creation", func(t *testing.T) {
		db := setupConfigTestDB(t)
		handler := NewContainerConfigurationHandler(db)

		// Create container and version first
		container := &models.Container{Name: "test-container"}
		require.NoError(t, db.Create(container).Error)

		version := &models.ContainerVersion{
			ContainerID:    container.ID,
			Version:        "v1.0.0",
			ComposeContent: "test",
		}
		require.NoError(t, db.Create(version).Error)

		// Prepare request
		reqBody := map[string]interface{}{
			"ui_schema": map[string]interface{}{
				"sections": []map[string]interface{}{
					{
						"id":    "database",
						"title": "Database Settings",
					},
				},
			},
			"dependency_rules": map[string]interface{}{
				"rules": []interface{}{},
			},
		}
		body, _ := json.Marshal(reqBody)

		// Create request
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{
			{Key: "id", Value: "1"},
			{Key: "version", Value: "1"},
		}

		handler.CreateConfiguration(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.ContainerConfiguration
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.NotZero(t, response.ID)
		assert.Equal(t, version.ID, response.ContainerVersionID)
	})

	t.Run("container version not found", func(t *testing.T) {
		db := setupConfigTestDB(t)
		handler := NewContainerConfigurationHandler(db)

		reqBody := map[string]interface{}{
			"ui_schema":        map[string]interface{}{},
			"dependency_rules": map[string]interface{}{},
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{
			{Key: "id", Value: "999"},
			{Key: "version", Value: "999"},
		}

		handler.CreateConfiguration(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		db := setupConfigTestDB(t)
		handler := NewContainerConfigurationHandler(db)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{
			{Key: "id", Value: "1"},
			{Key: "version", Value: "1"},
		}

		handler.CreateConfiguration(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestContainerConfigurationHandler_GetConfiguration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("existing configuration", func(t *testing.T) {
		db := setupConfigTestDB(t)
		handler := NewContainerConfigurationHandler(db)

		// Create test data
		container := &models.Container{Name: "test-container"}
		require.NoError(t, db.Create(container).Error)

		version := &models.ContainerVersion{
			ContainerID:    container.ID,
			Version:        "v1.0.0",
			ComposeContent: "test",
		}
		require.NoError(t, db.Create(version).Error)

		uiSchema := map[string]interface{}{"fields": []interface{}{}}
		uiSchemaJSON, _ := json.Marshal(uiSchema)

		config := &models.ContainerConfiguration{
			ContainerVersionID: version.ID,
			UISchema:           datatypes.JSON(uiSchemaJSON),
			DependencyRules:    datatypes.JSON([]byte(`{"rules":[]}`)),
		}
		require.NoError(t, db.Create(config).Error)

		// Create files and assets
		file := &models.ContainerFile{
			ContainerVersionID: version.ID,
			FilePath:           "config/app.yaml",
			FileType:           "template",
			StoragePath:        "/storage/app.yaml",
		}
		require.NoError(t, db.Create(file).Error)

		asset := &models.ContainerAsset{
			ContainerVersionID: version.ID,
			OriginalFileName:   "data.tar.gz",
			FilePath:           "data/data.tar.gz",
			AssetType:          "data",
			FileSize:           1000,
			Checksum:           "abc123",
			StorageType:        "embedded",
			StoragePath:        "/storage/data.tar.gz",
		}
		require.NoError(t, db.Create(asset).Error)

		// Create request
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Params = gin.Params{
			{Key: "version", Value: "1"},
		}

		handler.GetConfiguration(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.ContainerConfiguration
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, config.ID, response.ID)
		assert.Len(t, response.Files, 1)
		assert.Len(t, response.Assets, 1)
	})

	t.Run("configuration not found", func(t *testing.T) {
		db := setupConfigTestDB(t)
		handler := NewContainerConfigurationHandler(db)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Params = gin.Params{
			{Key: "version", Value: "999"},
		}

		handler.GetConfiguration(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestContainerConfigurationHandler_UpdateConfiguration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid update", func(t *testing.T) {
		db := setupConfigTestDB(t)
		handler := NewContainerConfigurationHandler(db)

		// Create test data
		container := &models.Container{Name: "test-container"}
		require.NoError(t, db.Create(container).Error)

		version := &models.ContainerVersion{
			ContainerID:    container.ID,
			Version:        "v1.0.0",
			ComposeContent: "test",
		}
		require.NoError(t, db.Create(version).Error)

		config := &models.ContainerConfiguration{
			ContainerVersionID: version.ID,
			UISchema:           datatypes.JSON([]byte(`{"fields":[]}`)),
			DependencyRules:    datatypes.JSON([]byte(`{"rules":[]}`)),
		}
		require.NoError(t, db.Create(config).Error)

		// Prepare update request
		reqBody := map[string]interface{}{
			"ui_schema": map[string]interface{}{
				"sections": []map[string]interface{}{
					{
						"id":    "updated",
						"title": "Updated Section",
					},
				},
			},
			"dependency_rules": map[string]interface{}{
				"rules": []interface{}{},
			},
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PUT", "/", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{
			{Key: "version", Value: "1"},
		}

		handler.UpdateConfiguration(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.ContainerConfiguration
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, string(response.UISchema), "updated")
	})

	t.Run("configuration not found", func(t *testing.T) {
		db := setupConfigTestDB(t)
		handler := NewContainerConfigurationHandler(db)

		reqBody := map[string]interface{}{
			"ui_schema":        map[string]interface{}{},
			"dependency_rules": map[string]interface{}{},
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PUT", "/", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{
			{Key: "version", Value: "999"},
		}

		handler.UpdateConfiguration(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestContainerConfigurationHandler_DeleteConfiguration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful deletion", func(t *testing.T) {
		db := setupConfigTestDB(t)
		handler := NewContainerConfigurationHandler(db)

		// Create test data
		container := &models.Container{Name: "test-container"}
		require.NoError(t, db.Create(container).Error)

		version := &models.ContainerVersion{
			ContainerID:    container.ID,
			Version:        "v1.0.0",
			ComposeContent: "test",
		}
		require.NoError(t, db.Create(version).Error)

		config := &models.ContainerConfiguration{
			ContainerVersionID: version.ID,
			UISchema:           datatypes.JSON([]byte(`{}`)),
			DependencyRules:    datatypes.JSON([]byte(`{}`)),
		}
		require.NoError(t, db.Create(config).Error)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/", nil)
		c.Params = gin.Params{
			{Key: "version", Value: "1"},
		}

		handler.DeleteConfiguration(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify deletion
		var count int64
		db.Model(&models.ContainerConfiguration{}).Where("container_version_id = ?", version.ID).Count(&count)
		assert.Equal(t, int64(0), count)
	})

	t.Run("delete non-existing configuration", func(t *testing.T) {
		db := setupConfigTestDB(t)
		handler := NewContainerConfigurationHandler(db)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("DELETE", "/", nil)
		c.Params = gin.Params{
			{Key: "version", Value: "999"},
		}

		handler.DeleteConfiguration(c)

		// Should still return OK even if nothing was deleted
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func setupValidationTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.Container{},
		&models.ContainerVersion{},
		&models.ContainerConfiguration{},
		&models.Service{},
		&models.ServiceContainer{},
		&models.User{},
	)
	require.NoError(t, err)

	return db
}

func TestContainerConfigurationHandler_ValidateConfiguration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("validation passes - no errors", func(t *testing.T) {
		db := setupValidationTestDB(t)
		handler := NewContainerConfigurationHandler(db)

		// Create test data
		user := &models.User{Email: "test@example.com", Name: "Test User", Role: "Developer"}
		require.NoError(t, db.Create(user).Error)

		service := &models.Service{Name: "test-service", UserID: user.ID}
		require.NoError(t, db.Create(service).Error)

		container := &models.Container{Name: "test-container"}
		require.NoError(t, db.Create(container).Error)

		version := &models.ContainerVersion{
			ContainerID:    container.ID,
			Version:        "v1.0.0",
			ComposeContent: "test",
		}
		require.NoError(t, db.Create(version).Error)

		rules := []map[string]interface{}{
			{
				"type":      "requires",
				"field":     "SSL.Enabled",
				"condition": "{{.SSL.Enabled}} == true",
				"target":    "SSL.Certificate",
				"message":   "SSL requires certificate",
			},
		}
		rulesJSON, _ := json.Marshal(rules)

		config := &models.ContainerConfiguration{
			ContainerVersionID: version.ID,
			UISchema:           datatypes.JSON(`{}`),
			DependencyRules:    datatypes.JSON(rulesJSON),
		}
		require.NoError(t, db.Create(config).Error)

		serviceContainer := &models.ServiceContainer{
			ServiceID:          service.ID,
			ContainerID:        container.ID,
			ContainerVersionID: version.ID,
		}
		require.NoError(t, db.Create(serviceContainer).Error)

		// Prepare request - SSL disabled, so certificate not required
		reqBody := map[string]interface{}{
			"values": map[string]interface{}{
				"SSL": map[string]interface{}{
					"Enabled": false,
				},
			},
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{
			{Key: "id", Value: "1"},
			{Key: "container_id", Value: "1"},
		}

		handler.ValidateConfiguration(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response["valid"].(bool))
		assert.Empty(t, response["errors"])
	})

	t.Run("validation fails - errors returned", func(t *testing.T) {
		db := setupValidationTestDB(t)
		handler := NewContainerConfigurationHandler(db)

		// Create test data
		user := &models.User{Email: "test@example.com", Name: "Test User", Role: "Developer"}
		require.NoError(t, db.Create(user).Error)

		service := &models.Service{Name: "test-service", UserID: user.ID}
		require.NoError(t, db.Create(service).Error)

		container := &models.Container{Name: "test-container"}
		require.NoError(t, db.Create(container).Error)

		version := &models.ContainerVersion{
			ContainerID:    container.ID,
			Version:        "v1.0.0",
			ComposeContent: "test",
		}
		require.NoError(t, db.Create(version).Error)

		rules := []map[string]interface{}{
			{
				"type":      "requires",
				"field":     "SSL.Enabled",
				"condition": "{{.SSL.Enabled}} == true",
				"target":    "SSL.Certificate",
				"message":   "SSL requires certificate path",
			},
		}
		rulesJSON, _ := json.Marshal(rules)

		config := &models.ContainerConfiguration{
			ContainerVersionID: version.ID,
			UISchema:           datatypes.JSON(`{}`),
			DependencyRules:    datatypes.JSON(rulesJSON),
		}
		require.NoError(t, db.Create(config).Error)

		serviceContainer := &models.ServiceContainer{
			ServiceID:          service.ID,
			ContainerID:        container.ID,
			ContainerVersionID: version.ID,
		}
		require.NoError(t, db.Create(serviceContainer).Error)

		// Prepare request - SSL enabled but no certificate
		reqBody := map[string]interface{}{
			"values": map[string]interface{}{
				"SSL": map[string]interface{}{
					"Enabled": true,
					// Certificate missing
				},
			},
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{
			{Key: "id", Value: "1"},
			{Key: "container_id", Value: "1"},
		}

		handler.ValidateConfiguration(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.False(t, response["valid"].(bool))
		assert.NotEmpty(t, response["errors"])

		errors := response["errors"].([]interface{})
		assert.Len(t, errors, 1)

		firstError := errors[0].(map[string]interface{})
		assert.Equal(t, "SSL.Certificate", firstError["field"])
		assert.Contains(t, firstError["message"], "SSL requires certificate")
		assert.Equal(t, "requires", firstError["rule"])
	})

	t.Run("service container not found", func(t *testing.T) {
		db := setupValidationTestDB(t)
		handler := NewContainerConfigurationHandler(db)

		reqBody := map[string]interface{}{
			"values": map[string]interface{}{},
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{
			{Key: "id", Value: "999"},
			{Key: "container_id", Value: "999"},
		}

		handler.ValidateConfiguration(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "Service container not found")
	})

	t.Run("container configuration not found", func(t *testing.T) {
		db := setupValidationTestDB(t)
		handler := NewContainerConfigurationHandler(db)

		// Create test data without configuration
		user := &models.User{Email: "test@example.com", Name: "Test User", Role: "Developer"}
		require.NoError(t, db.Create(user).Error)

		service := &models.Service{Name: "test-service", UserID: user.ID}
		require.NoError(t, db.Create(service).Error)

		container := &models.Container{Name: "test-container"}
		require.NoError(t, db.Create(container).Error)

		version := &models.ContainerVersion{
			ContainerID:    container.ID,
			Version:        "v1.0.0",
			ComposeContent: "test",
		}
		require.NoError(t, db.Create(version).Error)

		// No configuration created

		serviceContainer := &models.ServiceContainer{
			ServiceID:          service.ID,
			ContainerID:        container.ID,
			ContainerVersionID: version.ID,
		}
		require.NoError(t, db.Create(serviceContainer).Error)

		reqBody := map[string]interface{}{
			"values": map[string]interface{}{},
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{
			{Key: "id", Value: "1"},
			{Key: "container_id", Value: "1"},
		}

		handler.ValidateConfiguration(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "Container configuration not found")
	})

	t.Run("invalid JSON request", func(t *testing.T) {
		db := setupValidationTestDB(t)
		handler := NewContainerConfigurationHandler(db)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{
			{Key: "id", Value: "1"},
			{Key: "container_id", Value: "1"},
		}

		handler.ValidateConfiguration(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid dependency rules JSON", func(t *testing.T) {
		db := setupValidationTestDB(t)
		handler := NewContainerConfigurationHandler(db)

		// Create test data with invalid dependency rules JSON
		user := &models.User{Email: "test@example.com", Name: "Test User", Role: "Developer"}
		require.NoError(t, db.Create(user).Error)

		service := &models.Service{Name: "test-service", UserID: user.ID}
		require.NoError(t, db.Create(service).Error)

		container := &models.Container{Name: "test-container"}
		require.NoError(t, db.Create(container).Error)

		version := &models.ContainerVersion{
			ContainerID:    container.ID,
			Version:        "v1.0.0",
			ComposeContent: "test",
		}
		require.NoError(t, db.Create(version).Error)

		config := &models.ContainerConfiguration{
			ContainerVersionID: version.ID,
			UISchema:           datatypes.JSON(`{}`),
			DependencyRules:    datatypes.JSON(`invalid json`),
		}
		require.NoError(t, db.Create(config).Error)

		serviceContainer := &models.ServiceContainer{
			ServiceID:          service.ID,
			ContainerID:        container.ID,
			ContainerVersionID: version.ID,
		}
		require.NoError(t, db.Create(serviceContainer).Error)

		reqBody := map[string]interface{}{
			"values": map[string]interface{}{},
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{
			{Key: "id", Value: "1"},
			{Key: "container_id", Value: "1"},
		}

		handler.ValidateConfiguration(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "Failed to parse dependency rules")
	})

	t.Run("empty dependency rules - validation passes", func(t *testing.T) {
		db := setupValidationTestDB(t)
		handler := NewContainerConfigurationHandler(db)

		// Create test data with empty dependency rules
		user := &models.User{Email: "test@example.com", Name: "Test User", Role: "Developer"}
		require.NoError(t, db.Create(user).Error)

		service := &models.Service{Name: "test-service", UserID: user.ID}
		require.NoError(t, db.Create(service).Error)

		container := &models.Container{Name: "test-container"}
		require.NoError(t, db.Create(container).Error)

		version := &models.ContainerVersion{
			ContainerID:    container.ID,
			Version:        "v1.0.0",
			ComposeContent: "test",
		}
		require.NoError(t, db.Create(version).Error)

		config := &models.ContainerConfiguration{
			ContainerVersionID: version.ID,
			UISchema:           datatypes.JSON(`{}`),
			DependencyRules:    datatypes.JSON(`[]`),
		}
		require.NoError(t, db.Create(config).Error)

		serviceContainer := &models.ServiceContainer{
			ServiceID:          service.ID,
			ContainerID:        container.ID,
			ContainerVersionID: version.ID,
		}
		require.NoError(t, db.Create(serviceContainer).Error)

		reqBody := map[string]interface{}{
			"values": map[string]interface{}{
				"field": "value",
			},
		}
		body, _ := json.Marshal(reqBody)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{
			{Key: "id", Value: "1"},
			{Key: "container_id", Value: "1"},
		}

		handler.ValidateConfiguration(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.True(t, response["valid"].(bool))
		assert.Empty(t, response["errors"])
	})
}