package services

import (
	"encoding/json"
	"fmt"

	"github.com/burndler/burndler/internal/models"
	"github.com/burndler/burndler/internal/storage"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ModuleService handles module management operations
type ModuleService struct {
	db      *gorm.DB
	storage storage.Storage
	linter  *Linter
}

// NewModuleService creates a new ModuleService instance
func NewModuleService(db *gorm.DB, storage storage.Storage, linter *Linter) *ModuleService {
	return &ModuleService{
		db:      db,
		storage: storage,
		linter:  linter,
	}
}

// CreateModuleRequest represents the request to create a module
type CreateModuleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Repository  string `json:"repository"`
}

// UpdateModuleRequest represents the request to update a module
type UpdateModuleRequest struct {
	Description string `json:"description"`
	Author      string `json:"author"`
	Repository  string `json:"repository"`
	Active      *bool  `json:"active"`
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

// ModuleFilters represents filters for listing modules
type ModuleFilters struct {
	Active       *bool  `json:"active"`
	Author       string `json:"author"`
	PublishedOnly bool   `json:"published_only"`
	Page         int    `json:"page"`
	PageSize     int    `json:"page_size"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// CreateModule creates a new module
func (s *ModuleService) CreateModule(req CreateModuleRequest) (*models.Module, error) {
	// Check if module name already exists
	var existingModule models.Module
	if err := s.db.Where("name = ?", req.Name).First(&existingModule).Error; err == nil {
		return nil, fmt.Errorf("module with name '%s' already exists", req.Name)
	}

	module := &models.Module{
		Name:        req.Name,
		Description: req.Description,
		Author:      req.Author,
		Repository:  req.Repository,
		Active:      true,
	}

	if err := s.db.Create(module).Error; err != nil {
		return nil, fmt.Errorf("failed to create module: %w", err)
	}

	return module, nil
}

// GetModule retrieves a module by ID with optional version loading
func (s *ModuleService) GetModule(id uint, includeVersions bool) (*models.Module, error) {
	var module models.Module
	query := s.db

	if includeVersions {
		query = query.Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		})
	}

	if err := query.First(&module, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("module not found")
		}
		return nil, fmt.Errorf("failed to get module: %w", err)
	}

	return &module, nil
}

// GetModuleByName retrieves a module by name
func (s *ModuleService) GetModuleByName(name string, includeVersions bool) (*models.Module, error) {
	var module models.Module
	query := s.db

	if includeVersions {
		query = query.Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		})
	}

	if err := query.Where("name = ?", name).First(&module).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("module '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to get module: %w", err)
	}

	return &module, nil
}

// ListModules returns a paginated list of modules
func (s *ModuleService) ListModules(filters ModuleFilters) (*PaginatedResponse[models.Module], error) {
	var modules []models.Module
	var total int64

	query := s.db.Model(&models.Module{})

	// Apply filters
	if filters.Active != nil {
		query = query.Where("active = ?", *filters.Active)
	}

	if filters.Author != "" {
		query = query.Where("author LIKE ?", "%"+filters.Author+"%")
	}

	if filters.PublishedOnly {
		query = query.Joins("JOIN module_versions ON modules.id = module_versions.module_id").
			Where("module_versions.published = ?", true).
			Group("modules.id")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count modules: %w", err)
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
	if err := query.Offset(offset).Limit(filters.PageSize).Order("created_at DESC").Find(&modules).Error; err != nil {
		return nil, fmt.Errorf("failed to list modules: %w", err)
	}

	totalPages := int((total + int64(filters.PageSize) - 1) / int64(filters.PageSize))

	return &PaginatedResponse[models.Module]{
		Data:       modules,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateModule updates an existing module
func (s *ModuleService) UpdateModule(id uint, req UpdateModuleRequest) (*models.Module, error) {
	module, err := s.GetModule(id, false)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Description != "" {
		module.Description = req.Description
	}
	if req.Author != "" {
		module.Author = req.Author
	}
	if req.Repository != "" {
		module.Repository = req.Repository
	}
	if req.Active != nil {
		module.Active = *req.Active
	}

	if err := s.db.Save(module).Error; err != nil {
		return nil, fmt.Errorf("failed to update module: %w", err)
	}

	return module, nil
}

// DeleteModule soft deletes a module
func (s *ModuleService) DeleteModule(id uint) error {
	module, err := s.GetModule(id, true)
	if err != nil {
		return err
	}

	// Check if module has published versions
	if module.HasPublishedVersions() {
		return fmt.Errorf("cannot delete module with published versions")
	}

	if err := s.db.Delete(module).Error; err != nil {
		return fmt.Errorf("failed to delete module: %w", err)
	}

	return nil
}

// CreateVersion creates a new version for a module
func (s *ModuleService) CreateVersion(moduleID uint, req CreateVersionRequest) (*models.ModuleVersion, error) {
	// Verify module exists
	module, err := s.GetModule(moduleID, false)
	if err != nil {
		return nil, err
	}

	// Check if version already exists
	var existingVersion models.ModuleVersion
	if err := s.db.Where("module_id = ? AND version = ?", moduleID, req.Version).First(&existingVersion).Error; err == nil {
		return nil, fmt.Errorf("version '%s' already exists for module '%s'", req.Version, module.Name)
	}

	// Validate compose content
	if err := s.linter.ValidateCompose(req.Compose); err != nil {
		return nil, fmt.Errorf("compose validation failed: %w", err)
	}

	// Convert maps to JSON
	variablesBytes, _ := json.Marshal(req.Variables)
	resourcePathsBytes, _ := json.Marshal(req.ResourcePaths)
	dependenciesBytes, _ := json.Marshal(req.Dependencies)

	variablesJSON := datatypes.JSON(variablesBytes)
	resourcePathsJSON := datatypes.JSON(resourcePathsBytes)
	dependenciesJSON := datatypes.JSON(dependenciesBytes)

	version := &models.ModuleVersion{
		ModuleID:        moduleID,
		Version:         req.Version,
		ComposeContent:  req.Compose,
		Variables:       variablesJSON,
		ResourcePaths:   resourcePathsJSON,
		Dependencies:    dependenciesJSON,
		Published:       false,
	}

	if err := s.db.Create(version).Error; err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	// Load the module relationship
	if err := s.db.Preload("Module").First(version, version.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load version with module: %w", err)
	}

	return version, nil
}

// GetVersion retrieves a specific version of a module
func (s *ModuleService) GetVersion(moduleID uint, version string) (*models.ModuleVersion, error) {
	var moduleVersion models.ModuleVersion

	if err := s.db.Preload("Module").Where("module_id = ? AND version = ?", moduleID, version).First(&moduleVersion).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("version '%s' not found", version)
		}
		return nil, fmt.Errorf("failed to get version: %w", err)
	}

	return &moduleVersion, nil
}

// UpdateVersion updates an existing module version (only if unpublished)
func (s *ModuleService) UpdateVersion(moduleID uint, version string, req UpdateVersionRequest) (*models.ModuleVersion, error) {
	moduleVersion, err := s.GetVersion(moduleID, version)
	if err != nil {
		return nil, err
	}

	if !moduleVersion.CanModify() {
		return nil, fmt.Errorf("cannot modify published version")
	}

	// Update fields
	if req.Compose != "" {
		// Validate compose content
		if err := s.linter.ValidateCompose(req.Compose); err != nil {
			return nil, fmt.Errorf("compose validation failed: %w", err)
		}
		moduleVersion.ComposeContent = req.Compose
	}

	if req.Variables != nil {
		variablesBytes, _ := json.Marshal(req.Variables)
		moduleVersion.Variables = datatypes.JSON(variablesBytes)
	}

	if req.ResourcePaths != nil {
		resourcePathsBytes, _ := json.Marshal(req.ResourcePaths)
		moduleVersion.ResourcePaths = datatypes.JSON(resourcePathsBytes)
	}

	if req.Dependencies != nil {
		dependenciesBytes, _ := json.Marshal(req.Dependencies)
		moduleVersion.Dependencies = datatypes.JSON(dependenciesBytes)
	}

	if err := s.db.Save(moduleVersion).Error; err != nil {
		return nil, fmt.Errorf("failed to update version: %w", err)
	}

	return moduleVersion, nil
}

// PublishVersion publishes a module version making it immutable
func (s *ModuleService) PublishVersion(moduleID uint, version string) (*models.ModuleVersion, error) {
	moduleVersion, err := s.GetVersion(moduleID, version)
	if err != nil {
		return nil, err
	}

	if moduleVersion.Published {
		return nil, fmt.Errorf("version '%s' is already published", version)
	}

	// Final validation before publishing
	if err := s.linter.ValidateCompose(moduleVersion.ComposeContent); err != nil {
		return nil, fmt.Errorf("cannot publish version with invalid compose: %w", err)
	}

	moduleVersion.Publish()

	if err := s.db.Save(moduleVersion).Error; err != nil {
		return nil, fmt.Errorf("failed to publish version: %w", err)
	}

	return moduleVersion, nil
}

// ListVersions returns all versions for a module
func (s *ModuleService) ListVersions(moduleID uint, publishedOnly bool) ([]models.ModuleVersion, error) {
	// Verify module exists
	if _, err := s.GetModule(moduleID, false); err != nil {
		return nil, err
	}

	var versions []models.ModuleVersion
	query := s.db.Where("module_id = ?", moduleID)

	if publishedOnly {
		query = query.Where("published = ?", true)
	}

	if err := query.Order("created_at DESC").Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to list versions: %w", err)
	}

	return versions, nil
}