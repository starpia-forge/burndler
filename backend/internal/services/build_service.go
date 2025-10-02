package services

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/burndler/burndler/internal/models"
	"github.com/burndler/burndler/internal/storage"
	"gorm.io/gorm"
)

// BuildService orchestrates the build pipeline for service-based deployments
type BuildService struct {
	db                *gorm.DB
	storage           storage.Storage
	templateEngine    *TemplateEngine
	dependencyChecker *DependencyChecker
	merger            *Merger
	linter            *Linter
	packager          *Packager
}

// NewBuildService creates a new build service
func NewBuildService(
	db *gorm.DB,
	storage storage.Storage,
) *BuildService {
	return &BuildService{
		db:                db,
		storage:           storage,
		templateEngine:    NewTemplateEngine(),
		dependencyChecker: NewDependencyChecker(),
		merger:            NewMerger(),
		linter:            NewLinter(),
		packager:          NewPackager(storage),
	}
}

// BuildContext maintains state across build stages
type BuildContext struct {
	Build             *models.Build
	Service           *models.Service
	Configurations    map[uint]*models.ContainerConfiguration
	ResolvedVariables map[uint]map[string]interface{}
	RenderedFiles     map[string]string
	RenderedAssets    map[string][]byte
	DownloadAssets    []DownloadAssetInfo
	TempDirectory     string
}

// DownloadAssetInfo represents an asset that should be downloaded during installation
type DownloadAssetInfo struct {
	FilePath    string `json:"file_path"`
	DownloadURL string `json:"download_url"`
	Checksum    string `json:"checksum"`
	FileSize    int64  `json:"file_size"`
}

// ExecuteBuild executes the full build pipeline
func (bs *BuildService) ExecuteBuild(ctx context.Context, buildID string) error {
	// Load build with relationships
	var build models.Build
	if err := bs.db.Preload("Service.ServiceContainers.Container").
		Preload("Service.ServiceContainers.ContainerVersion").
		Where("id = ?", buildID).
		First(&build).Error; err != nil {
		return fmt.Errorf("failed to load build: %w", err)
	}

	// Verify service exists
	if build.Service == nil {
		return fmt.Errorf("build %s is not a service build", buildID)
	}

	// Initialize build context
	buildCtx := &BuildContext{
		Build:             &build,
		Service:           build.Service,
		Configurations:    make(map[uint]*models.ContainerConfiguration),
		ResolvedVariables: make(map[uint]map[string]interface{}),
		RenderedFiles:     make(map[string]string),
		RenderedAssets:    make(map[string][]byte),
		DownloadAssets:    make([]DownloadAssetInfo, 0),
	}

	// Execute build stages
	stages := []struct {
		name string
		fn   func(context.Context, *BuildContext) error
	}{
		{"validation", bs.validateConfiguration},
		{"configuration", bs.resolveConfiguration},
		{"template_render", bs.renderTemplates},
		{"asset_resolution", bs.resolveAssets},
		{"compose_merge", bs.mergeCompose},
		{"linting", bs.lintCompose},
		{"packaging", bs.packageInstaller},
	}

	for _, stage := range stages {
		build.Status = fmt.Sprintf("building:%s", stage.name)
		if err := bs.updateBuildStatus(&build); err != nil {
			return err
		}

		if err := stage.fn(ctx, buildCtx); err != nil {
			build.Status = "failed"
			build.Error = err.Error()
			_ = bs.updateBuildStatus(&build)
			return fmt.Errorf("stage %s failed: %w", stage.name, err)
		}
	}

	build.Status = "completed"
	return bs.updateBuildStatus(&build)
}

// validateConfiguration performs pre-flight checks
func (bs *BuildService) validateConfiguration(ctx context.Context, buildCtx *BuildContext) error {
	// Check service is active
	if !buildCtx.Service.Active {
		return fmt.Errorf("service is not active")
	}

	// Check has enabled containers
	enabledCount := 0
	for _, sc := range buildCtx.Service.ServiceContainers {
		if sc.Enabled {
			enabledCount++
		}
	}

	if enabledCount == 0 {
		return fmt.Errorf("service has no enabled containers")
	}

	return nil
}

