package services

import (
	"testing"

	"github.com/burndler/burndler/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupProjectServiceTest(t *testing.T) (*ProjectService, *ModuleService, *gorm.DB) {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(&models.Module{}, &models.ModuleVersion{}, &models.Project{}, &models.ProjectModule{})
	require.NoError(t, err)

	// Create storage and linter
	store := &mockStorage{} // Mock storage
	linter := NewLinter()

	moduleService := NewModuleService(db, store, linter)
	projectService := NewProjectService(db, moduleService)

	return projectService, moduleService, db
}

func TestProjectService_CreateProject(t *testing.T) {
	service, _, _ := setupProjectServiceTest(t)

	tests := []struct {
		name        string
		userID      uint
		req         CreateProjectRequest
		expectError bool
	}{
		{
			name:   "successful project creation",
			userID: 1,
			req: CreateProjectRequest{
				Name:        "my-webapp",
				Description: "My web application project",
			},
			expectError: false,
		},
		{
			name:   "duplicate project name for same user",
			userID: 1,
			req: CreateProjectRequest{
				Name:        "my-webapp", // Same name as above
				Description: "Another webapp",
			},
			expectError: true,
		},
		{
			name:   "same project name for different user",
			userID: 2,
			req: CreateProjectRequest{
				Name:        "my-webapp", // Same name but different user
				Description: "User 2's webapp",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, err := service.CreateProject(tt.userID, tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, project)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, project)
				assert.Equal(t, tt.req.Name, project.Name)
				assert.Equal(t, tt.req.Description, project.Description)
				assert.Equal(t, tt.userID, project.UserID)
			}
		})
	}
}

func TestProjectService_GetProject(t *testing.T) {
	service, _, db := setupProjectServiceTest(t)

	// Create test project
	project := &models.Project{
		Name:        "test-project",
		Description: "Test project",
		UserID:      1,
	}
	require.NoError(t, db.Create(project).Error)

	tests := []struct {
		name           string
		projectID      uint
		includeModules bool
		expectError    bool
	}{
		{
			name:           "get existing project",
			projectID:      project.ID,
			includeModules: false,
			expectError:    false,
		},
		{
			name:           "get non-existent project",
			projectID:      999,
			includeModules: false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetProject(tt.projectID, tt.includeModules)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, project.Name, result.Name)
			}
		})
	}
}

func TestProjectService_GetProjectByName(t *testing.T) {
	service, _, db := setupProjectServiceTest(t)

	// Create test projects
	project1 := &models.Project{
		Name:   "webapp",
		UserID: 1,
	}
	project2 := &models.Project{
		Name:   "webapp", // Same name, different user
		UserID: 2,
	}
	require.NoError(t, db.Create(project1).Error)
	require.NoError(t, db.Create(project2).Error)

	tests := []struct {
		name        string
		userID      uint
		projectName string
		expectError bool
		expectedID  uint
	}{
		{
			name:        "get project for user 1",
			userID:      1,
			projectName: "webapp",
			expectError: false,
			expectedID:  project1.ID,
		},
		{
			name:        "get project for user 2",
			userID:      2,
			projectName: "webapp",
			expectError: false,
			expectedID:  project2.ID,
		},
		{
			name:        "get non-existent project",
			userID:      1,
			projectName: "nonexistent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetProjectByName(tt.userID, tt.projectName, false)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedID, result.ID)
				assert.Equal(t, tt.projectName, result.Name)
			}
		})
	}
}

func TestProjectService_ListProjects(t *testing.T) {
	service, _, db := setupProjectServiceTest(t)

	// Create test projects
	projects := []*models.Project{
		{Name: "project1", UserID: 1},
		{Name: "project2", UserID: 1},
		{Name: "project3", UserID: 2},
	}

	for _, p := range projects {
		require.NoError(t, db.Create(p).Error)
	}

	tests := []struct {
		name          string
		filters       ProjectFilters
		expectedCount int
		expectedTotal int64
	}{
		{
			name:          "list all projects",
			filters:       ProjectFilters{Page: 1, PageSize: 10},
			expectedCount: 3,
			expectedTotal: 3,
		},
		{
			name:          "filter by user ID",
			filters:       ProjectFilters{UserID: 1, Page: 1, PageSize: 10},
			expectedCount: 2,
			expectedTotal: 2,
		},
		{
			name:          "filter by name",
			filters:       ProjectFilters{Name: "project1", Page: 1, PageSize: 10},
			expectedCount: 1,
			expectedTotal: 1,
		},
		{
			name:          "pagination test",
			filters:       ProjectFilters{Page: 1, PageSize: 2},
			expectedCount: 2,
			expectedTotal: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ListProjects(tt.filters)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Len(t, result.Data, tt.expectedCount)
			assert.Equal(t, tt.expectedTotal, result.Total)
		})
	}
}

