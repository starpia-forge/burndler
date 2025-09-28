package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Build represents a package build job
type Build struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string         `gorm:"not null" json:"name"`
	ServiceID *uint          `gorm:"index" json:"service_id"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	Status       string         `gorm:"not null;default:'queued'" json:"status"` // queued, building, completed, failed
	Progress     int            `gorm:"default:0" json:"progress"`               // 0-100
	DownloadURL  string         `json:"download_url,omitempty"`
	Error        string         `json:"error,omitempty"`
	ComposeYAML  string         `gorm:"type:text" json:"compose_yaml,omitempty"`
	ManifestJSON string         `gorm:"type:text" json:"manifest_json,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User    User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Service *Service `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
}

// TableName specifies the table name for Build model
func (Build) TableName() string {
	return "builds"
}

// BeforeCreate hook to generate UUID
func (b *Build) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// IsComplete checks if build is completed
func (b *Build) IsComplete() bool {
	return b.Status == "completed"
}

// IsFailed checks if build has failed
func (b *Build) IsFailed() bool {
	return b.Status == "failed"
}

// IsInProgress checks if build is in progress
func (b *Build) IsInProgress() bool {
	return b.Status == "building"
}

// IsServiceBuild checks if this build is based on a service
func (b *Build) IsServiceBuild() bool {
	return b.ServiceID != nil
}

// IsDirectBuild checks if this build is a direct compose build
func (b *Build) IsDirectBuild() bool {
	return b.ServiceID == nil
}

// GetBuildType returns the type of build
func (b *Build) GetBuildType() string {
	if b.IsServiceBuild() {
		return "service"
	}
	return "direct"
}

// GetSourceID returns the source ID (Service ID)
func (b *Build) GetSourceID() *uint {
	return b.ServiceID
}

// GetSourceName returns the source name (Service name)
func (b *Build) GetSourceName() string {
	if b.Service != nil {
		return b.Service.Name
	}
	return ""
}
