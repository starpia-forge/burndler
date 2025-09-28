package services

import (
	"encoding/json"
	"fmt"

	"github.com/burndler/burndler/internal/models"
	"github.com/burndler/burndler/internal/storage"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ServiceService handles service management operations
type ServiceService struct {
	db      *gorm.DB
	storage storage.Storage
}

// NewServiceService creates a new ServiceService instance
func NewServiceService(db *gorm.DB, storage storage.Storage) *ServiceService {
	return &ServiceService{
		db:      db,
		storage: storage,
	}
}

// CreateServiceRequest represents the request to create a service
type CreateServiceRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateServiceRequest represents the request to update a service
type UpdateServiceRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Active      *bool   `json:"active"`
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

// ServiceFilters represents filters for listing services
type ServiceFilters struct {
	Active   *bool  `json:"active"`
	UserID   uint   `json:"user_id"`
	Name     string `json:"name"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// ValidationResult represents the result of service validation
type ValidationResult struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// CreateService creates a new service
func (s *ServiceService) CreateService(userID uint, req CreateServiceRequest) (*models.Service, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	// Check if service name already exists for this user
	var existingService models.Service
	if err := s.db.Where("name = ? AND user_id = ?", req.Name, userID).First(&existingService).Error; err == nil {
		return nil, fmt.Errorf("service with name '%s' already exists", req.Name)
	}

	service := &models.Service{
		Name:        req.Name,
		Description: req.Description,
		UserID:      userID,
		Active:      true,
	}

	if err := s.db.Create(service).Error; err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	return service, nil
}

// GetService retrieves a service by ID
func (s *ServiceService) GetService(id uint, includeContainers bool) (*models.Service, error) {
	var service models.Service
	query := s.db.Model(&models.Service{})

	if includeContainers {
		query = query.Preload("ServiceContainers").Preload("ServiceContainers.Container").Preload("ServiceContainers.ContainerVersion")
	}

	if err := query.First(&service, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("service not found")
		}
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	return &service, nil
}

// GetServiceByName retrieves a service by name and user ID
func (s *ServiceService) GetServiceByName(userID uint, name string, includeContainers bool) (*models.Service, error) {
	var service models.Service
	query := s.db.Model(&models.Service{}).Where("name = ? AND user_id = ?", name, userID)

	if includeContainers {
		query = query.Preload("ServiceContainers").Preload("ServiceContainers.Container").Preload("ServiceContainers.ContainerVersion")
	}

	if err := query.First(&service).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("service not found")
		}
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	return &service, nil
}

// ListServices lists services with filtering and pagination
func (s *ServiceService) ListServices(filters ServiceFilters) (*PaginatedResponse[models.Service], error) {
	var services []models.Service
	var total int64

	query := s.db.Model(&models.Service{})

	// Apply filters
	if filters.Active != nil {
		query = query.Where("active = ?", *filters.Active)
	}
	if filters.UserID > 0 {
		query = query.Where("user_id = ?", filters.UserID)
	}
	if filters.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filters.Name+"%")
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count services: %w", err)
	}

	// Apply pagination
	offset := (filters.Page - 1) * filters.PageSize
	if err := query.Offset(offset).Limit(filters.PageSize).Order("created_at DESC").Find(&services).Error; err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	totalPages := int(total) / filters.PageSize
	if int(total)%filters.PageSize > 0 {
		totalPages++
	}

	return &PaginatedResponse[models.Service]{
		Data:       services,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateService updates an existing service
func (s *ServiceService) UpdateService(id uint, req UpdateServiceRequest) (*models.Service, error) {
	var service models.Service
	if err := s.db.First(&service, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("service not found")
		}
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		service.Name = *req.Name
	}
	if req.Description != nil {
		service.Description = *req.Description
	}
	if req.Active != nil {
		service.Active = *req.Active
	}

	if err := s.db.Save(&service).Error; err != nil {
		return nil, fmt.Errorf("failed to update service: %w", err)
	}

	return &service, nil
}

// DeleteService soft deletes a service
func (s *ServiceService) DeleteService(id uint) error {
	result := s.db.Delete(&models.Service{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete service: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("service not found")
	}
	return nil
}

// AddContainerToService adds a container to a service
func (s *ServiceService) AddContainerToService(serviceID uint, req AddContainerToServiceRequest) (*models.ServiceContainer, error) {
	// Verify service exists
	var service models.Service
	if err := s.db.First(&service, serviceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("service not found")
		}
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	// Verify container and version exist
	var container models.Container
	if err := s.db.First(&container, req.ContainerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("container not found")
		}
		return nil, fmt.Errorf("failed to get container: %w", err)
	}

	var containerVersion models.ContainerVersion
	if err := s.db.Where("id = ? AND container_id = ?", req.ContainerVersionID, req.ContainerID).First(&containerVersion).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("container version not found")
		}
		return nil, fmt.Errorf("failed to get container version: %w", err)
	}

	// Check if container is already added to this service
	var existingServiceContainer models.ServiceContainer
	if err := s.db.Where("service_id = ? AND container_id = ?", serviceID, req.ContainerID).First(&existingServiceContainer).Error; err == nil {
		return nil, fmt.Errorf("container already added to this service")
	}

	// Prepare override variables
	var overrideVars datatypes.JSON
	if req.OverrideVars != nil {
		jsonData, err := json.Marshal(req.OverrideVars)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal override variables: %w", err)
		}
		overrideVars = datatypes.JSON(jsonData)
	}

	serviceContainer := &models.ServiceContainer{
		ServiceID:          serviceID,
		ContainerID:        req.ContainerID,
		ContainerVersionID: req.ContainerVersionID,
		Order:              req.Order,
		Enabled:            req.Enabled,
		OverrideVars:       overrideVars,
	}

	if err := s.db.Create(serviceContainer).Error; err != nil {
		return nil, fmt.Errorf("failed to add container to service: %w", err)
	}

	// Load relationships
	if err := s.db.Preload("Container").Preload("ContainerVersion").First(serviceContainer, serviceContainer.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load service container relationships: %w", err)
	}

	return serviceContainer, nil
}

// UpdateServiceContainer updates a service container configuration
func (s *ServiceService) UpdateServiceContainer(serviceContainerID uint, req UpdateServiceContainerRequest) (*models.ServiceContainer, error) {
	var serviceContainer models.ServiceContainer
	if err := s.db.First(&serviceContainer, serviceContainerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("service container not found")
		}
		return nil, fmt.Errorf("failed to get service container: %w", err)
	}

	// Update fields if provided
	if req.Order != nil {
		serviceContainer.Order = *req.Order
	}
	if req.Enabled != nil {
		serviceContainer.Enabled = *req.Enabled
	}
	if req.OverrideVars != nil {
		jsonData, err := json.Marshal(req.OverrideVars)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal override variables: %w", err)
		}
		serviceContainer.OverrideVars = datatypes.JSON(jsonData)
	}

	if err := s.db.Save(&serviceContainer).Error; err != nil {
		return nil, fmt.Errorf("failed to update service container: %w", err)
	}

	// Load relationships
	if err := s.db.Preload("Container").Preload("ContainerVersion").First(&serviceContainer, serviceContainer.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load service container relationships: %w", err)
	}

	return &serviceContainer, nil
}

// RemoveContainerFromService removes a container from a service
func (s *ServiceService) RemoveContainerFromService(serviceID, containerID uint) error {
	result := s.db.Where("service_id = ? AND container_id = ?", serviceID, containerID).Delete(&models.ServiceContainer{})
	if result.Error != nil {
		return fmt.Errorf("failed to remove container from service: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("container not found in service")
	}
	return nil
}

// GetServiceContainers retrieves all containers for a service
func (s *ServiceService) GetServiceContainers(serviceID uint) ([]models.ServiceContainer, error) {
	var serviceContainers []models.ServiceContainer
	if err := s.db.Where("service_id = ?", serviceID).
		Preload("Container").
		Preload("ContainerVersion").
		Order("\"order\"").
		Find(&serviceContainers).Error; err != nil {
		return nil, fmt.Errorf("failed to get service containers: %w", err)
	}
	return serviceContainers, nil
}

// ValidateService validates a service configuration
func (s *ServiceService) ValidateService(serviceID uint) (*ValidationResult, error) {
	service, err := s.GetService(serviceID, true)
	if err != nil {
		return nil, err
	}

	result := &ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Check if service has containers
	if len(service.ServiceContainers) == 0 {
		result.Errors = append(result.Errors, "Service must have at least one container")
		result.Valid = false
	}

	// Check if service has enabled containers
	enabledCount := service.GetContainerCount()
	if enabledCount == 0 {
		result.Errors = append(result.Errors, "Service must have at least one enabled container")
		result.Valid = false
	}

	// Check for duplicate container orders
	orderMap := make(map[int]bool)
	for _, sc := range service.ServiceContainers {
		if sc.Enabled {
			if orderMap[sc.Order] {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Duplicate order %d found in containers", sc.Order))
			}
			orderMap[sc.Order] = true
		}
	}

	return result, nil
}

// CanBuild checks if a service can be built
func (s *ServiceService) CanBuild(serviceID uint) (bool, error) {
	service, err := s.GetService(serviceID, true)
	if err != nil {
		return false, err
	}

	return service.CanBuild(), nil
}