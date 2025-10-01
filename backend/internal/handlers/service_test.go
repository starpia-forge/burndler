package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/burndler/burndler/internal/models"
	"github.com/burndler/burndler/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupServiceHandlerTest(t *testing.T) (*gorm.DB, *ServiceHandler) {
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

	serviceService := services.NewServiceService(db, nil)
	buildService := services.NewBuildService(db, nil)
	handler := NewServiceHandler(serviceService, buildService, db)

	return db, handler
}

func createTestUser(t *testing.T, db *gorm.DB, role string) *models.User {
	user := &models.User{
		Email: fmt.Sprintf("test-%s@example.com", role),
		Name:  fmt.Sprintf("test-%s", role),
		Role:  role,
	}
	err := db.Create(user).Error
	assert.NoError(t, err)
	return user
}

func TestServiceHandler_CreateService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, handler := setupServiceHandlerTest(t)

	user := createTestUser(t, db, "Developer")

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid service creation",
			requestBody: CreateServiceRequest{
				Name:        "test-service",
				Description: "Test service description",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "empty name",
			requestBody: CreateServiceRequest{
				Name:        "",
				Description: "Test service description",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid-json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req, err := http.NewRequest("POST", "/services", bytes.NewBuffer(body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router := gin.New()

			// Mock middleware to set user context (matching JWT middleware format)
			router.Use(func(c *gin.Context) {
				c.Set("user_id", strconv.Itoa(int(user.ID)))
				c.Set("email", user.Email)
				c.Set("role", user.Role)
				c.Next()
			})

			router.POST("/services", handler.CreateService)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response models.Service
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.requestBody.(CreateServiceRequest).Name, response.Name)
				assert.Equal(t, tt.requestBody.(CreateServiceRequest).Description, response.Description)
				assert.Equal(t, user.ID, response.UserID)
			}
		})
	}
}

func TestServiceHandler_GetService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, handler := setupServiceHandlerTest(t)

	user := createTestUser(t, db, "Developer")

	// Create test service
	testService := &models.Service{
		Name:        "test-service",
		Description: "Test service",
		UserID:      user.ID,
		Active:      true,
	}
	err := db.Create(testService).Error
	assert.NoError(t, err)

	tests := []struct {
		name           string
		serviceID      string
		expectedStatus int
	}{
		{
			name:           "existing service",
			serviceID:      fmt.Sprintf("%d", testService.ID),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existing service",
			serviceID:      "999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid service ID",
			serviceID:      "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/services/"+tt.serviceID, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router := gin.New()

			// Mock middleware to set user context (matching JWT middleware format)
			router.Use(func(c *gin.Context) {
				c.Set("user_id", strconv.Itoa(int(user.ID)))
				c.Set("email", user.Email)
				c.Set("role", user.Role)
				c.Next()
			})

			router.GET("/services/:id", handler.GetService)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response models.Service
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, testService.ID, response.ID)
				assert.Equal(t, testService.Name, response.Name)
			}
		})
	}
}

func TestServiceHandler_ListServices(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, handler := setupServiceHandlerTest(t)

	user := createTestUser(t, db, "Developer")

	// Create test services
	service1 := &models.Service{Name: "service1", UserID: user.ID, Active: true}
	service2 := &models.Service{Name: "service2", UserID: user.ID, Active: true}

	err := db.Create([]*models.Service{service1, service2}).Error
	assert.NoError(t, err)

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "list all services",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "list with pagination",
			queryParams:    "?page=1&page_size=1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "list active services",
			queryParams:    "?active=true",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/services"+tt.queryParams, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router := gin.New()

			// Mock middleware to set user context (matching JWT middleware format)
			router.Use(func(c *gin.Context) {
				c.Set("user_id", strconv.Itoa(int(user.ID)))
				c.Set("email", user.Email)
				c.Set("role", user.Role)
				c.Next()
			})

			router.GET("/services", handler.ListServices)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response services.PaginatedResponse[models.Service]
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Len(t, response.Data, tt.expectedCount)
		})
	}
}

func TestServiceHandler_UpdateService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, handler := setupServiceHandlerTest(t)

	user := createTestUser(t, db, "Developer")

	// Create test service
	testService := &models.Service{
		Name:        "original-name",
		Description: "Original description",
		UserID:      user.ID,
		Active:      true,
	}
	err := db.Create(testService).Error
	assert.NoError(t, err)

	tests := []struct {
		name           string
		serviceID      string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name:      "valid update",
			serviceID: fmt.Sprintf("%d", testService.ID),
			requestBody: UpdateServiceRequest{
				Description: &[]string{"Updated description"}[0],
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existing service",
			serviceID:      "999",
			requestBody:    UpdateServiceRequest{},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid service ID",
			serviceID:      "invalid",
			requestBody:    UpdateServiceRequest{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			req, err := http.NewRequest("PUT", "/services/"+tt.serviceID, bytes.NewBuffer(body))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router := gin.New()

			// Mock middleware to set user context (matching JWT middleware format)
			router.Use(func(c *gin.Context) {
				c.Set("user_id", strconv.Itoa(int(user.ID)))
				c.Set("email", user.Email)
				c.Set("role", user.Role)
				c.Next()
			})

			router.PUT("/services/:id", handler.UpdateService)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response models.Service
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				if req := tt.requestBody.(UpdateServiceRequest); req.Description != nil {
					assert.Equal(t, *req.Description, response.Description)
				}
			}
		})
	}
}

func TestServiceHandler_DeleteService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, handler := setupServiceHandlerTest(t)

	user := createTestUser(t, db, "Developer")

	// Create test service
	testService := &models.Service{
		Name:        "test-service",
		Description: "Test service",
		UserID:      user.ID,
		Active:      true,
	}
	err := db.Create(testService).Error
	assert.NoError(t, err)

	tests := []struct {
		name           string
		serviceID      string
		expectedStatus int
	}{
		{
			name:           "delete existing service",
			serviceID:      fmt.Sprintf("%d", testService.ID),
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "delete non-existing service",
			serviceID:      "999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid service ID",
			serviceID:      "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("DELETE", "/services/"+tt.serviceID, nil)
			assert.NoError(t, err)

			w := httptest.NewRecorder()
			router := gin.New()

			// Mock middleware to set user context (matching JWT middleware format)
			router.Use(func(c *gin.Context) {
				c.Set("user_id", strconv.Itoa(int(user.ID)))
				c.Set("email", user.Email)
				c.Set("role", user.Role)
				c.Next()
			})

			router.DELETE("/services/:id", handler.DeleteService)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
