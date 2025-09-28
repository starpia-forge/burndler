package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func TestProjectModule_TableName(t *testing.T) {
	pm := ProjectModule{}
	assert.Equal(t, "project_modules", pm.TableName())
}

func TestProjectModule_GetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		pm       ProjectModule
		expected string
	}{
		{
			name: "returns module name with version",
			pm: ProjectModule{
				Module: Module{
					Name: "webapp",
				},
				ModuleVersion: ModuleVersion{
					Version: "v1.0.0",
				},
			},
			expected: "webapp:v1.0.0",
		},
		{
			name: "returns unknown when no module data",
			pm: ProjectModule{
				Module:        Module{},
				ModuleVersion: ModuleVersion{},
			},
			expected: "Unknown Module",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pm.GetDisplayName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProjectModule_IsConfigured(t *testing.T) {
	tests := []struct {
		name     string
		pm       ProjectModule
		expected bool
	}{
		{
			name: "returns true when has override variables",
			pm: ProjectModule{
				OverrideVars: datatypes.JSON(`{"key": "value"}`),
			},
			expected: true,
		},
		{
			name: "returns false when no override variables",
			pm: ProjectModule{
				OverrideVars: nil,
			},
			expected: false,
		},
		{
			name: "returns false when empty override variables",
			pm: ProjectModule{
				OverrideVars: datatypes.JSON(`{}`),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pm.IsConfigured()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProjectModule_GetEffectiveVariables(t *testing.T) {
	pm := ProjectModule{
		ModuleVersion: ModuleVersion{
			Variables: datatypes.JSON(`{"default_key": "default_value", "shared_key": "module_value"}`),
		},
		OverrideVars: datatypes.JSON(`{"override_key": "override_value", "shared_key": "project_value"}`),
	}

	result := pm.GetEffectiveVariables()

	// Should have all keys
	assert.Contains(t, result, "default_key")
	assert.Contains(t, result, "override_key")
	assert.Contains(t, result, "shared_key")

	// Default value should be preserved
	assert.Equal(t, "default_value", result["default_key"])

	// Override value should be preserved
	assert.Equal(t, "override_value", result["override_key"])

	// Project override should take precedence
	assert.Equal(t, "project_value", result["shared_key"])
}