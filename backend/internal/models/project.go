package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Project represents a collection of modules for deployment
type Project struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"not null" json:"name"`
	Description     string         `json:"description"`
	UserID          uint           `gorm:"not null" json:"user_id"`
	Variables       datatypes.JSON `gorm:"type:text" json:"variables"`
	EnvironmentVars datatypes.JSON `gorm:"type:text" json:"environment_vars"`
	Active          bool           `gorm:"default:true" json:"active"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User           User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ProjectModules []ProjectModule `gorm:"foreignKey:ProjectID" json:"project_modules,omitempty"`
	Builds         []Build         `gorm:"foreignKey:ProjectID" json:"builds,omitempty"`
}

// TableName specifies the table name for Project model
func (Project) TableName() string {
	return "projects"
}

// GetModuleCount returns the number of modules in this project
func (p *Project) GetModuleCount() int {
	count := 0
	for _, pm := range p.ProjectModules {
		if pm.Enabled {
			count++
		}
	}
	return count
}

// GetEnabledModules returns all enabled project modules ordered by position
func (p *Project) GetEnabledModules() []ProjectModule {
	var enabled []ProjectModule
	for _, pm := range p.ProjectModules {
		if pm.Enabled {
			enabled = append(enabled, pm)
		}
	}
	return enabled
}

// HasModule checks if project contains a specific module
func (p *Project) HasModule(moduleID uint) bool {
	for _, pm := range p.ProjectModules {
		if pm.ModuleID == moduleID && pm.Enabled {
			return true
		}
	}
	return false
}

// CanBuild checks if project is ready for building
func (p *Project) CanBuild() bool {
	return p.Active && p.GetModuleCount() > 0
}