package services

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/burndler/burndler/internal/models"
	"github.com/burndler/burndler/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ContainerServiceMockStorage extends MockStorage with file storage tracking
type ContainerServiceMockStorage struct {
	MockStorage
	files map[string][]byte
}

func setupContainerServiceTest(t *testing.T) (*gorm.DB, *ContainerService, *ContainerServiceMockStorage) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.Container{},
		&models.ContainerVersion{},
		&models.ContainerResource{},
	)
	require.NoError(t, err)

	mockStorage := &ContainerServiceMockStorage{
		files: make(map[string][]byte),
	}

	linter := NewLinter()
	service := NewContainerService(db, mockStorage, linter)

	return db, service, mockStorage
}

// Override methods to track files
func (m *ContainerServiceMockStorage) Delete(ctx context.Context, key string) error {
	delete(m.files, key)
	return m.MockStorage.Delete(ctx, key)
}

func (m *ContainerServiceMockStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, exists := m.files[key]
	return exists, nil
}

func (m *ContainerServiceMockStorage) UploadMultipart(ctx context.Context, key string, file *multipart.FileHeader) (storage.UploadResult, error) {
	// Read file content
	src, err := file.Open()
	if err != nil {
		return storage.UploadResult{}, err
	}
	defer func() {
		_ = src.Close()
	}()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(src); err != nil {
		return storage.UploadResult{}, err
	}

	m.files[key] = buf.Bytes()

	return storage.UploadResult{
		Key:         key,
		URL:         "http://mock/" + key,
		Size:        file.Size,
		ContentType: file.Header.Get("Content-Type"),
		Checksum:    "mock-checksum",
	}, nil
}

func (m *ContainerServiceMockStorage) DownloadBatch(ctx context.Context, keys []string) (map[string][]byte, error) {
	results := make(map[string][]byte)
	for _, key := range keys {
		if data, ok := m.files[key]; ok {
			results[key] = data
		}
	}
	return results, nil
}

// Helper to create a test container
func createTestContainer(t *testing.T, service *ContainerService) *models.Container {
	container, err := service.CreateContainer(CreateContainerRequest{
		Name:        "test-container",
		Description: "Test container",
		Author:      "test-author",
		Repository:  "http://test.com/repo",
	})
	require.NoError(t, err)
	return container
}

// Helper to create a test version
func createTestVersion(t *testing.T, service *ContainerService, containerID uint) *models.ContainerVersion {
	version, err := service.CreateVersion(containerID, CreateVersionRequest{
		Version: "v1.0.0",
		Compose: `version: "3.8"
services:
  web:
    image: nginx:latest
`,
		Variables:     map[string]interface{}{"var1": "value1"},
		ResourcePaths: []string{},
		Dependencies:  map[string]string{},
	})
	require.NoError(t, err)
	return version
}

// Helper to create a multipart file header
func createMockFileHeader(filename string, content []byte) *multipart.FileHeader {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filename)
	_, _ = part.Write(content)
	_ = writer.Close()

	req, _ := http.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	_ = req.ParseMultipartForm(32 << 20)

	file, _, _ := req.FormFile("file")
	if file != nil {
		defer func() {
			_ = file.Close()
		}()
	}

	return req.MultipartForm.File["file"][0]
}

