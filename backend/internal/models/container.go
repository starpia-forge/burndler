package models

import (
	"time"

	"gorm.io/gorm"
)

// Container represents a reusable deployment unit
type Container struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"uniqueIndex;not null" json:"name"`
	Description string         `json:"description"`
	Author      string         `json:"author"`
	Repository  string         `json:"repository"`
	Active      bool           `gorm:"default:true" json:"active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Versions []ContainerVersion `gorm:"foreignKey:ContainerID" json:"versions,omitempty"`
}

// TableName specifies the table name for Container model
func (Container) TableName() string {
	return "containers"
}

// GetLatestVersion returns the latest published version
func (c *Container) GetLatestVersion() *ContainerVersion {
	for i := len(c.Versions) - 1; i >= 0; i-- {
		if c.Versions[i].Published {
			return &c.Versions[i]
		}
	}
	return nil
}

// HasPublishedVersions checks if container has any published versions
func (c *Container) HasPublishedVersions() bool {
	for _, version := range c.Versions {
		if version.Published {
			return true
		}
	}
	return false
}