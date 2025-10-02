package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/burndler/burndler/internal/models"
	"github.com/burndler/burndler/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ContainerConfigurationHandler struct {
	db *gorm.DB
}

func NewContainerConfigurationHandler(db *gorm.DB) *ContainerConfigurationHandler {
	return &ContainerConfigurationHandler{db: db}
}

// CreateConfiguration creates a new container configuration
// DEPRECATED: This handler will be replaced in Phase 3 with Container-level configuration API
func (h *ContainerConfigurationHandler) CreateConfiguration(c *gin.Context) {
	containerID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	versionID, _ := strconv.ParseUint(c.Param("version"), 10, 64)

	var req struct {
		Name            string          `json:"name"`
		Description     string          `json:"description"`
		MinimumVersion  string          `json:"minimum_version"`
		UISchema        json.RawMessage `json:"ui_schema"`
		DependencyRules json.RawMessage `json:"dependency_rules"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify container exists
	var container models.Container
	if err := h.db.First(&container, containerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Container not found"})
		return
	}

	// Set defaults
	if req.Name == "" {
		req.Name = "default"
	}
	if req.MinimumVersion == "" {
		// Get version from version ID
		var version models.ContainerVersion
		if err := h.db.First(&version, versionID).Error; err == nil {
			req.MinimumVersion = version.Version
		} else {
			req.MinimumVersion = "v0.1.0"
		}
	}

	config := &models.ContainerConfiguration{
		ContainerID:     uint(containerID),
		Name:            req.Name,
		Description:     req.Description,
		MinimumVersion:  req.MinimumVersion,
		UISchema:        datatypes.JSON(req.UISchema),
		DependencyRules: datatypes.JSON(req.DependencyRules),
	}

	if err := h.db.Create(config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create configuration"})
		return
	}

	// Update version to reference this configuration
	_ = h.db.Model(&models.ContainerVersion{}).Where("id = ?", versionID).Update("configuration_id", config.ID).Error
	// Ignore error - this is for backward compatibility with old handlers (will be removed in Phase 3)

	c.JSON(http.StatusCreated, config)
}

// GetConfiguration retrieves a container configuration
// DEPRECATED: This handler will be replaced in Phase 3 with Container-level configuration API
func (h *ContainerConfigurationHandler) GetConfiguration(c *gin.Context) {
	versionID, _ := strconv.ParseUint(c.Param("version"), 10, 64)

	// Load version to get configuration_id
	var version models.ContainerVersion
	if err := h.db.First(&version, versionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Version not found"})
		return
	}

	if version.ConfigurationID == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		return
	}

	var config models.ContainerConfiguration
	if err := h.db.Preload("Files").
		Preload("Assets").
		First(&config, *version.ConfigurationID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateConfiguration updates a container configuration
// DEPRECATED: This handler will be replaced in Phase 3 with Container-level configuration API
func (h *ContainerConfigurationHandler) UpdateConfiguration(c *gin.Context) {
	versionID, _ := strconv.ParseUint(c.Param("version"), 10, 64)

	var req struct {
		UISchema        json.RawMessage `json:"ui_schema"`
		DependencyRules json.RawMessage `json:"dependency_rules"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Load version to get configuration_id
	var version models.ContainerVersion
	if err := h.db.First(&version, versionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Version not found"})
		return
	}

	if version.ConfigurationID == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		return
	}

	var config models.ContainerConfiguration
	if err := h.db.First(&config, *version.ConfigurationID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		return
	}

	config.UISchema = datatypes.JSON(req.UISchema)
	config.DependencyRules = datatypes.JSON(req.DependencyRules)

	if err := h.db.Save(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// DeleteConfiguration deletes a container configuration
// DEPRECATED: This handler will be replaced in Phase 3 with Container-level configuration API
func (h *ContainerConfigurationHandler) DeleteConfiguration(c *gin.Context) {
	versionID, _ := strconv.ParseUint(c.Param("version"), 10, 64)

	// Load version to get configuration_id
	var version models.ContainerVersion
	if err := h.db.First(&version, versionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Version not found"})
		return
	}

	if version.ConfigurationID == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
		return
	}

	// Remove configuration reference from version
	if err := h.db.Model(&version).Update("configuration_id", nil).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update version"})
		return
	}

	// Delete configuration
	if err := h.db.Delete(&models.ContainerConfiguration{}, *version.ConfigurationID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete configuration"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configuration deleted successfully"})
}

