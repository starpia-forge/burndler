package services

import (
	"encoding/json"
	"fmt"

	"github.com/burndler/burndler/internal/models"
	"github.com/burndler/burndler/internal/storage"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ContainerService handles container management operations
type ContainerService struct {
	db      *gorm.DB
	storage storage.Storage
	linter  *Linter
}

// NewContainerService creates a new ContainerService instance
func NewContainerService(db *gorm.DB, storage storage.Storage, linter *Linter) *ContainerService {
	return &ContainerService{
		db:      db,
		storage: storage,
		linter:  linter,
	}
}

// CreateContainerRequest represents the request to create a container
type CreateContainerRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Repository  string `json:"repository"`
}

// UpdateContainerRequest represents the request to update a container
type UpdateContainerRequest struct {
	Description string `json:"description"`
	Author      string `json:"author"`
	Repository  string `json:"repository"`
	Active      *bool  `json:"active"`
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

// ContainerFilters represents filters for listing containers
type ContainerFilters struct {
	Active        *bool  `json:"active"`
	Author        string `json:"author"`
	PublishedOnly bool   `json:"published_only"`
	Page          int    `json:"page"`
	PageSize      int    `json:"page_size"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// CreateContainer creates a new container
func (s *ContainerService) CreateContainer(req CreateContainerRequest) (*models.Container, error) {
	// Check if container name already exists
	var existingContainer models.Container
	if err := s.db.Where("name = ?", req.Name).First(&existingContainer).Error; err == nil {
		return nil, fmt.Errorf("container with name '%s' already exists", req.Name)
	}

	container := &models.Container{
		Name:        req.Name,
		Description: req.Description,
		Author:      req.Author,
		Repository:  req.Repository,
		Active:      true,
	}

	if err := s.db.Create(container).Error; err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	return container, nil
}

// GetContainer retrieves a container by ID with optional version loading
func (s *ContainerService) GetContainer(id uint, includeVersions bool) (*models.Container, error) {
	var container models.Container
	query := s.db

	if includeVersions {
		query = query.Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		})
	}

	if err := query.First(&container, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("container not found")
		}
		return nil, fmt.Errorf("failed to get container: %w", err)
	}

	return &container, nil
}

// GetContainerByName retrieves a container by name
func (s *ContainerService) GetContainerByName(name string, includeVersions bool) (*models.Container, error) {
	var container models.Container
	query := s.db

	if includeVersions {
		query = query.Preload("Versions", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		})
	}

	if err := query.Where("name = ?", name).First(&container).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("container '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to get container: %w", err)
	}

	return &container, nil
}

// ListContainers returns a paginated list of containers
func (s *ContainerService) ListContainers(filters ContainerFilters) (*PaginatedResponse[models.Container], error) {
	var containers []models.Container
	var total int64

	query := s.db.Model(&models.Container{})

	// Apply filters
	if filters.Active != nil {
		query = query.Where("active = ?", *filters.Active)
	}

	if filters.Author != "" {
		query = query.Where("author LIKE ?", "%"+filters.Author+"%")
	}

	if filters.PublishedOnly {
		query = query.Joins("JOIN container_versions ON containers.id = container_versions.container_id").
			Where("container_versions.published = ?", true).
			Group("containers.id")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count containers: %w", err)
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
	if err := query.Offset(offset).Limit(filters.PageSize).Order("created_at DESC").Find(&containers).Error; err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	totalPages := int((total + int64(filters.PageSize) - 1) / int64(filters.PageSize))

	return &PaginatedResponse[models.Container]{
		Data:       containers,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateContainer updates an existing container
func (s *ContainerService) UpdateContainer(id uint, req UpdateContainerRequest) (*models.Container, error) {
	container, err := s.GetContainer(id, false)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Description != "" {
		container.Description = req.Description
	}
	if req.Author != "" {
		container.Author = req.Author
	}
	if req.Repository != "" {
		container.Repository = req.Repository
	}
	if req.Active != nil {
		container.Active = *req.Active
	}

	if err := s.db.Save(container).Error; err != nil {
		return nil, fmt.Errorf("failed to update container: %w", err)
	}

	return container, nil
}

// DeleteContainer soft deletes a container
func (s *ContainerService) DeleteContainer(id uint) error {
	container, err := s.GetContainer(id, true)
	if err != nil {
		return err
	}

	// Check if container has published versions
	if container.HasPublishedVersions() {
		return fmt.Errorf("cannot delete container with published versions")
	}

	if err := s.db.Delete(container).Error; err != nil {
		return fmt.Errorf("failed to delete container: %w", err)
	}

	return nil
}

// CreateVersion creates a new version for a container
func (s *ContainerService) CreateVersion(containerID uint, req CreateVersionRequest) (*models.ContainerVersion, error) {
	// Verify container exists
	container, err := s.GetContainer(containerID, false)
	if err != nil {
		return nil, err
	}

	// Check if version already exists
	var existingVersion models.ContainerVersion
	if err := s.db.Where("container_id = ? AND version = ?", containerID, req.Version).First(&existingVersion).Error; err == nil {
		return nil, fmt.Errorf("version '%s' already exists for container '%s'", req.Version, container.Name)
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

	version := &models.ContainerVersion{
		ContainerID:    containerID,
		Version:        req.Version,
		ComposeContent: req.Compose,
		Variables:      variablesJSON,
		ResourcePaths:  resourcePathsJSON,
		Dependencies:   dependenciesJSON,
		Published:      false,
	}

	if err := s.db.Create(version).Error; err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	// Load the container relationship
	if err := s.db.Preload("Container").First(version, version.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load version with container: %w", err)
	}

	return version, nil
}

// GetVersion retrieves a specific version of a container
func (s *ContainerService) GetVersion(containerID uint, version string) (*models.ContainerVersion, error) {
	var containerVersion models.ContainerVersion

	if err := s.db.Preload("Container").Where("container_id = ? AND version = ?", containerID, version).First(&containerVersion).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("version '%s' not found", version)
		}
		return nil, fmt.Errorf("failed to get version: %w", err)
	}

	return &containerVersion, nil
}

// UpdateVersion updates an existing container version (only if unpublished)
func (s *ContainerService) UpdateVersion(containerID uint, version string, req UpdateVersionRequest) (*models.ContainerVersion, error) {
	containerVersion, err := s.GetVersion(containerID, version)
	if err != nil {
		return nil, err
	}

	if !containerVersion.CanModify() {
		return nil, fmt.Errorf("cannot modify published version")
	}

	// Update fields
	if req.Compose != "" {
		// Validate compose content
		if err := s.linter.ValidateCompose(req.Compose); err != nil {
			return nil, fmt.Errorf("compose validation failed: %w", err)
		}
		containerVersion.ComposeContent = req.Compose
	}

	if req.Variables != nil {
		variablesBytes, _ := json.Marshal(req.Variables)
		containerVersion.Variables = datatypes.JSON(variablesBytes)
	}

	if req.ResourcePaths != nil {
		resourcePathsBytes, _ := json.Marshal(req.ResourcePaths)
		containerVersion.ResourcePaths = datatypes.JSON(resourcePathsBytes)
	}

	if req.Dependencies != nil {
		dependenciesBytes, _ := json.Marshal(req.Dependencies)
		containerVersion.Dependencies = datatypes.JSON(dependenciesBytes)
	}

	if err := s.db.Save(containerVersion).Error; err != nil {
		return nil, fmt.Errorf("failed to update version: %w", err)
	}

	return containerVersion, nil
}

// PublishVersion publishes a container version making it immutable
func (s *ContainerService) PublishVersion(containerID uint, version string) (*models.ContainerVersion, error) {
	containerVersion, err := s.GetVersion(containerID, version)
	if err != nil {
		return nil, err
	}

	if containerVersion.Published {
		return nil, fmt.Errorf("version '%s' is already published", version)
	}

	// Final validation before publishing
	if err := s.linter.ValidateCompose(containerVersion.ComposeContent); err != nil {
		return nil, fmt.Errorf("cannot publish version with invalid compose: %w", err)
	}

	containerVersion.Publish()

	if err := s.db.Save(containerVersion).Error; err != nil {
		return nil, fmt.Errorf("failed to publish version: %w", err)
	}

	return containerVersion, nil
}

// ListVersions returns all versions for a container
func (s *ContainerService) ListVersions(containerID uint, publishedOnly bool) ([]models.ContainerVersion, error) {
	// Verify container exists
	if _, err := s.GetContainer(containerID, false); err != nil {
		return nil, err
	}

	var versions []models.ContainerVersion
	query := s.db.Where("container_id = ?", containerID)

	if publishedOnly {
		query = query.Where("published = ?", true)
	}

	if err := query.Order("created_at DESC").Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to list versions: %w", err)
	}

	return versions, nil
}
