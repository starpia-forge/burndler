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

// ModuleHandler handles module-related HTTP endpoints
type ModuleHandler struct {
	moduleService *services.ModuleService
	db            *gorm.DB
}

// NewModuleHandler creates a new module handler
func NewModuleHandler(moduleService *services.ModuleService, db *gorm.DB) *ModuleHandler {
	return &ModuleHandler{
		moduleService: moduleService,
		db:            db,
	}
}

// CreateModuleRequest represents the request to create a module
type CreateModuleRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
	Author      string `json:"author" binding:"max=100"`
	Repository  string `json:"repository" binding:"max=200"`
}

// UpdateModuleRequest represents the request to update a module
type UpdateModuleRequest struct {
	Description *string `json:"description" binding:"omitempty,max=500"`
	Author      *string `json:"author" binding:"omitempty,max=100"`
	Repository  *string `json:"repository" binding:"omitempty,max=200"`
	Active      *bool   `json:"active"`
}

// ModuleListQuery represents query parameters for listing modules
type ModuleListQuery struct {
	Page        int    `form:"page,default=1" binding:"min=1"`
	PageSize    int    `form:"page_size,default=10" binding:"min=1"`
	Active      *bool  `form:"active"`
	Author      string `form:"author"`
	ShowDeleted bool   `form:"show_deleted,default=false"`
	Published   bool   `form:"published_only,default=false"`
}

// CreateVersionRequest represents the request to create a module version
type CreateVersionRequest struct {
	Version       string                 `json:"version" binding:"required"`
	Compose       string                 `json:"compose" binding:"required"`
	Variables     map[string]interface{} `json:"variables"`
	ResourcePaths []string               `json:"resource_paths"`
	Dependencies  map[string]string      `json:"dependencies"`
}

// UpdateVersionRequest represents the request to update a module version
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

// ListModules handles GET /api/v1/modules
func (h *ModuleHandler) ListModules(c *gin.Context) {
	var query ModuleListQuery
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
	filters := services.ModuleFilters{
		Page:         query.Page,
		PageSize:     query.PageSize,
		Active:       query.Active,
		Author:       query.Author,
		PublishedOnly: query.Published,
	}

	// Handle soft delete logic - GORM automatically handles soft delete filtering
	// Only show_deleted=true will include soft deleted records
	// No need to manually filter by Active field for soft delete

	result, err := h.moduleService.ListModules(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to list modules",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CreateModule handles POST /api/v1/modules
func (h *ModuleHandler) CreateModule(c *gin.Context) {
	var req CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_FAILED",
			Message: "Invalid request format or missing required fields",
		})
		return
	}

	// Convert to service request
	serviceReq := services.CreateModuleRequest{
		Name:        req.Name,
		Description: req.Description,
		Author:      req.Author,
		Repository:  req.Repository,
	}

	module, err := h.moduleService.CreateModule(serviceReq)
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
			Message: "Failed to create module",
		})
		return
	}

	c.JSON(http.StatusCreated, module)
}

// GetModule handles GET /api/v1/modules/:id
func (h *ModuleHandler) GetModule(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid module ID",
		})
		return
	}

	// Check if user wants to include versions
	includeVersions := c.Query("include_versions") == "true"

	module, err := h.moduleService.GetModule(uint(id), includeVersions)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "MODULE_NOT_FOUND",
				Message: "Module not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to get module",
		})
		return
	}

	c.JSON(http.StatusOK, module)
}

// UpdateModule handles PUT /api/v1/modules/:id
func (h *ModuleHandler) UpdateModule(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid module ID",
		})
		return
	}

	var req UpdateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_FAILED",
			Message: "Invalid request format",
		})
		return
	}

	// Convert to service request
	serviceReq := services.UpdateModuleRequest{
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

	module, err := h.moduleService.UpdateModule(uint(id), serviceReq)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "MODULE_NOT_FOUND",
				Message: "Module not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to update module",
		})
		return
	}

	c.JSON(http.StatusOK, module)
}

// DeleteModule handles DELETE /api/v1/modules/:id
func (h *ModuleHandler) DeleteModule(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid module ID",
		})
		return
	}

	err = h.moduleService.DeleteModule(uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "MODULE_NOT_FOUND",
				Message: "Module not found",
			})
			return
		}
		if strings.Contains(err.Error(), "published versions") {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "MODULE_HAS_PUBLISHED_VERSIONS",
				Message: "Cannot delete module with published versions",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to delete module",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListVersions handles GET /api/v1/modules/:id/versions
func (h *ModuleHandler) ListVersions(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid module ID",
		})
		return
	}

	publishedOnly := c.Query("published_only") == "true"

	versions, err := h.moduleService.ListVersions(uint(id), publishedOnly)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "MODULE_NOT_FOUND",
				Message: "Module not found",
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

// CreateVersion handles POST /api/v1/modules/:id/versions
func (h *ModuleHandler) CreateVersion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid module ID",
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

	version, err := h.moduleService.CreateVersion(uint(id), serviceReq)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "MODULE_NOT_FOUND",
				Message: "Module not found",
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

// GetVersion handles GET /api/v1/modules/:id/versions/:version
func (h *ModuleHandler) GetVersion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid module ID",
		})
		return
	}

	versionParam := c.Param("version")

	version, err := h.moduleService.GetVersion(uint(id), versionParam)
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

// UpdateVersion handles PUT /api/v1/modules/:id/versions/:version
func (h *ModuleHandler) UpdateVersion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid module ID",
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

	version, err := h.moduleService.UpdateVersion(uint(id), versionParam, serviceReq)
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

// PublishVersion handles POST /api/v1/modules/:id/versions/:version/publish
func (h *ModuleHandler) PublishVersion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid module ID",
		})
		return
	}

	versionParam := c.Param("version")

	version, err := h.moduleService.PublishVersion(uint(id), versionParam)
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