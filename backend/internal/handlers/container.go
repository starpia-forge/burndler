package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/burndler/burndler/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ContainerHandler handles container-related HTTP endpoints
type ContainerHandler struct {
	containerService *services.ContainerService
	db            *gorm.DB
}

// NewContainerHandler creates a new container handler
func NewContainerHandler(containerService *services.ContainerService, db *gorm.DB) *ContainerHandler {
	return &ContainerHandler{
		containerService: containerService,
		db:            db,
	}
}

// CreateContainerRequest represents the request to create a container
type CreateContainerRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
	Author      string `json:"author" binding:"max=100"`
	Repository  string `json:"repository" binding:"max=200"`
}

// UpdateContainerRequest represents the request to update a container
type UpdateContainerRequest struct {
	Description *string `json:"description" binding:"omitempty,max=500"`
	Author      *string `json:"author" binding:"omitempty,max=100"`
	Repository  *string `json:"repository" binding:"omitempty,max=200"`
	Active      *bool   `json:"active"`
}

// ContainerListQuery represents query parameters for listing containers
type ContainerListQuery struct {
	Page        int    `form:"page,default=1" binding:"min=1"`
	PageSize    int    `form:"page_size,default=10" binding:"min=1"`
	Active      *bool  `form:"active"`
	Author      string `form:"author"`
	ShowDeleted bool   `form:"show_deleted,default=false"`
	Published   bool   `form:"published_only,default=false"`
}

// CreateVersionRequest represents the request to create a container version
type CreateVersionRequest struct {
	Version       string                 `json:"version" binding:"required"`
	Compose       string                 `json:"compose" binding:"required"`
	Variables     map[string]interface{} `json:"variables"`
	ResourcePaths []string               `json:"resource_paths"`
	Dependencies  map[string]string      `json:"dependencies"`
}

// UpdateVersionRequest represents the request to update a container version
type UpdateVersionRequest struct {
	Compose       string                 `json:"compose"`
	Variables     map[string]interface{} `json:"variables"`
	ResourcePaths []string               `json:"resource_paths"`
	Dependencies  map[string]string      `json:"dependencies"`
}

// ValidateSemVer validates semantic versioning format
func ValidateSemVer(version string) error {
	// Ensure version starts with 'v'
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	_, err := semver.NewVersion(version)
	if err != nil {
		return errors.New("version must follow semantic versioning format (e.g., v1.0.0)")
	}

	return nil
}

// ListContainers handles GET /api/v1/containers
func (h *ContainerHandler) ListContainers(c *gin.Context) {
	var query ContainerListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_QUERY_PARAMS",
			Message: "Invalid query parameters",
		})
		return
	}

	// Validate page_size range manually
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// Convert to service filters
	filters := services.ContainerFilters{
		Page:         query.Page,
		PageSize:     query.PageSize,
		Active:       query.Active,
		Author:       query.Author,
		PublishedOnly: query.Published,
	}

	// Handle soft delete logic - GORM automatically handles soft delete filtering
	// Only show_deleted=true will include soft deleted records
	// No need to manually filter by Active field for soft delete

	result, err := h.containerService.ListContainers(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to list containers",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateContainer handles POST /api/v1/containers
func (h *ContainerHandler) CreateContainer(c *gin.Context) {
	var req CreateContainerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_FAILED",
			Message: "Invalid request format or missing required fields",
		})
		return
	}

	// Convert to service request
	serviceReq := services.CreateContainerRequest{
		Name:        req.Name,
		Description: req.Description,
		Author:      req.Author,
		Repository:  req.Repository,
	}

	container, err := h.containerService.CreateContainer(serviceReq)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "MODULE_EXISTS",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to create container",
		})
		return
	}

	c.JSON(http.StatusCreated, container)
}

// GetContainer handles GET /api/v1/containers/:id
func (h *ContainerHandler) GetContainer(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid container ID",
		})
		return
	}

	// Check if user wants to include versions
	includeVersions := c.Query("include_versions") == "true"

	container, err := h.containerService.GetContainer(uint(id), includeVersions)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "MODULE_NOT_FOUND",
				Message: "Container not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to get container",
		})
		return
	}

	c.JSON(http.StatusOK, container)
}

// UpdateContainer handles PUT /api/v1/containers/:id
func (h *ContainerHandler) UpdateContainer(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid container ID",
		})
		return
	}

	var req UpdateContainerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_FAILED",
			Message: "Invalid request format",
		})
		return
	}

	// Convert to service request
	serviceReq := services.UpdateContainerRequest{
		Active: req.Active,
	}
	if req.Description != nil {
		serviceReq.Description = *req.Description
	}
	if req.Author != nil {
		serviceReq.Author = *req.Author
	}
	if req.Repository != nil {
		serviceReq.Repository = *req.Repository
	}

	container, err := h.containerService.UpdateContainer(uint(id), serviceReq)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "MODULE_NOT_FOUND",
				Message: "Container not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to update container",
		})
		return
	}

	c.JSON(http.StatusOK, container)
}

