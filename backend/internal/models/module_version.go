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
	Variables       datatypes.JSON `gorm:"type:jsonb" json:"variables"`
	ResourcePaths   datatypes.JSON `gorm:"type:jsonb" json:"resource_paths"`
	Dependencies    datatypes.JSON `gorm:"type:jsonb" json:"dependencies"`
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
	if mv.Published {
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