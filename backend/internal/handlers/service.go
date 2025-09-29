package handlers

import (
	"net/http"
	"strconv"

	"github.com/burndler/burndler/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ServiceHandler handles service-related HTTP endpoints
type ServiceHandler struct {
	serviceService *services.ServiceService
	db             *gorm.DB
}

// NewServiceHandler creates a new service handler
func NewServiceHandler(serviceService *services.ServiceService, db *gorm.DB) *ServiceHandler {
	return &ServiceHandler{
		serviceService: serviceService,
		db:             db,
	}
}

// CreateServiceRequest represents the request to create a service
type CreateServiceRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
}

// UpdateServiceRequest represents the request to update a service
type UpdateServiceRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=100"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	Active      *bool   `json:"active"`
}

// ServiceListQuery represents query parameters for listing services
type ServiceListQuery struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=10" binding:"min=1"`
	Active   *bool  `form:"active"`
	UserID   uint   `form:"user_id"`
	Name     string `form:"name"`
}

// AddContainerToServiceRequest represents the request to add a container to service
type AddContainerToServiceRequest struct {
	ContainerID        uint                   `json:"container_id" binding:"required"`
	ContainerVersionID uint                   `json:"container_version_id" binding:"required"`
	Order              int                    `json:"order"`
	Enabled            bool                   `json:"enabled"`
	OverrideVars       map[string]interface{} `json:"override_vars"`
}

// UpdateServiceContainerRequest represents the request to update a service container
type UpdateServiceContainerRequest struct {
	Order        *int                   `json:"order"`
	Enabled      *bool                  `json:"enabled"`
	OverrideVars map[string]interface{} `json:"override_vars"`
}

// CreateService handles POST /api/v1/services
func (h *ServiceHandler) CreateService(c *gin.Context) {
	var req CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request body",
		})
		return
	}

	// Get current user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	userIDString, ok := userIDStr.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Invalid user ID format in token",
		})
		return
	}

	// Convert user ID to uint
	userID, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Invalid user ID format",
		})
		return
	}

	// Convert to service request
	serviceReq := services.CreateServiceRequest{
		Name:        req.Name,
		Description: req.Description,
	}

	service, err := h.serviceService.CreateService(uint(userID), serviceReq)
	if err != nil {
		if err.Error() == "name is required" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "INVALID_REQUEST",
				Message: "Service name is required",
			})
			return
		}
		if err.Error() == "service with name '"+req.Name+"' already exists" {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "SERVICE_EXISTS",
				Message: "A service with this name already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to create service",
		})
		return
	}

	c.JSON(http.StatusCreated, service)
}

// GetService handles GET /api/v1/services/:id
func (h *ServiceHandler) GetService(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid service ID",
		})
		return
	}

	includeContainers := c.Query("include_containers") == "true"

	service, err := h.serviceService.GetService(uint(id), includeContainers)
	if err != nil {
		if err.Error() == "service not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "SERVICE_NOT_FOUND",
				Message: "Service not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to get service",
		})
		return
	}

	c.JSON(http.StatusOK, service)
}

// ListServices handles GET /api/v1/services
func (h *ServiceHandler) ListServices(c *gin.Context) {
	var query ServiceListQuery
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

	// Get current user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "User not authenticated",
		})
		return
	}

	userIDString, ok := userIDStr.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Invalid user ID format in token",
		})
		return
	}

	// Convert user ID to uint
	userID, err := strconv.ParseUint(userIDString, 10, 32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Invalid user ID format",
		})
		return
	}

	// Get role from context for admin check
	role, roleExists := c.Get("role")
	if !roleExists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "User role not found",
		})
		return
	}

	userRole, ok := role.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Invalid role format in token",
		})
		return
	}

	// Convert to service filters
	filters := services.ServiceFilters{
		Active:   query.Active,
		UserID:   uint(userID), // Only show user's own services
		Name:     query.Name,
		Page:     query.Page,
		PageSize: query.PageSize,
	}

	// Admin users can see all services if user_id is specified
	if userRole == "Admin" && query.UserID > 0 {
		filters.UserID = query.UserID
	}

	result, err := h.serviceService.ListServices(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to list services",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UpdateService handles PUT /api/v1/services/:id
func (h *ServiceHandler) UpdateService(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid service ID",
		})
		return
	}

	var req UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request body",
		})
		return
	}

	// Convert to service request
	serviceReq := services.UpdateServiceRequest{
		Name:        req.Name,
		Description: req.Description,
		Active:      req.Active,
	}

	service, err := h.serviceService.UpdateService(uint(id), serviceReq)
	if err != nil {
		if err.Error() == "service not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "SERVICE_NOT_FOUND",
				Message: "Service not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to update service",
		})
		return
	}

	c.JSON(http.StatusOK, service)
}

