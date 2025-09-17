package models

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
)

// ProjectModule represents the many-to-many relationship between projects and modules
type ProjectModule struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	ProjectID       uint           `gorm:"not null;index" json:"project_id"`
	ModuleID        uint           `gorm:"not null;index" json:"module_id"`
	ModuleVersionID uint           `gorm:"not null;index" json:"module_version_id"`
	Order           int            `gorm:"default:0" json:"order"`
	Enabled         bool           `gorm:"default:true" json:"enabled"`
	OverrideVars    datatypes.JSON `gorm:"type:text" json:"override_vars"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`

	// Relationships
	Project       Project       `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Module        Module        `gorm:"foreignKey:ModuleID" json:"module,omitempty"`
	ModuleVersion ModuleVersion `gorm:"foreignKey:ModuleVersionID" json:"module_version,omitempty"`
}

// TableName specifies the table name for ProjectModule model
func (ProjectModule) TableName() string {
	return "project_modules"
}

// GetDisplayName returns a display name for the project module
func (pm *ProjectModule) GetDisplayName() string {
	if pm.Module.Name != "" && pm.ModuleVersion.Version != "" {
		return pm.Module.Name + ":" + pm.ModuleVersion.Version
	}
	return "Unknown Module"
}

// IsConfigured checks if the project module has override variables
func (pm *ProjectModule) IsConfigured() bool {
	if pm.OverrideVars == nil {
		return false
	}

	var overrideVars map[string]interface{}
	if err := json.Unmarshal(pm.OverrideVars, &overrideVars); err != nil {
		return false
	}

	return len(overrideVars) > 0
}

// GetVariables returns the effective variables for this module
// combining module defaults with project overrides
func (pm *ProjectModule) GetEffectiveVariables() map[string]interface{} {
	variables := make(map[string]interface{})

	// Start with module version variables
	if pm.ModuleVersion.Variables != nil {
		var moduleVars map[string]interface{}
		if err := json.Unmarshal(pm.ModuleVersion.Variables, &moduleVars); err != nil {
			// Log error but continue with empty moduleVars
			moduleVars = make(map[string]interface{})
		}
		for k, v := range moduleVars {
			variables[k] = v
		}
	}

	// Override with project-specific variables
	if pm.OverrideVars != nil {
		var overrideVars map[string]interface{}
		if err := json.Unmarshal(pm.OverrideVars, &overrideVars); err != nil {
			// Log error but continue with empty overrideVars
			overrideVars = make(map[string]interface{})
		}
		for k, v := range overrideVars {
			variables[k] = v
		}
	}

	return variables
}