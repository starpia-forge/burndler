package models

import (
	"time"

	"gorm.io/gorm"
)

// Module represents a reusable deployment unit
type Module struct {
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
	Versions []ModuleVersion `gorm:"foreignKey:ModuleID" json:"versions,omitempty"`
}

// TableName specifies the table name for Module model
func (Module) TableName() string {
	return "modules"
}

// GetLatestVersion returns the latest published version
func (m *Module) GetLatestVersion() *ModuleVersion {
	for i := len(m.Versions) - 1; i >= 0; i-- {
		if m.Versions[i].Published {
			return &m.Versions[i]
		}
	}
	return nil
}

// HasPublishedVersions checks if module has any published versions
func (m *Module) HasPublishedVersions() bool {
	for _, version := range m.Versions {
		if version.Published {
			return true
		}
	}
	return false
}