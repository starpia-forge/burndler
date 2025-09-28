package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Project represents a collection of containers for deployment
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
	ProjectContainers []ProjectContainer `gorm:"foreignKey:ProjectID" json:"project_containers,omitempty"`
	Builds         []Build         `gorm:"foreignKey:ProjectID" json:"builds,omitempty"`
}

// TableName specifies the table name for Project model
func (Project) TableName() string {
	return "projects"
}

// GetContainerCount returns the number of containers in this project
func (p *Project) GetContainerCount() int {
	count := 0
	for _, pc := range p.ProjectContainers {
		if pc.Enabled {
			count++
		}
	}
	return count
}

// GetEnabledContainers returns all enabled project containers ordered by position
func (p *Project) GetEnabledContainers() []ProjectContainer {
	var enabled []ProjectContainer
	for _, pc := range p.ProjectContainers {
		if pc.Enabled {
			enabled = append(enabled, pc)
		}
	}
	return enabled
}

// HasContainer checks if project contains a specific container
func (p *Project) HasContainer(containerID uint) bool {
	for _, pc := range p.ProjectContainers {
		if pc.ContainerID == containerID && pc.Enabled {
			return true
		}
	}
	return false
}

// CanBuild checks if project is ready for building
func (p *Project) CanBuild() bool {
	return p.Active && p.GetContainerCount() > 0
}