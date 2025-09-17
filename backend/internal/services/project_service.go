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
	moduleService *ModuleService
}

// NewProjectService creates a new ProjectService instance
func NewProjectService(db *gorm.DB, moduleService *ModuleService) *ProjectService {
	return &ProjectService{
		db:            db,
		moduleService: moduleService,
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

// AddModuleToProjectRequest represents the request to add a module to a project
type AddModuleToProjectRequest struct {
	ModuleID        uint                   `json:"module_id" binding:"required"`
	ModuleVersionID uint                   `json:"module_version_id" binding:"required"`
	Order           int                    `json:"order"`
	Enabled         bool                   `json:"enabled"`
	OverrideVars    map[string]interface{} `json:"override_vars"`
}

// UpdateProjectModuleRequest represents the request to update a project module
type UpdateProjectModuleRequest struct {
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

// GetProject retrieves a project by ID with optional module loading
func (s *ProjectService) GetProject(id uint, includeModules bool) (*models.Project, error) {
	var project models.Project
	query := s.db

	if includeModules {
		query = query.Preload("ProjectModules", func(db *gorm.DB) *gorm.DB {
			return db.Order("project_modules.order ASC")
		}).Preload("ProjectModules.Module").Preload("ProjectModules.ModuleVersion")
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
func (s *ProjectService) GetProjectByName(userID uint, name string, includeModules bool) (*models.Project, error) {
	var project models.Project
	query := s.db

	if includeModules {
		query = query.Preload("ProjectModules", func(db *gorm.DB) *gorm.DB {
			return db.Order("project_modules.order ASC")
		}).Preload("ProjectModules.Module").Preload("ProjectModules.ModuleVersion")
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

// DeleteProject deletes a project and all its module associations
func (s *ProjectService) DeleteProject(id uint) error {
	project, err := s.GetProject(id, false)
	if err != nil {
		return err
	}

	// Delete in transaction to maintain consistency
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete all project modules first
		if err := tx.Where("project_id = ?", id).Delete(&models.ProjectModule{}).Error; err != nil {
			return fmt.Errorf("failed to delete project modules: %w", err)
		}

		// Delete the project
		if err := tx.Delete(project).Error; err != nil {
			return fmt.Errorf("failed to delete project: %w", err)
		}

		return nil
	})
}

// AddModuleToProject adds a module version to a project
func (s *ProjectService) AddModuleToProject(projectID uint, req AddModuleToProjectRequest) (*models.ProjectModule, error) {
	// Verify project exists
	project, err := s.GetProject(projectID, false)
	if err != nil {
		return nil, err
	}

	// Verify module version exists
	var moduleVersion models.ModuleVersion
	if err := s.db.Preload("Module").First(&moduleVersion, req.ModuleVersionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("module version not found")
		}
		return nil, fmt.Errorf("failed to get module version: %w", err)
	}

	// Verify the module version belongs to the specified module
	if moduleVersion.ModuleID != req.ModuleID {
		return nil, fmt.Errorf("module version does not belong to specified module")
	}

	// Check if module version is published
	if !moduleVersion.Published {
		return nil, fmt.Errorf("cannot add unpublished module version to project")
	}

	// Check if module is already in project
	var existingProjectModule models.ProjectModule
	if err := s.db.Where("project_id = ? AND module_id = ?", projectID, req.ModuleID).First(&existingProjectModule).Error; err == nil {
		return nil, fmt.Errorf("module '%s' is already added to project '%s'", moduleVersion.Module.Name, project.Name)
	}

	// Set default order if not provided
	if req.Order == 0 {
		var maxOrder int
		s.db.Model(&models.ProjectModule{}).Where("project_id = ?", projectID).Select("COALESCE(MAX(\"order\"), 0)").Scan(&maxOrder)
		req.Order = maxOrder + 1
	}

	// Convert override variables to JSON
	overrideVarsBytes, _ := json.Marshal(req.OverrideVars)
	overrideVarsJSON := datatypes.JSON(overrideVarsBytes)

	projectModule := &models.ProjectModule{
		ProjectID:       projectID,
		ModuleID:        req.ModuleID,
		ModuleVersionID: req.ModuleVersionID,
		Order:           req.Order,
		Enabled:         req.Enabled,
		OverrideVars:    overrideVarsJSON,
	}

	if err := s.db.Create(projectModule).Error; err != nil {
		return nil, fmt.Errorf("failed to add module to project: %w", err)
	}

	// Load relationships
	if err := s.db.Preload("Module").Preload("ModuleVersion").First(projectModule, projectModule.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load project module with relationships: %w", err)
	}

	return projectModule, nil
}

// UpdateProjectModule updates a project module configuration
func (s *ProjectService) UpdateProjectModule(projectModuleID uint, req UpdateProjectModuleRequest) (*models.ProjectModule, error) {
	var projectModule models.ProjectModule
	if err := s.db.Preload("Module").Preload("ModuleVersion").First(&projectModule, projectModuleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("project module not found")
		}
		return nil, fmt.Errorf("failed to get project module: %w", err)
	}

	// Update fields
	if req.Order != nil {
		projectModule.Order = *req.Order
	}
	if req.Enabled != nil {
		projectModule.Enabled = *req.Enabled
	}
	if req.OverrideVars != nil {
		overrideVarsBytes, _ := json.Marshal(req.OverrideVars)
		projectModule.OverrideVars = datatypes.JSON(overrideVarsBytes)
	}

	if err := s.db.Save(&projectModule).Error; err != nil {
		return nil, fmt.Errorf("failed to update project module: %w", err)
	}

	return &projectModule, nil
}

// RemoveModuleFromProject removes a module from a project
func (s *ProjectService) RemoveModuleFromProject(projectID, moduleID uint) error {
	// Verify project exists
	if _, err := s.GetProject(projectID, false); err != nil {
		return err
	}

	// Find and delete the project module
	result := s.db.Where("project_id = ? AND module_id = ?", projectID, moduleID).Delete(&models.ProjectModule{})
	if result.Error != nil {
		return fmt.Errorf("failed to remove module from project: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("module not found in project")
	}

	return nil
}

// GetProjectModules returns all modules in a project ordered by their order field
func (s *ProjectService) GetProjectModules(projectID uint) ([]models.ProjectModule, error) {
	// Verify project exists
	if _, err := s.GetProject(projectID, false); err != nil {
		return nil, err
	}

	var projectModules []models.ProjectModule
	if err := s.db.Where("project_id = ?", projectID).
		Preload("Module").
		Preload("ModuleVersion").
		Order("\"order\" ASC").
		Find(&projectModules).Error; err != nil {
		return nil, fmt.Errorf("failed to get project modules: %w", err)
	}

	return projectModules, nil
}

// ReorderProjectModules updates the order of modules in a project
func (s *ProjectService) ReorderProjectModules(projectID uint, moduleOrders map[uint]int) error {
	// Verify project exists
	if _, err := s.GetProject(projectID, false); err != nil {
		return err
	}

	// Update in transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		for moduleID, order := range moduleOrders {
			if err := tx.Model(&models.ProjectModule{}).
				Where("project_id = ? AND module_id = ?", projectID, moduleID).
				Update("order", order).Error; err != nil {
				return fmt.Errorf("failed to update order for module %d: %w", moduleID, err)
			}
		}
		return nil
	})
}