// ValidateConfiguration validates configuration values against dependency rules
func (h *ContainerConfigurationHandler) ValidateConfiguration(c *gin.Context) {
	serviceID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	containerID, _ := strconv.ParseUint(c.Param("container_id"), 10, 64)

	var req struct {
		Values map[string]interface{} `json:"values"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Load service container
	var serviceContainer models.ServiceContainer
	if err := h.db.Where("service_id = ? AND container_id = ?", serviceID, containerID).
		First(&serviceContainer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service container not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve service container"})
		return
	}

	// Load container version to get configuration ID
	var version models.ContainerVersion
	if err := h.db.First(&version, serviceContainer.ContainerVersionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Container version not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve version"})
		return
	}

	// Check if version has a configuration
	if version.ConfigurationID == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Container configuration not found"})
		return
	}

	// Load container configuration
	var config models.ContainerConfiguration
	if err := h.db.Where("id = ?", *version.ConfigurationID).
		First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Container configuration not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve configuration"})
		return
	}

	// Parse dependency rules
	var rules []services.DependencyRule
	if len(config.DependencyRules) > 0 {
		if err := json.Unmarshal(config.DependencyRules, &rules); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse dependency rules"})
			return
		}
	}

	// Validate configuration
	checker := services.NewDependencyChecker()
	errors := checker.ValidateConfiguration(rules, req.Values)

	c.JSON(http.StatusOK, gin.H{
		"valid":  len(errors) == 0,
		"errors": errors,
	})
}

// GetServiceContainerConfiguration retrieves configuration for a specific container in a service
func (h *ContainerConfigurationHandler) GetServiceContainerConfiguration(c *gin.Context) {
	serviceID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	containerID, _ := strconv.ParseUint(c.Param("container_id"), 10, 64)

	// Load service container to get version
	var serviceContainer models.ServiceContainer
	if err := h.db.Where("service_id = ? AND container_id = ?", serviceID, containerID).
		First(&serviceContainer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, "Service container")
			return
		}
		InternalError(c, "Failed to retrieve service container")
		return
	}

	// Load container version to get configuration ID
	var version models.ContainerVersion
	if err := h.db.First(&version, serviceContainer.ContainerVersionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, "Container version")
			return
		}
		InternalError(c, "Failed to retrieve container version")
		return
	}

	// Check if version has a configuration
	if version.ConfigurationID == nil {
		NotFound(c, "Container configuration")
		return
	}

	// Load container configuration (UI schema and dependency rules)
	var containerConfig models.ContainerConfiguration
	if err := h.db.Where("id = ?", *version.ConfigurationID).
		Preload("Files").
		Preload("Assets").
		First(&containerConfig).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, "Container configuration")
			return
		}
		InternalError(c, "Failed to retrieve container configuration")
		return
	}

	// Load service configuration (current values)
	var serviceConfig models.ServiceConfiguration
	currentValues := make(map[string]interface{})
	err := h.db.Where("service_id = ? AND container_id = ?", serviceID, containerID).
		First(&serviceConfig).Error

	if err == nil && len(serviceConfig.ConfigurationValues) > 0 {
		if err := json.Unmarshal(serviceConfig.ConfigurationValues, &currentValues); err != nil {
			InternalError(c, "Failed to parse configuration values")
			return
		}
	} else if err != nil && err != gorm.ErrRecordNotFound {
		InternalError(c, "Failed to retrieve service configuration")
		return
	}

	// Parse UI schema and dependency rules
	var uiSchema map[string]interface{}
	var dependencyRules []services.DependencyRule

	if len(containerConfig.UISchema) > 0 {
		if err := json.Unmarshal(containerConfig.UISchema, &uiSchema); err != nil {
			InternalError(c, "Failed to parse UI schema")
			return
		}
	}

	if len(containerConfig.DependencyRules) > 0 {
		if err := json.Unmarshal(containerConfig.DependencyRules, &dependencyRules); err != nil {
			InternalError(c, "Failed to parse dependency rules")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"ui_schema":        uiSchema,
		"dependency_rules": dependencyRules,
		"current_values":   currentValues,
		"files":            containerConfig.Files,
		"assets":           containerConfig.Assets,
	})
}

// SaveServiceContainerConfiguration saves configuration values for a specific container in a service
func (h *ContainerConfigurationHandler) SaveServiceContainerConfiguration(c *gin.Context) {
	serviceID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	containerID, _ := strconv.ParseUint(c.Param("container_id"), 10, 64)

	var req struct {
		ConfigurationValues map[string]interface{} `json:"configuration_values"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// Verify service exists
	var service models.Service
	if err := h.db.First(&service, serviceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, "Service")
			return
		}
		InternalError(c, "Failed to retrieve service")
		return
	}

	// Verify container exists
	var container models.Container
	if err := h.db.First(&container, containerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, "Container")
			return
		}
		InternalError(c, "Failed to retrieve container")
		return
	}

	// Load service container to get version
	var serviceContainer models.ServiceContainer
	if err := h.db.Where("service_id = ? AND container_id = ?", serviceID, containerID).
		First(&serviceContainer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, "Service container")
			return
		}
		InternalError(c, "Failed to retrieve service container")
		return
	}

	// Load container version to get configuration ID
	var version models.ContainerVersion
	if err := h.db.First(&version, serviceContainer.ContainerVersionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, "Container version")
			return
		}
		InternalError(c, "Failed to retrieve container version")
		return
	}

	// Check if version has a configuration
	if version.ConfigurationID == nil {
		NotFound(c, "Container configuration")
		return
	}

	// Load container configuration for validation
	var containerConfig models.ContainerConfiguration
	if err := h.db.Where("id = ?", *version.ConfigurationID).
		First(&containerConfig).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, "Container configuration")
			return
		}
		InternalError(c, "Failed to retrieve container configuration")
		return
	}

	// Validate against dependency rules
	var rules []services.DependencyRule
	if len(containerConfig.DependencyRules) > 0 {
		if err := json.Unmarshal(containerConfig.DependencyRules, &rules); err != nil {
			InternalError(c, "Failed to parse dependency rules")
			return
		}
	}

	checker := services.NewDependencyChecker()
	validationErrors := checker.ValidateConfiguration(rules, req.ConfigurationValues)

	if len(validationErrors) > 0 {
		var errors []ValidationErrorItem
		for _, valErr := range validationErrors {
			errors = append(errors, ValidationErrorItem{
				Field:   valErr.Field,
				Message: valErr.Message,
			})
		}
		ValidationErrors(c, errors)
		return
	}

	// Marshal configuration values
	valuesJSON, err := json.Marshal(req.ConfigurationValues)
	if err != nil {
		InternalError(c, "Failed to encode configuration values")
		return
	}

	// Check if configuration exists
	var existingConfig models.ServiceConfiguration
	err = h.db.Where("service_id = ? AND container_id = ?", serviceID, containerID).
		First(&existingConfig).Error

	switch err {
	case gorm.ErrRecordNotFound:
		// Create new configuration
		config := &models.ServiceConfiguration{
			ServiceID:           uint(serviceID),
			ContainerID:         container.ID,
			ConfigurationValues: valuesJSON,
		}
		if err := h.db.Create(config).Error; err != nil {
			InternalError(c, "Failed to create configuration")
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"message": "Configuration saved successfully",
			"config":  config,
		})
	case nil:
		// Update existing configuration
		if err := h.db.Model(&existingConfig).
			Update("configuration_values", valuesJSON).Error; err != nil {
			InternalError(c, "Failed to update configuration")
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Configuration updated successfully",
			"config":  existingConfig,
		})
	default:
		InternalError(c, "Failed to retrieve service configuration")
		return
	}
}