// resolveConfiguration loads and validates all configurations
func (bs *BuildService) resolveConfiguration(ctx context.Context, buildCtx *BuildContext) error {
	for _, sc := range buildCtx.Service.ServiceContainers {
		if !sc.Enabled {
			continue
		}

		// Load container version to get configuration ID (Phase 3 - Container-level configuration)
		var version models.ContainerVersion
		if err := bs.db.First(&version, sc.ContainerVersionID).Error; err != nil {
			return fmt.Errorf("failed to load container version %d: %w", sc.ContainerVersionID, err)
		}

		// Check if version has a configuration
		if version.ConfigurationID == nil {
			// No configuration defined, skip
			continue
		}

		// Load container configuration via ConfigurationID
		var config models.ContainerConfiguration
		if err := bs.db.Where("id = ?", *version.ConfigurationID).
			Preload("Files").
			Preload("Assets").
			First(&config).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Configuration referenced but not found, skip
				continue
			}
			return fmt.Errorf("failed to load configuration for container %d: %w", sc.ContainerID, err)
		}

		buildCtx.Configurations[sc.ContainerID] = &config

		// Resolve variables
		variables := bs.resolveVariables(buildCtx.Service, &sc, &config)
		buildCtx.ResolvedVariables[sc.ContainerID] = variables

		// Validate dependencies
		if len(config.DependencyRules) > 0 {
			var rules []DependencyRule
			if err := json.Unmarshal(config.DependencyRules, &rules); err != nil {
				return fmt.Errorf("failed to parse dependency rules for container %d: %w", sc.ContainerID, err)
			}

			errors := bs.dependencyChecker.ValidateConfiguration(rules, variables)
			if len(errors) > 0 {
				return fmt.Errorf("dependency validation failed for container %d: %v", sc.ContainerID, errors)
			}
		}
	}

	return nil
}

// renderTemplates renders all template files
func (bs *BuildService) renderTemplates(ctx context.Context, buildCtx *BuildContext) error {
	for containerID, config := range buildCtx.Configurations {
		variables := buildCtx.ResolvedVariables[containerID]

		// Render template files
		for _, file := range config.Files {
			if file.FileType != "template" {
				continue
			}

			// Load template content from storage
			content, err := bs.loadFileContent(ctx, file.StoragePath)
			if err != nil {
				return fmt.Errorf("failed to load template %s: %w", file.FilePath, err)
			}

			// Render template
			rendered, err := bs.templateEngine.Render(
				string(content),
				file.TemplateFormat,
				variables,
			)
			if err != nil {
				return fmt.Errorf("failed to render template %s: %w", file.FilePath, err)
			}

			// Store rendered content
			namespacedPath := bs.applyNamespace(file.FilePath, containerID, buildCtx.Service)
			buildCtx.RenderedFiles[namespacedPath] = rendered
		}

		// Copy static files
		for _, file := range config.Files {
			if file.FileType == "static" {
				content, err := bs.loadFileContent(ctx, file.StoragePath)
				if err != nil {
					return fmt.Errorf("failed to load static file %s: %w", file.FilePath, err)
				}

				namespacedPath := bs.applyNamespace(file.FilePath, containerID, buildCtx.Service)
				buildCtx.RenderedFiles[namespacedPath] = string(content)
			}
		}
	}

	return nil
}