// DeleteService handles DELETE /api/v1/services/:id
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid service ID",
		})
		return
	}

	err = h.serviceService.DeleteService(uint(id))
	if err != nil {
		if err.Error() == "service not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "SERVICE_NOT_FOUND",
				Message: "Service not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to delete service",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetServiceContainers handles GET /api/v1/services/:id/containers
func (h *ServiceHandler) GetServiceContainers(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid service ID",
		})
		return
	}

	containers, err := h.serviceService.GetServiceContainers(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to get service containers",
		})
		return
	}

	c.JSON(http.StatusOK, containers)
}

// AddContainerToService handles POST /api/v1/services/:id/containers
func (h *ServiceHandler) AddContainerToService(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid service ID",
		})
		return
	}

	var req AddContainerToServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request body",
		})
		return
	}

	// Convert to service request
	serviceReq := services.AddContainerToServiceRequest{
		ContainerID:        req.ContainerID,
		ContainerVersionID: req.ContainerVersionID,
		Order:              req.Order,
		Enabled:            req.Enabled,
		OverrideVars:       req.OverrideVars,
	}

	serviceContainer, err := h.serviceService.AddContainerToService(uint(id), serviceReq)
	if err != nil {
		if err.Error() == "service not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "SERVICE_NOT_FOUND",
				Message: "Service not found",
			})
			return
		}
		if err.Error() == "container not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "CONTAINER_NOT_FOUND",
				Message: "Container not found",
			})
			return
		}
		if err.Error() == "container version not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "CONTAINER_VERSION_NOT_FOUND",
				Message: "Container version not found",
			})
			return
		}
		if err.Error() == "container already added to this service" {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "CONTAINER_ALREADY_ADDED",
				Message: "Container already added to this service",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to add container to service",
		})
		return
	}

	c.JSON(http.StatusCreated, serviceContainer)
}

// UpdateServiceContainer handles PUT /api/v1/services/:id/containers/:container_id
func (h *ServiceHandler) UpdateServiceContainer(c *gin.Context) {
	containerIDParam := c.Param("container_id")
	containerID, err := strconv.ParseUint(containerIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid container ID",
		})
		return
	}

	var req UpdateServiceContainerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request body",
		})
		return
	}

	// Convert to service request
	serviceReq := services.UpdateServiceContainerRequest{
		Order:        req.Order,
		Enabled:      req.Enabled,
		OverrideVars: req.OverrideVars,
	}

	serviceContainer, err := h.serviceService.UpdateServiceContainer(uint(containerID), serviceReq)
	if err != nil {
		if err.Error() == "service container not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "SERVICE_CONTAINER_NOT_FOUND",
				Message: "Service container not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to update service container",
		})
		return
	}

	c.JSON(http.StatusOK, serviceContainer)
}

// RemoveContainerFromService handles DELETE /api/v1/services/:id/containers/:container_id
func (h *ServiceHandler) RemoveContainerFromService(c *gin.Context) {
	idParam := c.Param("id")
	serviceID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid service ID",
		})
		return
	}

	containerIDParam := c.Param("container_id")
	containerID, err := strconv.ParseUint(containerIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid container ID",
		})
		return
	}

	err = h.serviceService.RemoveContainerFromService(uint(serviceID), uint(containerID))
	if err != nil {
		if err.Error() == "container not found in service" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "CONTAINER_NOT_FOUND_IN_SERVICE",
				Message: "Container not found in service",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to remove container from service",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ValidateService handles POST /api/v1/services/:id/validate
func (h *ServiceHandler) ValidateService(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid service ID",
		})
		return
	}

	result, err := h.serviceService.ValidateService(uint(id))
	if err != nil {
		if err.Error() == "service not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "SERVICE_NOT_FOUND",
				Message: "Service not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to validate service",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// BuildService handles POST /api/v1/services/:id/build
func (h *ServiceHandler) BuildService(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid service ID",
		})
		return
	}

	canBuild, err := h.serviceService.CanBuild(uint(id))
	if err != nil {
		if err.Error() == "service not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "SERVICE_NOT_FOUND",
				Message: "Service not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to check service build status",
		})
		return
	}

	if !canBuild {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "SERVICE_NOT_BUILDABLE",
			Message: "Service is not ready for building",
		})
		return
	}

	// TODO: Implement actual build logic
	// For now, return a success response
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Service build initiated",
	})
}
