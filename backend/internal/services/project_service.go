package services

import (
	"encoding/json"
	"fmt"

	"github.com/burndler/burndler/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ProjectService handles project management operations
type ProjectService struct {
	db            *gorm.DB
	containerService *ContainerService
}

// NewProjectService creates a new ProjectService instance
func NewProjectService(db *gorm.DB, containerService *ContainerService) *ProjectService {
	return &ProjectService{
		db:            db,
		containerService: containerService,
	}
}

// CreateProjectRequest represents the request to create a project
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateProjectRequest represents the request to update a project
type UpdateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AddContainerToProjectRequest represents the request to add a container to a project
type AddContainerToProjectRequest struct {
	ContainerID        uint                   `json:"container_id" binding:"required"`
	ContainerVersionID uint                   `json:"container_version_id" binding:"required"`
	Order           int                    `json:"order"`
	Enabled         bool                   `json:"enabled"`
	OverrideVars    map[string]interface{} `json:"override_vars"`
}

// UpdateProjectContainerRequest represents the request to update a project container
type UpdateProjectContainerRequest struct {
	Order        *int                   `json:"order"`
	Enabled      *bool                  `json:"enabled"`
	OverrideVars map[string]interface{} `json:"override_vars"`
}

// ProjectFilters represents filters for listing projects
type ProjectFilters struct {
	UserID   uint   `json:"user_id"`
	Name     string `json:"name"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// CreateProject creates a new project
func (s *ProjectService) CreateProject(userID uint, req CreateProjectRequest) (*models.Project, error) {
	// Check if project name already exists for this user
	var existingProject models.Project
	if err := s.db.Where("user_id = ? AND name = ?", userID, req.Name).First(&existingProject).Error; err == nil {
		return nil, fmt.Errorf("project with name '%s' already exists for user", req.Name)
	}

	project := &models.Project{
		Name:        req.Name,
		Description: req.Description,
		UserID:      userID,
	}

	if err := s.db.Create(project).Error; err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return project, nil
}

// GetProject retrieves a project by ID with optional container loading
func (s *ProjectService) GetProject(id uint, includeContainers bool) (*models.Project, error) {
	var project models.Project
	query := s.db

	if includeContainers {
		query = query.Preload("ProjectContainers", func(db *gorm.DB) *gorm.DB {
			return db.Order("project_containers.order ASC")
		}).Preload("ProjectContainers.Container").Preload("ProjectContainers.ContainerVersion")
	}

	if err := query.First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

// GetProjectByName retrieves a project by name for a specific user
func (s *ProjectService) GetProjectByName(userID uint, name string, includeContainers bool) (*models.Project, error) {
	var project models.Project
	query := s.db

	if includeContainers {
		query = query.Preload("ProjectContainers", func(db *gorm.DB) *gorm.DB {
			return db.Order("project_containers.order ASC")
		}).Preload("ProjectContainers.Container").Preload("ProjectContainers.ContainerVersion")
	}

	if err := query.Where("user_id = ? AND name = ?", userID, name).First(&project).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("project '%s' not found for user", name)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

// ListProjects returns a paginated list of projects
func (s *ProjectService) ListProjects(filters ProjectFilters) (*PaginatedResponse[models.Project], error) {
	var projects []models.Project
	var total int64

	query := s.db.Model(&models.Project{})

	// Apply filters
	if filters.UserID != 0 {
		query = query.Where("user_id = ?", filters.UserID)
	}

	if filters.Name != "" {
		query = query.Where("name LIKE ?", "%"+filters.Name+"%")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count projects: %w", err)
	}

	// Set pagination defaults
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 {
		filters.PageSize = 10
	}
	if filters.PageSize > 100 {
		filters.PageSize = 100
	}

	offset := (filters.Page - 1) * filters.PageSize

	// Get paginated results
	if err := query.Offset(offset).Limit(filters.PageSize).Order("created_at DESC").Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	totalPages := int((total + int64(filters.PageSize) - 1) / int64(filters.PageSize))

	return &PaginatedResponse[models.Project]{
		Data:       projects,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateProject updates an existing project
func (s *ProjectService) UpdateProject(id uint, req UpdateProjectRequest) (*models.Project, error) {
	project, err := s.GetProject(id, false)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != "" {
		// Check if new name conflicts with existing project for same user
		var existingProject models.Project
		if err := s.db.Where("user_id = ? AND name = ? AND id != ?", project.UserID, req.Name, id).First(&existingProject).Error; err == nil {
			return nil, fmt.Errorf("project with name '%s' already exists for user", req.Name)
		}
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}

	if err := s.db.Save(project).Error; err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return project, nil
}

// DeleteProject deletes a project and all its container associations
func (s *ProjectService) DeleteProject(id uint) error {
	project, err := s.GetProject(id, false)
	if err != nil {
		return err
	}

	// Delete in transaction to maintain consistency
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete all project containers first
		if err := tx.Where("project_id = ?", id).Delete(&models.ProjectContainer{}).Error; err != nil {
			return fmt.Errorf("failed to delete project containers: %w", err)
		}

		// Delete the project
		if err := tx.Delete(project).Error; err != nil {
			return fmt.Errorf("failed to delete project: %w", err)
		}

		return nil
	})
}

// AddContainerToProject adds a container version to a project
func (s *ProjectService) AddContainerToProject(projectID uint, req AddContainerToProjectRequest) (*models.ProjectContainer, error) {
	// Verify project exists
	project, err := s.GetProject(projectID, false)
	if err != nil {
		return nil, err
	}

	// Verify container version exists
	var containerVersion models.ContainerVersion
	if err := s.db.Preload("Container").First(&containerVersion, req.ContainerVersionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("container version not found")
		}
		return nil, fmt.Errorf("failed to get container version: %w", err)
	}

	// Verify the container version belongs to the specified container
	if containerVersion.ContainerID != req.ContainerID {
		return nil, fmt.Errorf("container version does not belong to specified container")
	}

	// Check if container version is published
	if !containerVersion.Published {
		return nil, fmt.Errorf("cannot add unpublished container version to project")
	}

	// Check if container is already in project
	var existingProjectContainer models.ProjectContainer
	if err := s.db.Where("project_id = ? AND container_id = ?", projectID, req.ContainerID).First(&existingProjectContainer).Error; err == nil {
		return nil, fmt.Errorf("container '%s' is already added to project '%s'", containerVersion.Container.Name, project.Name)
	}

	// Set default order if not provided
	if req.Order == 0 {
		var maxOrder int
		s.db.Model(&models.ProjectContainer{}).Where("project_id = ?", projectID).Select("COALESCE(MAX(\"order\"), 0)").Scan(&maxOrder)
		req.Order = maxOrder + 1
	}

	// Convert override variables to JSON
	overrideVarsBytes, _ := json.Marshal(req.OverrideVars)
	overrideVarsJSON := datatypes.JSON(overrideVarsBytes)

	projectContainer := &models.ProjectContainer{
		ProjectID:       projectID,
		ContainerID:     req.ContainerID,
		ContainerVersionID: req.ContainerVersionID,
		Order:           req.Order,
		Enabled:         req.Enabled,
		OverrideVars:    overrideVarsJSON,
	}

	if err := s.db.Create(projectContainer).Error; err != nil {
		return nil, fmt.Errorf("failed to add module to project: %w", err)
	}

	// Load relationships
	if err := s.db.Preload("Container").Preload("ContainerVersion").First(projectContainer, projectContainer.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load project module with relationships: %w", err)
	}

	return projectContainer, nil
}

// UpdateProjectModule updates a project module configuration
func (s *ProjectService) UpdateProjectContainer(projectContainerID uint, req UpdateProjectContainerRequest) (*models.ProjectContainer, error) {
	var projectContainer models.ProjectContainer
	if err := s.db.Preload("Container").Preload("ContainerVersion").First(&projectContainer, projectContainerID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("project module not found")
		}
		return nil, fmt.Errorf("failed to get project module: %w", err)
	}

	// Update fields
	if req.Order != nil {
		projectContainer.Order = *req.Order
	}
	if req.Enabled != nil {
		projectContainer.Enabled = *req.Enabled
	}
	if req.OverrideVars != nil {
		overrideVarsBytes, _ := json.Marshal(req.OverrideVars)
		projectContainer.OverrideVars = datatypes.JSON(overrideVarsBytes)
	}

	if err := s.db.Save(&projectContainer).Error; err != nil {
		return nil, fmt.Errorf("failed to update project module: %w", err)
	}

	return &projectContainer, nil
}

// RemoveModuleFromProject removes a module from a project
func (s *ProjectService) RemoveContainerFromProject(projectID, containerID uint) error {
	// Verify project exists
	if _, err := s.GetProject(projectID, false); err != nil {
		return err
	}

	// Find and delete the project module
	result := s.db.Where("project_id = ? AND container_id = ?", projectID, containerID).Delete(&models.ProjectContainer{})
	if result.Error != nil {
		return fmt.Errorf("failed to remove module from project: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("module not found in project")
	}

	return nil
}

// GetProjectContainers returns all modules in a project ordered by their order field
func (s *ProjectService) GetProjectContainers(projectID uint) ([]models.ProjectContainer, error) {
	// Verify project exists
	if _, err := s.GetProject(projectID, false); err != nil {
		return nil, err
	}

	var projectContainers []models.ProjectContainer
	if err := s.db.Where("project_id = ?", projectID).
		Preload("Container").
		Preload("ContainerVersion").
		Order("\"order\" ASC").
		Find(&projectContainers).Error; err != nil {
		return nil, fmt.Errorf("failed to get project modules: %w", err)
	}

	return projectContainers, nil
}

// ReorderProjectContainers updates the order of modules in a project
func (s *ProjectService) ReorderProjectContainers(projectID uint, containerOrders map[uint]int) error {
	// Verify project exists
	if _, err := s.GetProject(projectID, false); err != nil {
		return err
	}

	// Update in transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		for containerID, order := range containerOrders {
			if err := tx.Model(&models.ProjectContainer{}).
				Where("project_id = ? AND container_id = ?", projectID, containerID).
				Update("order", order).Error; err != nil {
				return fmt.Errorf("failed to update order for container %d: %w", containerID, err)
			}
		}
		return nil
	})
}