// DeleteContainer handles DELETE /api/v1/containers/:id
func (h *ContainerHandler) DeleteContainer(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid container ID",
		})
		return
	}

	err = h.containerService.DeleteContainer(uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "MODULE_NOT_FOUND",
				Message: "Container not found",
			})
			return
		}
		if strings.Contains(err.Error(), "published versions") {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "MODULE_HAS_PUBLISHED_VERSIONS",
				Message: "Cannot delete container with published versions",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to delete container",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListVersions handles GET /api/v1/containers/:id/versions
func (h *ContainerHandler) ListVersions(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid container ID",
		})
		return
	}

	publishedOnly := c.Query("published_only") == "true"

	versions, err := h.containerService.ListVersions(uint(id), publishedOnly)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "MODULE_NOT_FOUND",
				Message: "Container not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to list versions",
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"data": versions,
	})
}

// CreateVersion handles POST /api/v1/containers/:id/versions
func (h *ContainerHandler) CreateVersion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid container ID",
		})
		return
	}

	var req CreateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_FAILED",
			Message: "Invalid request format or missing required fields",
		})
		return
	}

	// Validate semantic versioning
	if err := ValidateSemVer(req.Version); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_VERSION_FORMAT",
			Message: err.Error(),
		})
		return
	}

	// Ensure version starts with 'v'
	if !strings.HasPrefix(req.Version, "v") {
		req.Version = "v" + req.Version
	}

	// Convert to service request
	serviceReq := services.CreateVersionRequest{
		Version:       req.Version,
		Compose:       req.Compose,
		Variables:     req.Variables,
		ResourcePaths: req.ResourcePaths,
		Dependencies:  req.Dependencies,
	}

	version, err := h.containerService.CreateVersion(uint(id), serviceReq)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "MODULE_NOT_FOUND",
				Message: "Container not found",
			})
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "VERSION_EXISTS",
				Message: err.Error(),
			})
			return
		}
		if strings.Contains(err.Error(), "validation failed") {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "COMPOSE_VALIDATION_FAILED",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to create version",
		})
		return
	}

	c.JSON(http.StatusCreated, version)
}

// GetVersion handles GET /api/v1/containers/:id/versions/:version
func (h *ContainerHandler) GetVersion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid container ID",
		})
		return
	}

	versionParam := c.Param("version")

	version, err := h.containerService.GetVersion(uint(id), versionParam)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "VERSION_NOT_FOUND",
				Message: "Version not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to get version",
		})
		return
	}

	c.JSON(http.StatusOK, version)
}

// UpdateVersion handles PUT /api/v1/containers/:id/versions/:version
func (h *ContainerHandler) UpdateVersion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid container ID",
		})
		return
	}

	versionParam := c.Param("version")

	var req UpdateVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_FAILED",
			Message: "Invalid request format",
		})
		return
	}

	// Convert to service request
	serviceReq := services.UpdateVersionRequest{
		Compose:       req.Compose,
		Variables:     req.Variables,
		ResourcePaths: req.ResourcePaths,
		Dependencies:  req.Dependencies,
	}

	version, err := h.containerService.UpdateVersion(uint(id), versionParam, serviceReq)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "VERSION_NOT_FOUND",
				Message: "Version not found",
			})
			return
		}
		if strings.Contains(err.Error(), "cannot modify published") {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "VERSION_PUBLISHED",
				Message: "Cannot modify published version",
			})
			return
		}
		if strings.Contains(err.Error(), "validation failed") {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "COMPOSE_VALIDATION_FAILED",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to update version",
		})
		return
	}

	c.JSON(http.StatusOK, version)
}

// PublishVersion handles POST /api/v1/containers/:id/versions/:version/publish
func (h *ContainerHandler) PublishVersion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid container ID",
		})
		return
	}

	versionParam := c.Param("version")

	version, err := h.containerService.PublishVersion(uint(id), versionParam)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "VERSION_NOT_FOUND",
				Message: "Version not found",
			})
			return
		}
		if strings.Contains(err.Error(), "already published") {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "VERSION_ALREADY_PUBLISHED",
				Message: err.Error(),
			})
			return
		}
		if strings.Contains(err.Error(), "validation failed") {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "COMPOSE_VALIDATION_FAILED",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to publish version",
		})
		return
	}

	c.JSON(http.StatusOK, version)
}