// resolveAssets resolves which assets to include
func (bs *BuildService) resolveAssets(ctx context.Context, buildCtx *BuildContext) error {
	for containerID, config := range buildCtx.Configurations {
		variables := buildCtx.ResolvedVariables[containerID]

		for _, asset := range config.Assets {
			// Evaluate include condition
			if asset.IncludeCondition != "" {
				include, err := bs.dependencyChecker.EvaluateCondition(asset.IncludeCondition, variables)
				if err != nil {
					return fmt.Errorf("failed to evaluate asset condition %s: %w", asset.IncludeCondition, err)
				}
				if !include {
					continue // Skip this asset
				}
			}

			// Handle based on storage type
			switch asset.StorageType {
			case "embedded":
				// Load asset content
				content, err := bs.loadFileContent(ctx, asset.StoragePath)
				if err != nil {
					return fmt.Errorf("failed to load embedded asset %s: %w", asset.FilePath, err)
				}

				// Store in rendered assets
				namespacedPath := bs.applyNamespace(asset.FilePath, containerID, buildCtx.Service)
				buildCtx.RenderedAssets[namespacedPath] = content

			case "download":
				// Track download asset for manifest
				namespacedPath := bs.applyNamespace(asset.FilePath, containerID, buildCtx.Service)

				// Generate download URL from storage path
				downloadURL := asset.DownloadURL
				if downloadURL == "" {
					// If no download URL is set, use the storage path as reference
					downloadURL = fmt.Sprintf("/api/v1/assets/download?path=%s", asset.StoragePath)
				}

				downloadInfo := DownloadAssetInfo{
					FilePath:    namespacedPath,
					DownloadURL: downloadURL,
					Checksum:    asset.Checksum,
					FileSize:    asset.FileSize,
				}

				buildCtx.DownloadAssets = append(buildCtx.DownloadAssets, downloadInfo)
			}
		}
	}

	return nil
}

