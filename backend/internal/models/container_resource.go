package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ContainerResource represents a resource file or directory for a container version
// Resources are stored in Storage (S3/LocalFS) and referenced by this model
type ContainerResource struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	ContainerVersionID uint      `gorm:"not null;index" json:"container_version_id"`
	Path               string    `gorm:"size:512;not null" json:"path"`        // Relative path within container (e.g., "config/app.yaml")
	StorageKey         string    `gorm:"size:512;not null" json:"storage_key"` // Storage key for retrieval
	FileSize           int64     `gorm:"not null" json:"file_size"`
	Checksum           string    `gorm:"size:64" json:"checksum"` // SHA256 checksum
	IsDirectory        bool      `gorm:"default:false" json:"is_directory"`
	ContentType        string    `gorm:"size:100" json:"content_type"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`

	// Relationships
	ContainerVersion ContainerVersion `gorm:"foreignKey:ContainerVersionID;constraint:OnDelete:CASCADE" json:"container_version,omitempty"`
}

// TableName specifies the table name for ContainerResource model
func (ContainerResource) TableName() string {
	return "container_resources"
}

// BeforeCreate validates required fields before creating a ContainerResource
func (cr *ContainerResource) BeforeCreate(tx *gorm.DB) error {
	if cr.ContainerVersionID == 0 {
		return fmt.Errorf("container_version_id is required")
	}
	if cr.Path == "" {
		return fmt.Errorf("path is required")
	}
	if cr.StorageKey == "" {
		return fmt.Errorf("storage_key is required")
	}
	return nil
}
