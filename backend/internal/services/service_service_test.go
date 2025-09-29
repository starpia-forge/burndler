package services

import (
	"testing"

	"github.com/burndler/burndler/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupServiceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.Container{},
		&models.ContainerVersion{},
		&models.Service{},
		&models.ServiceContainer{},
	)
	assert.NoError(t, err)

	return db
}

func TestServiceService_CreateService(t *testing.T) {
	db := setupServiceTestDB(t)
	service := NewServiceService(db, nil)

	// Create test user
	user := &models.User{
		Email: "test@example.com",
		Name:  "testuser",
		Role:  "Developer",
	}
	err := db.Create(user).Error
	assert.NoError(t, err)

	tests := []struct {
		name    string
		req     CreateServiceRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid service creation",
			req: CreateServiceRequest{
				Name:        "test-service",
				Description: "Test service description",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			req: CreateServiceRequest{
				Name:        "",
				Description: "Test service description",
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "duplicate name for same user",
			req: CreateServiceRequest{
				Name:        "test-service",
				Description: "Duplicate service",
			},
			wantErr: true,
			errMsg:  "service with name 'test-service' already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CreateService(user.ID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.req.Name, result.Name)
				assert.Equal(t, tt.req.Description, result.Description)
				assert.Equal(t, user.ID, result.UserID)
				assert.True(t, result.Active)
			}
		})
	}
}

func TestServiceService_GetService(t *testing.T) {
	db := setupServiceTestDB(t)
	service := NewServiceService(db, nil)

	// Create test user
	user := &models.User{
		Email: "test@example.com",
		Name:  "testuser",
		Role:  "Developer",
	}
	err := db.Create(user).Error
	assert.NoError(t, err)

	// Create test service
	testService := &models.Service{
		Name:        "test-service",
		Description: "Test service",
		UserID:      user.ID,
		Active:      true,
	}
	err = db.Create(testService).Error
	assert.NoError(t, err)

	tests := []struct {
		name              string
		serviceID         uint
		includeContainers bool
		wantErr           bool
		errMsg            string
	}{
		{
			name:              "existing service without containers",
			serviceID:         testService.ID,
			includeContainers: false,
			wantErr:           false,
		},
		{
			name:              "existing service with containers",
			serviceID:         testService.ID,
			includeContainers: true,
			wantErr:           false,
		},
		{
			name:              "non-existing service",
			serviceID:         999,
			includeContainers: false,
			wantErr:           true,
			errMsg:            "service not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetService(tt.serviceID, tt.includeContainers)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.serviceID, result.ID)
			}
		})
	}
}

func TestServiceService_ListServices(t *testing.T) {
	db := setupServiceTestDB(t)
	service := NewServiceService(db, nil)

	// Create test users
	user1 := &models.User{
		Email: "user1@example.com",
		Name:  "user1",
		Role:  "Developer",
	}
	user2 := &models.User{
		Email: "user2@example.com",
		Name:  "user2",
		Role:  "Engineer",
	}
	err := db.Create([]*models.User{user1, user2}).Error
	assert.NoError(t, err)

	// Create test services
	service1 := &models.Service{Name: "service1", UserID: user1.ID, Active: true}
	service2 := &models.Service{Name: "service2", UserID: user1.ID, Active: true}
	service3 := &models.Service{Name: "service3", UserID: user2.ID, Active: true}

	err = db.Create(service1).Error
	assert.NoError(t, err)
	err = db.Create(service2).Error
	assert.NoError(t, err)
	// Manually set service2 to inactive after creation
	err = db.Model(service2).Update("active", false).Error
	assert.NoError(t, err)
	err = db.Create(service3).Error
	assert.NoError(t, err)

	tests := []struct {
		name     string
		filters  ServiceFilters
		expected int
	}{
		{
			name: "all services",
			filters: ServiceFilters{
				Page:     1,
				PageSize: 10,
			},
			expected: 3,
		},
		{
			name: "active services only",
			filters: ServiceFilters{
				Page:     1,
				PageSize: 10,
				Active:   &[]bool{true}[0],
			},
			expected: 2,
		},
		{
			name: "user1 services only",
			filters: ServiceFilters{
				Page:     1,
				PageSize: 10,
				UserID:   user1.ID,
			},
			expected: 2,
		},
		{
			name: "pagination test",
			filters: ServiceFilters{
				Page:     1,
				PageSize: 2,
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ListServices(tt.filters)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Len(t, result.Data, tt.expected)
		})
	}
}

func TestServiceService_UpdateService(t *testing.T) {
	db := setupServiceTestDB(t)
	service := NewServiceService(db, nil)

	// Create test user
	user := &models.User{
		Email: "test@example.com",
		Name:  "testuser",
		Role:  "Developer",
	}
	err := db.Create(user).Error
	assert.NoError(t, err)

	// Create test service
	testService := &models.Service{
		Name:        "original-name",
		Description: "Original description",
		UserID:      user.ID,
		Active:      true,
	}
	err = db.Create(testService).Error
	assert.NoError(t, err)

	tests := []struct {
		name    string
		req     UpdateServiceRequest
		wantErr bool
	}{
		{
			name: "update description",
			req: UpdateServiceRequest{
				Description: &[]string{"Updated description"}[0],
			},
			wantErr: false,
		},
		{
			name: "update name",
			req: UpdateServiceRequest{
				Name: &[]string{"updated-name"}[0],
			},
			wantErr: false,
		},
		{
			name: "update active status",
			req: UpdateServiceRequest{
				Active: &[]bool{false}[0],
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.UpdateService(testService.ID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				if tt.req.Name != nil {
					assert.Equal(t, *tt.req.Name, result.Name)
				}
				if tt.req.Description != nil {
					assert.Equal(t, *tt.req.Description, result.Description)
				}
				if tt.req.Active != nil {
					assert.Equal(t, *tt.req.Active, result.Active)
				}
			}
		})
	}
}

func TestServiceService_DeleteService(t *testing.T) {
	db := setupServiceTestDB(t)
	service := NewServiceService(db, nil)

	// Create test user
	user := &models.User{
		Email: "test@example.com",
		Name:  "testuser",
		Role:  "Developer",
	}
	err := db.Create(user).Error
	assert.NoError(t, err)

	// Create test service
	testService := &models.Service{
		Name:        "test-service",
		Description: "Test service",
		UserID:      user.ID,
		Active:      true,
	}
	err = db.Create(testService).Error
	assert.NoError(t, err)

	tests := []struct {
		name      string
		serviceID uint
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "delete existing service",
			serviceID: testService.ID,
			wantErr:   false,
		},
		{
			name:      "delete non-existing service",
			serviceID: 999,
			wantErr:   true,
			errMsg:    "service not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteService(tt.serviceID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)

				// Verify service is soft deleted
				var deletedService models.Service
				err = db.Unscoped().First(&deletedService, tt.serviceID).Error
				assert.NoError(t, err)
				assert.NotNil(t, deletedService.DeletedAt)
			}
		})
	}
}