// mergeCompose merges all container compose files
func (bs *BuildService) mergeCompose(ctx context.Context, buildCtx *BuildContext) error {
	// Prepare modules for merging
	modules := []Module{}

	for _, sc := range buildCtx.Service.ServiceContainers {
		if !sc.Enabled {
			continue
		}

		// Get container compose content
		var containerVersion models.ContainerVersion
		if err := bs.db.First(&containerVersion, sc.ContainerVersionID).Error; err != nil {
			return fmt.Errorf("failed to load container version %d: %w", sc.ContainerVersionID, err)
		}

		// Get variables for this container
		variables := make(map[string]string)
		if resolvedVars, ok := buildCtx.ResolvedVariables[sc.ContainerID]; ok {
			for k, v := range resolvedVars {
				variables[k] = fmt.Sprintf("%v", v)
			}
		}

		// Create namespace
		var container models.Container
		bs.db.First(&container, sc.ContainerID)
		namespace := fmt.Sprintf("%s_%d__%s", buildCtx.Service.Name, buildCtx.Service.ID, container.Name)

		modules = append(modules, Module{
			Name:      namespace,
			Compose:   containerVersion.ComposeContent,
			Variables: variables,
		})
	}

	// Get service-level variables
	serviceVariables := make(map[string]string)
	if buildCtx.Service.Variables != nil {
		var serviceVars map[string]interface{}
		if err := json.Unmarshal(buildCtx.Service.Variables, &serviceVars); err == nil {
			for k, v := range serviceVars {
				serviceVariables[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	// Merge compose files
	mergeReq := &MergeRequest{
		Modules:          modules,
		ServiceVariables: serviceVariables,
	}

	result, err := bs.merger.Merge(mergeReq)
	if err != nil {
		return fmt.Errorf("failed to merge compose files: %w", err)
	}

	// Store merged compose
	buildCtx.Build.ComposeYAML = result.MergedCompose

	return nil
}

// lintCompose validates the merged compose file
func (bs *BuildService) lintCompose(ctx context.Context, buildCtx *BuildContext) error {
	lintReq := &LintRequest{
		Compose:    buildCtx.Build.ComposeYAML,
		StrictMode: true,
	}

	result, err := bs.linter.Lint(lintReq)
	if err != nil {
		return fmt.Errorf("failed to lint compose: %w", err)
	}

	if !result.Valid {
		return fmt.Errorf("compose validation failed with %d errors", len(result.Errors))
	}

	return nil
}

// packageInstaller creates the final installer package
func (bs *BuildService) packageInstaller(ctx context.Context, buildCtx *BuildContext) error {
	// Prepare resource files
	resourceFiles := make([]ResourceFile, 0)

	// Add rendered template files
	for path, content := range buildCtx.RenderedFiles {
		resourceFiles = append(resourceFiles, ResourceFile{
			Path:    filepath.Join("resources", path),
			Content: []byte(content),
		})
	}

	// Add rendered assets
	for path, content := range buildCtx.RenderedAssets {
		resourceFiles = append(resourceFiles, ResourceFile{
			Path:    filepath.Join("resources", path),
			Content: content,
		})
	}

	packageReq := &PackageRequest{
		Name:           fmt.Sprintf("%s-%s", buildCtx.Service.Name, buildCtx.Build.ID.String()),
		Compose:        buildCtx.Build.ComposeYAML,
		Resources:      resourceFiles,
		DownloadAssets: buildCtx.DownloadAssets,
	}

	url, err := bs.packager.CreatePackage(ctx, packageReq)
	if err != nil {
		return fmt.Errorf("failed to create package: %w", err)
	}

	// Update build with download URL
	buildCtx.Build.DownloadURL = url

	return nil
}

// resolveVariables resolves variables with proper precedence
func (bs *BuildService) resolveVariables(
	service *models.Service,
	serviceContainer *models.ServiceContainer,
	config *models.ContainerConfiguration,
) map[string]interface{} {
	variables := make(map[string]interface{})

	// Global variables
	variables["SERVICE_NAME"] = service.Name
	variables["SERVICE_ID"] = service.ID

	// Service variables
	if service.Variables != nil {
		var serviceVars map[string]interface{}
		if err := json.Unmarshal(service.Variables, &serviceVars); err == nil {
			for k, v := range serviceVars {
				variables[k] = v
			}
		}
	}

	// Container overrides (highest precedence)
	effectiveVars := serviceContainer.GetEffectiveVariables()
	for k, v := range effectiveVars {
		variables[k] = v
	}

	return variables
}

// applyNamespace applies namespace prefix to file path
func (bs *BuildService) applyNamespace(
	filePath string,
	containerID uint,
	service *models.Service,
) string {
	var container models.Container
	bs.db.First(&container, containerID)

	namespace := fmt.Sprintf("%s_%d", service.Name, service.ID)
	return filepath.Join(namespace, container.Name, filePath)
}

// loadFileContent loads file content from storage
func (bs *BuildService) loadFileContent(ctx context.Context, storagePath string) ([]byte, error) {
	reader, err := bs.storage.Download(ctx, storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to download from storage: %w", err)
	}
	defer func() { _ = reader.Close() }()

	// Read all content
	var content []byte
	buf := make([]byte, 4096)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			content = append(content, buf[:n]...)
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("failed to read content: %w", err)
		}
	}

	return content, nil
}

// updateBuildStatus updates build status in database
func (bs *BuildService) updateBuildStatus(build *models.Build) error {
	return bs.db.Save(build).Error
}

// EvaluateCondition is a helper to evaluate template conditions
func (bs *BuildService) EvaluateCondition(condition string, variables map[string]interface{}) (bool, error) {
	// For now, use template engine's condition evaluation
	// This is a simplified version - real implementation would use a proper expression evaluator

	// Replace template variables
	expr := condition
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		expr = strings.ReplaceAll(expr, placeholder, fmt.Sprintf("%v", value))
	}

	// Simple boolean evaluation
	expr = strings.TrimSpace(expr)
	if expr == "true" {
		return true, nil
	}
	if expr == "false" {
		return false, nil
	}

	// For more complex expressions, use dependency checker
	return bs.dependencyChecker.EvaluateCondition(condition, variables)
}