func TestContainerService_UploadResource(t *testing.T) {
	db, service, mockStorage := setupContainerServiceTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	container := createTestContainer(t, service)
	version := createTestVersion(t, service, container.ID)

	t.Run("successful upload", func(t *testing.T) {
		fileContent := []byte("test file content")
		fileHeader := createMockFileHeader("test.txt", fileContent)

		resource, err := service.UploadResource(context.Background(), container.ID, version.Version, "test.txt", fileHeader)

		assert.NoError(t, err)
		assert.NotNil(t, resource)
		assert.Equal(t, version.ID, resource.ContainerVersionID)
		assert.Equal(t, "test.txt", resource.Path)
		assert.NotEmpty(t, resource.StorageKey)
		assert.Equal(t, fileHeader.Size, resource.FileSize)
		assert.Equal(t, "mock-checksum", resource.Checksum)
		assert.False(t, resource.IsDirectory)

		// Verify file was stored
		_, exists := mockStorage.files[resource.StorageKey]
		assert.True(t, exists)
	})

	t.Run("upload to non-existent container", func(t *testing.T) {
		fileContent := []byte("test content")
		fileHeader := createMockFileHeader("test.txt", fileContent)

		_, err := service.UploadResource(context.Background(), 9999, version.Version, "test.txt", fileHeader)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "container not found")
	})

	t.Run("upload to non-existent version", func(t *testing.T) {
		fileContent := []byte("test content")
		fileHeader := createMockFileHeader("test.txt", fileContent)

		_, err := service.UploadResource(context.Background(), container.ID, "v99.99.99", "test.txt", fileHeader)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("upload to published version", func(t *testing.T) {
		// Create and publish a version
		publishedVersion, err := service.CreateVersion(container.ID, CreateVersionRequest{
			Version: "v2.0.0",
			Compose: `version: "3.8"
services:
  web:
    image: nginx:latest
`,
		})
		require.NoError(t, err)

		_, err = service.PublishVersion(container.ID, publishedVersion.Version)
		require.NoError(t, err)

		// Try to upload resource
		fileContent := []byte("test content")
		fileHeader := createMockFileHeader("test.txt", fileContent)

		_, err = service.UploadResource(context.Background(), container.ID, publishedVersion.Version, "test.txt", fileHeader)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot upload resources to published version")
	})
}

func TestContainerService_ListResources(t *testing.T) {
	db, service, _ := setupContainerServiceTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	container := createTestContainer(t, service)
	version := createTestVersion(t, service, container.ID)

	t.Run("list empty resources", func(t *testing.T) {
		resources, err := service.ListResources(container.ID, version.Version)

		assert.NoError(t, err)
		assert.Empty(t, resources)
	})

	t.Run("list resources after upload", func(t *testing.T) {
		// Upload multiple resources
		fileContent1 := []byte("content1")
		fileHeader1 := createMockFileHeader("file1.txt", fileContent1)
		_, err := service.UploadResource(context.Background(), container.ID, version.Version, "file1.txt", fileHeader1)
		require.NoError(t, err)

		fileContent2 := []byte("content2")
		fileHeader2 := createMockFileHeader("file2.json", fileContent2)
		_, err = service.UploadResource(context.Background(), container.ID, version.Version, "dir/file2.json", fileHeader2)
		require.NoError(t, err)

		// List resources
		resources, err := service.ListResources(container.ID, version.Version)

		assert.NoError(t, err)
		assert.Len(t, resources, 2)
		// Verify ordered by path
		assert.Equal(t, "dir/file2.json", resources[0].Path)
		assert.Equal(t, "file1.txt", resources[1].Path)
	})

	t.Run("list resources for non-existent version", func(t *testing.T) {
		_, err := service.ListResources(container.ID, "v99.99.99")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestContainerService_DeleteResource(t *testing.T) {
	db, service, mockStorage := setupContainerServiceTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	container := createTestContainer(t, service)
	version := createTestVersion(t, service, container.ID)

	t.Run("successful delete", func(t *testing.T) {
		// Upload a resource
		fileContent := []byte("test content")
		fileHeader := createMockFileHeader("test.txt", fileContent)
		resource, err := service.UploadResource(context.Background(), container.ID, version.Version, "test.txt", fileHeader)
		require.NoError(t, err)

		// Verify file exists in storage
		_, exists := mockStorage.files[resource.StorageKey]
		require.True(t, exists)

		// Delete resource
		err = service.DeleteResource(context.Background(), container.ID, version.Version, resource.ID)

		assert.NoError(t, err)

		// Verify file removed from storage
		_, exists = mockStorage.files[resource.StorageKey]
		assert.False(t, exists)

		// Verify database record removed
		var dbResource models.ContainerResource
		err = db.First(&dbResource, resource.ID).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	t.Run("delete from published version", func(t *testing.T) {
		// Create and publish a version with resource
		publishedVersion, err := service.CreateVersion(container.ID, CreateVersionRequest{
			Version: "v3.0.0",
			Compose: `version: "3.8"
services:
  web:
    image: nginx:latest
`,
		})
		require.NoError(t, err)

		fileContent := []byte("test content")
		fileHeader := createMockFileHeader("test.txt", fileContent)
		resource, err := service.UploadResource(context.Background(), container.ID, publishedVersion.Version, "test.txt", fileHeader)
		require.NoError(t, err)

		_, err = service.PublishVersion(container.ID, publishedVersion.Version)
		require.NoError(t, err)

		// Try to delete resource
		err = service.DeleteResource(context.Background(), container.ID, publishedVersion.Version, resource.ID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete resources from published version")
	})

	t.Run("delete non-existent resource", func(t *testing.T) {
		err := service.DeleteResource(context.Background(), container.ID, version.Version, 9999)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource not found")
	})

	t.Run("delete with non-existent version", func(t *testing.T) {
		err := service.DeleteResource(context.Background(), container.ID, "v99.99.99", 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestContainerService_ResourceIntegration(t *testing.T) {
	db, service, _ := setupContainerServiceTest(t)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	container := createTestContainer(t, service)
	version := createTestVersion(t, service, container.ID)

	// Upload multiple resources
	files := []struct {
		path    string
		content string
	}{
		{"config.yaml", "key: value"},
		{"scripts/setup.sh", "#!/bin/bash\necho 'setup'"},
		{"data/sample.json", `{"test": true}`},
	}

	for _, file := range files {
		fileHeader := createMockFileHeader(file.path, []byte(file.content))
		_, err := service.UploadResource(context.Background(), container.ID, version.Version, file.path, fileHeader)
		require.NoError(t, err)
	}

	// List all resources
	resources, err := service.ListResources(container.ID, version.Version)
	require.NoError(t, err)
	assert.Len(t, resources, 3)

	// Verify paths are ordered
	assert.Equal(t, "config.yaml", resources[0].Path)
	assert.Equal(t, "data/sample.json", resources[1].Path)
	assert.Equal(t, "scripts/setup.sh", resources[2].Path)

	// Delete middle resource
	err = service.DeleteResource(context.Background(), container.ID, version.Version, resources[1].ID)
	require.NoError(t, err)

	// List again
	resources, err = service.ListResources(container.ID, version.Version)
	require.NoError(t, err)
	assert.Len(t, resources, 2)
	assert.Equal(t, "config.yaml", resources[0].Path)
	assert.Equal(t, "scripts/setup.sh", resources[1].Path)
}