func TestProjectService_UpdateProject(t *testing.T) {
	service, _, db := setupProjectServiceTest(t)

	// Create test project
	project := &models.Project{
		Name:        "original-name",
		Description: "Original description",
		UserID:      1,
	}
	require.NoError(t, db.Create(project).Error)

	// Create another project for conflict testing
	conflictProject := &models.Project{
		Name:   "conflict-name",
		UserID: 1,
	}
	require.NoError(t, db.Create(conflictProject).Error)

	tests := []struct {
		name        string
		projectID   uint
		req         UpdateProjectRequest
		expectError bool
	}{
		{
			name:      "successful update",
			projectID: project.ID,
			req: UpdateProjectRequest{
				Name:        "updated-name",
				Description: "Updated description",
			},
			expectError: false,
		},
		{
			name:      "name conflict",
			projectID: project.ID,
			req: UpdateProjectRequest{
				Name: "conflict-name", // Conflicts with existing project
			},
			expectError: true,
		},
		{
			name:        "update non-existent project",
			projectID:   999,
			req:         UpdateProjectRequest{Name: "new-name"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.UpdateProject(tt.projectID, tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.req.Name != "" {
					assert.Equal(t, tt.req.Name, result.Name)
				}
				if tt.req.Description != "" {
					assert.Equal(t, tt.req.Description, result.Description)
				}
			}
		})
	}
}

func TestProjectService_DeleteProject(t *testing.T) {
	service, _, db := setupProjectServiceTest(t)

	// Create test project
	project := &models.Project{
		Name:   "test-project",
		UserID: 1,
	}
	require.NoError(t, db.Create(project).Error)

	tests := []struct {
		name        string
		projectID   uint
		expectError bool
	}{
		{
			name:        "delete existing project",
			projectID:   project.ID,
			expectError: false,
		},
		{
			name:        "delete non-existent project",
			projectID:   999,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteProject(tt.projectID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify project is deleted
				var deletedProject models.Project
				err := db.First(&deletedProject, tt.projectID).Error
				assert.Error(t, err)
				assert.Equal(t, gorm.ErrRecordNotFound, err)
			}
		})
	}
}

func TestProjectService_AddModuleToProject(t *testing.T) {
	service, _, db := setupProjectServiceTest(t)

	// Create test project
	project := &models.Project{
		Name:   "test-project",
		UserID: 1,
	}
	require.NoError(t, db.Create(project).Error)

	// Create test module with published version
	module := &models.Module{
		Name:   "test-module",
		Active: true,
	}
	require.NoError(t, db.Create(module).Error)

	publishedVersion := &models.ModuleVersion{
		ModuleID:       module.ID,
		Version:        "v1.0.0",
		ComposeContent: "version: '3.8'\nservices:\n  web:\n    image: nginx:latest",
		Published:      true,
	}
	require.NoError(t, db.Create(publishedVersion).Error)

	unpublishedVersion := &models.ModuleVersion{
		ModuleID:       module.ID,
		Version:        "v1.1.0",
		ComposeContent: "version: '3.8'\nservices:\n  web:\n    image: nginx:latest",
		Published:      false,
	}
	require.NoError(t, db.Create(unpublishedVersion).Error)

	tests := []struct {
		name        string
		projectID   uint
		req         AddModuleToProjectRequest
		expectError bool
	}{
		{
			name:      "successful module addition",
			projectID: project.ID,
			req: AddModuleToProjectRequest{
				ModuleID:        module.ID,
				ModuleVersionID: publishedVersion.ID,
				Order:           1,
				Enabled:         true,
				OverrideVars:    map[string]interface{}{"port": 8080},
			},
			expectError: false,
		},
		{
			name:      "add unpublished version",
			projectID: project.ID,
			req: AddModuleToProjectRequest{
				ModuleID:        module.ID,
				ModuleVersionID: unpublishedVersion.ID,
				Enabled:         true,
			},
			expectError: true,
		},
		{
			name:      "add duplicate module",
			projectID: project.ID,
			req: AddModuleToProjectRequest{
				ModuleID:        module.ID,
				ModuleVersionID: publishedVersion.ID,
				Enabled:         true,
			},
			expectError: true,
		},
		{
			name:      "add to non-existent project",
			projectID: 999,
			req: AddModuleToProjectRequest{
				ModuleID:        module.ID,
				ModuleVersionID: publishedVersion.ID,
				Enabled:         true,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.AddModuleToProject(tt.projectID, tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.req.ModuleID, result.ModuleID)
				assert.Equal(t, tt.req.ModuleVersionID, result.ModuleVersionID)
				assert.Equal(t, tt.req.Enabled, result.Enabled)
			}
		})
	}
}

