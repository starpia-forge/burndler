package models

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
)

// ServiceContainer represents the many-to-many relationship between services and containers
type ServiceContainer struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	ServiceID          uint           `gorm:"not null;index" json:"service_id"`
	ContainerID        uint           `gorm:"not null;index" json:"container_id"`
	ContainerVersionID uint           `gorm:"not null;index" json:"container_version_id"`
	Order              int            `gorm:"default:0" json:"order"`
	Enabled            bool           `gorm:"default:true" json:"enabled"`
	OverrideVars       datatypes.JSON `gorm:"type:text" json:"override_vars"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`

	// Relationships
	Service          Service          `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Container        Container        `gorm:"foreignKey:ContainerID" json:"container,omitempty"`
	ContainerVersion ContainerVersion `gorm:"foreignKey:ContainerVersionID" json:"container_version,omitempty"`
}

// TableName specifies the table name for ServiceContainer model
func (ServiceContainer) TableName() string {
	return "service_containers"
}

// GetDisplayName returns a display name for the service container
func (sc *ServiceContainer) GetDisplayName() string {
	if sc.Container.Name != "" && sc.ContainerVersion.Version != "" {
		return sc.Container.Name + ":" + sc.ContainerVersion.Version
	}
	return "Unknown Container"
}

// IsConfigured checks if the service container has override variables
func (sc *ServiceContainer) IsConfigured() bool {
	if sc.OverrideVars == nil {
		return false
	}

	var overrideVars map[string]interface{}
	if err := json.Unmarshal(sc.OverrideVars, &overrideVars); err != nil {
		return false
	}

	return len(overrideVars) > 0
}

// GetEffectiveVariables returns the effective variables for this container
// combining container defaults with service overrides
func (sc *ServiceContainer) GetEffectiveVariables() map[string]interface{} {
	variables := make(map[string]interface{})

	// Start with container version variables
	if sc.ContainerVersion.Variables != nil {
		var containerVars map[string]interface{}
		if err := json.Unmarshal(sc.ContainerVersion.Variables, &containerVars); err != nil {
			// Log error but continue with empty containerVars
			containerVars = make(map[string]interface{})
		}
		for k, v := range containerVars {
			variables[k] = v
		}
	}

	// Override with service-specific variables
	if sc.OverrideVars != nil {
		var overrideVars map[string]interface{}
		if err := json.Unmarshal(sc.OverrideVars, &overrideVars); err != nil {
			// Log error but continue with empty overrideVars
			overrideVars = make(map[string]interface{})
		}
		for k, v := range overrideVars {
			variables[k] = v
		}
	}

	return variables
}