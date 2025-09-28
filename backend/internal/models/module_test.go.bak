package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModule_TableName(t *testing.T) {
	module := Module{}
	assert.Equal(t, "modules", module.TableName())
}

func TestModule_GetLatestVersion(t *testing.T) {
	tests := []struct {
		name     string
		module   Module
		expected *ModuleVersion
	}{
		{
			name: "returns latest published version",
			module: Module{
				Versions: []ModuleVersion{
					{Version: "v1.0.0", Published: true},
					{Version: "v1.1.0", Published: true},
					{Version: "v1.2.0", Published: false},
				},
			},
			expected: &ModuleVersion{Version: "v1.1.0", Published: true},
		},
		{
			name: "returns nil when no published versions",
			module: Module{
				Versions: []ModuleVersion{
					{Version: "v1.0.0", Published: false},
					{Version: "v1.1.0", Published: false},
				},
			},
			expected: nil,
		},
		{
			name:     "returns nil when no versions",
			module:   Module{Versions: []ModuleVersion{}},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.module.GetLatestVersion()
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.Version, result.Version)
				assert.Equal(t, tt.expected.Published, result.Published)
			}
		})
	}
}

func TestModule_HasPublishedVersions(t *testing.T) {
	tests := []struct {
		name     string
		module   Module
		expected bool
	}{
		{
			name: "returns true when has published versions",
			module: Module{
				Versions: []ModuleVersion{
					{Version: "v1.0.0", Published: false},
					{Version: "v1.1.0", Published: true},
				},
			},
			expected: true,
		},
		{
			name: "returns false when no published versions",
			module: Module{
				Versions: []ModuleVersion{
					{Version: "v1.0.0", Published: false},
					{Version: "v1.1.0", Published: false},
				},
			},
			expected: false,
		},
		{
			name:     "returns false when no versions",
			module:   Module{Versions: []ModuleVersion{}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.module.HasPublishedVersions()
			assert.Equal(t, tt.expected, result)
		})
	}
}