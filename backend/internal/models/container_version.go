package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ContainerVersion represents a versioned container release
type ContainerVersion struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	ContainerID     uint           `gorm:"not null;index" json:"container_id"`
	Version         string         `gorm:"not null" json:"version"`
	ComposeContent  string         `gorm:"type:text;not null" json:"compose_content"`
	Variables       datatypes.JSON `gorm:"type:text" json:"variables"`
	ResourcePaths   datatypes.JSON `gorm:"type:text" json:"resource_paths"`
	Dependencies    datatypes.JSON `gorm:"type:text" json:"dependencies"`
	ConfigurationID *uint          `gorm:"index" json:"configuration_id,omitempty"`
	Published       bool           `gorm:"default:false" json:"published"`
	PublishedAt     *time.Time     `json:"published_at"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Container     Container                `gorm:"foreignKey:ContainerID" json:"container,omitempty"`
	Configuration *ContainerConfiguration `gorm:"foreignKey:ConfigurationID" json:"configuration,omitempty"`
}

// TableName specifies the table name for ContainerVersion model
func (ContainerVersion) TableName() string {
	return "container_versions"
}

// BeforeUpdate ensures published versions cannot be modified
func (cv *ContainerVersion) BeforeUpdate(tx *gorm.DB) error {
	// Check if this record was already published before this update
	var original ContainerVersion
	if err := tx.Where("id = ?", cv.ID).First(&original).Error; err != nil {
		return nil // If we can't find the original, allow the update
	}

	// If the original was already published, prevent modification
	if original.Published {
		return gorm.ErrInvalidTransaction
	}

	return nil
}

// Publish marks the version as published
func (cv *ContainerVersion) Publish() {
	now := time.Now()
	cv.Published = true
	cv.PublishedAt = &now
}

// CanModify checks if version can be modified
func (cv *ContainerVersion) CanModify() bool {
	return !cv.Published
}

// GetFullName returns container name with version
func (cv *ContainerVersion) GetFullName() string {
	return cv.Container.Name + ":" + cv.Version
}