package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Service represents a collection of containers for deployment
type Service struct {
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
	User              User               `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ServiceContainers []ServiceContainer `gorm:"foreignKey:ServiceID" json:"service_containers,omitempty"`
	Builds            []Build            `gorm:"foreignKey:ServiceID" json:"builds,omitempty"`
}

// TableName specifies the table name for Service model
func (Service) TableName() string {
	return "services"
}

// GetContainerCount returns the number of containers in this service
func (s *Service) GetContainerCount() int {
	count := 0
	for _, sc := range s.ServiceContainers {
		if sc.Enabled {
			count++
		}
	}
	return count
}

// GetEnabledContainers returns all enabled service containers ordered by position
func (s *Service) GetEnabledContainers() []ServiceContainer {
	var enabled []ServiceContainer
	for _, sc := range s.ServiceContainers {
		if sc.Enabled {
			enabled = append(enabled, sc)
		}
	}
	return enabled
}

// HasContainer checks if service contains a specific container
func (s *Service) HasContainer(containerID uint) bool {
	for _, sc := range s.ServiceContainers {
		if sc.ContainerID == containerID && sc.Enabled {
			return true
		}
	}
	return false
}

// CanBuild checks if service is ready for building
func (s *Service) CanBuild() bool {
	return s.Active && s.GetContainerCount() > 0
}