// ExportServiceConfiguration exports all service configurations as JSON
func (h *ContainerConfigurationHandler) ExportServiceConfiguration(c *gin.Context) {
	serviceID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	// Load all service configurations for this service
	var configs []models.ServiceConfiguration
	if err := h.db.Where("service_id = ?", serviceID).
		Preload("Container").
		Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load configurations"})
		return
	}

	// Build export structure
	export := make(map[string]interface{})
	export["version"] = "1.0"
	export["service_id"] = serviceID

	containerConfigs := make(map[string]interface{})
	for _, config := range configs {
		var values map[string]interface{}
		if len(config.ConfigurationValues) > 0 {
			if err := json.Unmarshal(config.ConfigurationValues, &values); err != nil {
				continue // Skip invalid JSON
			}
		}
		containerConfigs[config.Container.Name] = values
	}
	export["containers"] = containerConfigs

	c.JSON(http.StatusOK, export)
}

// ImportServiceConfiguration imports configurations from JSON
func (h *ContainerConfigurationHandler) ImportServiceConfiguration(c *gin.Context) {
	serviceID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var importData struct {
		Version    string                            `json:"version"`
		Containers map[string]map[string]interface{} `json:"containers"`
	}

	if err := c.ShouldBindJSON(&importData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate version
	if importData.Version != "1.0" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported export version"})
		return
	}

	// Verify service exists
	var service models.Service
	if err := h.db.First(&service, serviceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load service"})
		return
	}

	// Import configurations
	importedCount := 0
	skippedContainers := []string{}

	for containerName, values := range importData.Containers {
		// Find container by name
		var container models.Container
		if err := h.db.Where("name = ?", containerName).First(&container).Error; err != nil {
			skippedContainers = append(skippedContainers, containerName)
			continue // Skip unknown containers
		}

		// Update or create service configuration
		valuesJSON, err := json.Marshal(values)
		if err != nil {
			skippedContainers = append(skippedContainers, containerName)
			continue
		}

		// Check if configuration exists
		var existingConfig models.ServiceConfiguration
		err = h.db.Where("service_id = ? AND container_id = ?", serviceID, container.ID).
			First(&existingConfig).Error

		if err == gorm.ErrRecordNotFound {
			// Create new configuration
			config := &models.ServiceConfiguration{
				ServiceID:           uint(serviceID),
				ContainerID:         container.ID,
				ConfigurationValues: valuesJSON,
			}
			if err := h.db.Create(config).Error; err != nil {
				skippedContainers = append(skippedContainers, containerName)
				continue
			}
		} else if err == nil {
			// Update existing configuration
			if err := h.db.Model(&existingConfig).
				Update("configuration_values", valuesJSON).Error; err != nil {
				skippedContainers = append(skippedContainers, containerName)
				continue
			}
		} else {
			// Database error
			skippedContainers = append(skippedContainers, containerName)
			continue
		}

		importedCount++
	}

	response := gin.H{
		"message":  "Configuration imported successfully",
		"imported": importedCount,
	}

	if len(skippedContainers) > 0 {
		response["skipped"] = skippedContainers
	}

	c.JSON(http.StatusOK, response)
}
