package models

import (
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Test BeforeCreate hook generates UUID
func TestBuild_BeforeCreate(t *testing.T) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&Build{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Test case 1: Build without ID should get UUID
	build := &Build{
		Name:   "test-build",
		Status: "queued",
		UserID: 1,
	}

	err = db.Create(build).Error
	if err != nil {
		t.Fatalf("Failed to create build: %v", err)
	}

	if build.ID == uuid.Nil {
		t.Error("Expected UUID to be generated, got nil")
	}

	// Test case 2: Build with existing ID should keep it
	existingID := uuid.New()
	build2 := &Build{
		ID:     existingID,
		Name:   "test-build-2",
		Status: "queued",
		UserID: 1,
	}

	err = db.Create(build2).Error
	if err != nil {
		t.Fatalf("Failed to create build with existing ID: %v", err)
	}

	if build2.ID != existingID {
		t.Errorf("Expected ID to remain %v, got %v", existingID, build2.ID)
	}
}

// Test IsComplete method
func TestBuild_IsComplete(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"completed status", "completed", true},
		{"queued status", "queued", false},
		{"building status", "building", false},
		{"failed status", "failed", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			build := &Build{Status: tt.status}
			if got := build.IsComplete(); got != tt.expected {
				t.Errorf("IsComplete() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Test IsFailed method
func TestBuild_IsFailed(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"failed status", "failed", true},
		{"completed status", "completed", false},
		{"queued status", "queued", false},
		{"building status", "building", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			build := &Build{Status: tt.status}
			if got := build.IsFailed(); got != tt.expected {
				t.Errorf("IsFailed() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Test IsInProgress method
func TestBuild_IsInProgress(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"building status", "building", true},
		{"completed status", "completed", false},
		{"queued status", "queued", false},
		{"failed status", "failed", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			build := &Build{Status: tt.status}
			if got := build.IsInProgress(); got != tt.expected {
				t.Errorf("IsInProgress() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Test TableName method
func TestBuild_TableName(t *testing.T) {
	build := Build{}
	expected := "builds"
	if got := build.TableName(); got != expected {
		t.Errorf("TableName() = %v, want %v", got, expected)
	}
}

// Test IsProjectBuild method
func TestBuild_IsProjectBuild(t *testing.T) {
	tests := []struct {
		name      string
		projectID *uint
		expected  bool
	}{
		{"project build", ptrUint(1), true},
		{"direct build", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			build := &Build{ProjectID: tt.projectID}
			if got := build.IsProjectBuild(); got != tt.expected {
				t.Errorf("IsProjectBuild() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Test IsDirectBuild method
func TestBuild_IsDirectBuild(t *testing.T) {
	tests := []struct {
		name      string
		projectID *uint
		expected  bool
	}{
		{"direct build", nil, true},
		{"project build", ptrUint(1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			build := &Build{ProjectID: tt.projectID}
			if got := build.IsDirectBuild(); got != tt.expected {
				t.Errorf("IsDirectBuild() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Test GetBuildType method
func TestBuild_GetBuildType(t *testing.T) {
	tests := []struct {
		name      string
		projectID *uint
		expected  string
	}{
		{"project build type", ptrUint(1), "project"},
		{"direct build type", nil, "direct"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			build := &Build{ProjectID: tt.projectID}
			if got := build.GetBuildType(); got != tt.expected {
				t.Errorf("GetBuildType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Helper function to create a pointer to uint
func ptrUint(i uint) *uint {
	return &i
}