func TestProjectService_GetProjectModules(t *testing.T) {
	service, _, db := setupProjectServiceTest(t)

	// Create test project
	project := &models.Project{
		Name:   "test-project",
		UserID: 1,
	}
	require.NoError(t, db.Create(project).Error)

	// Create test modules
	module1 := &models.Module{Name: "module1", Active: true}
	module2 := &models.Module{Name: "module2", Active: true}
	require.NoError(t, db.Create(module1).Error)
	require.NoError(t, db.Create(module2).Error)

	version1 := &models.ModuleVersion{
		ModuleID: module1.ID,
		Version:  "v1.0.0",
		ComposeContent: "version: '3.8'\nservices:\n  web:\n    image: nginx:latest",
		Published: true,
	}
	version2 := &models.ModuleVersion{
		ModuleID: module2.ID,
		Version:  "v1.0.0",
		ComposeContent: "version: '3.8'\nservices:\n  app:\n    image: node:latest",
		Published: true,
	}
	require.NoError(t, db.Create(version1).Error)
	require.NoError(t, db.Create(version2).Error)

	// Create project modules in specific order
	projectModule1 := &models.ProjectModule{
		ProjectID:       project.ID,
		ModuleID:        module1.ID,
		ModuleVersionID: version1.ID,
		Order:           2,
		Enabled:         true,
	}
	projectModule2 := &models.ProjectModule{
		ProjectID:       project.ID,
		ModuleID:        module2.ID,
		ModuleVersionID: version2.ID,
		Order:           1,
		Enabled:         true,
	}
	require.NoError(t, db.Create(projectModule1).Error)
	require.NoError(t, db.Create(projectModule2).Error)

	// Test getting project modules
	modules, err := service.GetProjectModules(project.ID)
	assert.NoError(t, err)
	assert.Len(t, modules, 2)

	// Verify order (module2 should come first with order=1)
	assert.Equal(t, module2.ID, modules[0].ModuleID)
	assert.Equal(t, module1.ID, modules[1].ModuleID)
	assert.Equal(t, 1, modules[0].Order)
	assert.Equal(t, 2, modules[1].Order)
}

func TestProjectService_ReorderProjectModules(t *testing.T) {
	service, _, db := setupProjectServiceTest(t)

	// Create test project and modules
	project := &models.Project{Name: "test-project", UserID: 1}
	require.NoError(t, db.Create(project).Error)

	module1 := &models.Module{Name: "module1", Active: true}
	module2 := &models.Module{Name: "module2", Active: true}
	require.NoError(t, db.Create(module1).Error)
	require.NoError(t, db.Create(module2).Error)

	version1 := &models.ModuleVersion{
		ModuleID: module1.ID,
		Version:  "v1.0.0",
		ComposeContent: "version: '3.8'\nservices:\n  web:\n    image: nginx:latest",
		Published: true,
	}
	version2 := &models.ModuleVersion{
		ModuleID: module2.ID,
		Version:  "v1.0.0",
		ComposeContent: "version: '3.8'\nservices:\n  app:\n    image: node:latest",
		Published: true,
	}
	require.NoError(t, db.Create(version1).Error)
	require.NoError(t, db.Create(version2).Error)

	projectModule1 := &models.ProjectModule{
		ProjectID:       project.ID,
		ModuleID:        module1.ID,
		ModuleVersionID: version1.ID,
		Order:           1,
		Enabled:         true,
	}
	projectModule2 := &models.ProjectModule{
		ProjectID:       project.ID,
		ModuleID:        module2.ID,
		ModuleVersionID: version2.ID,
		Order:           2,
		Enabled:         true,
	}
	require.NoError(t, db.Create(projectModule1).Error)
	require.NoError(t, db.Create(projectModule2).Error)

	// Reorder modules
	newOrders := map[uint]int{
		module1.ID: 3,
		module2.ID: 1,
	}

	err := service.ReorderProjectModules(project.ID, newOrders)
	assert.NoError(t, err)

	// Verify new order
	modules, err := service.GetProjectModules(project.ID)
	assert.NoError(t, err)
	assert.Len(t, modules, 2)

	// module2 should be first with order=1, module1 should be second with order=3
	assert.Equal(t, module2.ID, modules[0].ModuleID)
	assert.Equal(t, module1.ID, modules[1].ModuleID)
	assert.Equal(t, 1, modules[0].Order)
	assert.Equal(t, 3, modules[1].Order)
}