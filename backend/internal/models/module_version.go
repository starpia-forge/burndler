package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ModuleVersion represents a versioned module release
type ModuleVersion struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	ModuleID        uint           `gorm:"not null;index" json:"module_id"`
	Version         string         `gorm:"not null" json:"version"`
	ComposeContent  string         `gorm:"type:text;not null" json:"compose_content"`
	Variables       datatypes.JSON `gorm:"type:text" json:"variables"`
	ResourcePaths   datatypes.JSON `gorm:"type:text" json:"resource_paths"`
	Dependencies    datatypes.JSON `gorm:"type:text" json:"dependencies"`
	Published       bool           `gorm:"default:false" json:"published"`
	PublishedAt     *time.Time     `json:"published_at"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Module Module `gorm:"foreignKey:ModuleID" json:"module,omitempty"`
}

// TableName specifies the table name for ModuleVersion model
func (ModuleVersion) TableName() string {
	return "module_versions"
}

// BeforeUpdate ensures published versions cannot be modified
func (mv *ModuleVersion) BeforeUpdate(tx *gorm.DB) error {
	// Check if this record was already published before this update
	var original ModuleVersion
	if err := tx.Where("id = ?", mv.ID).First(&original).Error; err != nil {
		return nil // If we can't find the original, allow the update
	}

	// If the original was already published, prevent modification
	if original.Published {
		return gorm.ErrInvalidTransaction
	}

	return nil
}

// Publish marks the version as published
func (mv *ModuleVersion) Publish() {
	now := time.Now()
	mv.Published = true
	mv.PublishedAt = &now
}

// CanModify checks if version can be modified
func (mv *ModuleVersion) CanModify() bool {
	return !mv.Published
}

// GetFullName returns module name with version
func (mv *ModuleVersion) GetFullName() string {
	return mv.Module.Name + ":" + mv.Version
}