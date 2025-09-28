package models

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
)

// ProjectContainer represents the many-to-many relationship between projects and containers
type ProjectContainer struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	ProjectID       uint           `gorm:"not null;index" json:"project_id"`
	ContainerID        uint           `gorm:"not null;index" json:"container_id"`
	ContainerVersionID uint           `gorm:"not null;index" json:"container_version_id"`
	Order           int            `gorm:"default:0" json:"order"`
	Enabled         bool           `gorm:"default:true" json:"enabled"`
	OverrideVars    datatypes.JSON `gorm:"type:text" json:"override_vars"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`

	// Relationships
	Project          Project          `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Container        Container        `gorm:"foreignKey:ContainerID" json:"container,omitempty"`
	ContainerVersion ContainerVersion `gorm:"foreignKey:ContainerVersionID" json:"container_version,omitempty"`
}

// TableName specifies the table name for ProjectContainer model
func (ProjectContainer) TableName() string {
	return "project_containers"
}

// GetDisplayName returns a display name for the project container
func (pc *ProjectContainer) GetDisplayName() string {
	if pc.Container.Name != "" && pc.ContainerVersion.Version != "" {
		return pc.Container.Name + ":" + pc.ContainerVersion.Version
	}
	return "Unknown Container"
}

// IsConfigured checks if the project container has override variables
func (pc *ProjectContainer) IsConfigured() bool {
	if pc.OverrideVars == nil {
		return false
	}

	var overrideVars map[string]interface{}
	if err := json.Unmarshal(pc.OverrideVars, &overrideVars); err != nil {
		return false
	}

	return len(overrideVars) > 0
}

// GetEffectiveVariables returns the effective variables for this container
// combining container defaults with project overrides
func (pc *ProjectContainer) GetEffectiveVariables() map[string]interface{} {
	variables := make(map[string]interface{})

	// Start with container version variables
	if pc.ContainerVersion.Variables != nil {
		var containerVars map[string]interface{}
		if err := json.Unmarshal(pc.ContainerVersion.Variables, &containerVars); err != nil {
			// Log error but continue with empty containerVars
			containerVars = make(map[string]interface{})
		}
		for k, v := range containerVars {
			variables[k] = v
		}
	}

	// Override with project-specific variables
	if pc.OverrideVars != nil {
		var overrideVars map[string]interface{}
		if err := json.Unmarshal(pc.OverrideVars, &overrideVars); err != nil {
			// Log error but continue with empty overrideVars
			overrideVars = make(map[string]interface{})
		}
		for k, v := range overrideVars {
			variables[k] = v
		}
	}

	return variables
}