package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProject_TableName(t *testing.T) {
	project := Project{}
	assert.Equal(t, "projects", project.TableName())
}

func TestProject_GetModuleCount(t *testing.T) {
	project := Project{
		ProjectModules: []ProjectModule{
			{ModuleID: 1, Enabled: true},
			{ModuleID: 2, Enabled: true},
			{ModuleID: 3, Enabled: false}, // disabled
		},
	}

	count := project.GetModuleCount()
	assert.Equal(t, 2, count)
}

func TestProject_GetEnabledModules(t *testing.T) {
	project := Project{
		ProjectModules: []ProjectModule{
			{ModuleID: 1, Enabled: true},
			{ModuleID: 2, Enabled: false},
			{ModuleID: 3, Enabled: true},
		},
	}

	enabled := project.GetEnabledModules()
	assert.Len(t, enabled, 2)
	assert.Equal(t, uint(1), enabled[0].ModuleID)
	assert.Equal(t, uint(3), enabled[1].ModuleID)
}

func TestProject_HasModule(t *testing.T) {
	project := Project{
		ProjectModules: []ProjectModule{
			{ModuleID: 1, Enabled: true},
			{ModuleID: 2, Enabled: false},
			{ModuleID: 3, Enabled: true},
		},
	}

	tests := []struct {
		name     string
		moduleID uint
		expected bool
	}{
		{
			name:     "has enabled module",
			moduleID: 1,
			expected: true,
		},
		{
			name:     "has disabled module",
			moduleID: 2,
			expected: false,
		},
		{
			name:     "does not have module",
			moduleID: 4,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := project.HasModule(tt.moduleID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProject_CanBuild(t *testing.T) {
	tests := []struct {
		name     string
		project  Project
		expected bool
	}{
		{
			name: "can build active project with modules",
			project: Project{
				Active: true,
				ProjectModules: []ProjectModule{
					{ModuleID: 1, Enabled: true},
				},
			},
			expected: true,
		},
		{
			name: "cannot build inactive project",
			project: Project{
				Active: false,
				ProjectModules: []ProjectModule{
					{ModuleID: 1, Enabled: true},
				},
			},
			expected: false,
		},
		{
			name: "cannot build project with no modules",
			project: Project{
				Active:         true,
				ProjectModules: []ProjectModule{},
			},
			expected: false,
		},
		{
			name: "cannot build project with only disabled modules",
			project: Project{
				Active: true,
				ProjectModules: []ProjectModule{
					{ModuleID: 1, Enabled: false},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.project.CanBuild()
			assert.Equal(t, tt.expected, result)
		})
	}
}