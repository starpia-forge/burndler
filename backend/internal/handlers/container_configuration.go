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
func (h *ContainerConfigurationHandler) CreateConfiguration(c *gin.Context) {
	containerID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	versionID, _ := strconv.ParseUint(c.Param("version"), 10, 64)

	var req struct {
		UISchema        json.RawMessage `json:"ui_schema"`
		DependencyRules json.RawMessage `json:"dependency_rules"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify container version exists
	var version models.ContainerVersion
	if err := h.db.Where("id = ? AND container_id = ?", versionID, containerID).First(&version).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Container version not found"})
		return
	}

	config := &models.ContainerConfiguration{
		ContainerVersionID: uint(versionID),
		UISchema:           datatypes.JSON(req.UISchema),
		DependencyRules:    datatypes.JSON(req.DependencyRules),
	}

	if err := h.db.Create(config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create configuration"})
		return
	}

	c.JSON(http.StatusCreated, config)
}

// GetConfiguration retrieves a container configuration
func (h *ContainerConfigurationHandler) GetConfiguration(c *gin.Context) {
	versionID, _ := strconv.ParseUint(c.Param("version"), 10, 64)

	var config models.ContainerConfiguration
	if err := h.db.Where("container_version_id = ?", versionID).
		Preload("Files").
		Preload("Assets").
		First(&config).Error; err != nil {
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

	var config models.ContainerConfiguration
	if err := h.db.Where("container_version_id = ?", versionID).First(&config).Error; err != nil {
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
func (h *ContainerConfigurationHandler) DeleteConfiguration(c *gin.Context) {
	versionID, _ := strconv.ParseUint(c.Param("version"), 10, 64)

	if err := h.db.Where("container_version_id = ?", versionID).Delete(&models.ContainerConfiguration{}).Error; err != nil {
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

	// Load container configuration
	var config models.ContainerConfiguration
	if err := h.db.Where("container_version_id = ?", serviceContainer.ContainerVersionID).
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