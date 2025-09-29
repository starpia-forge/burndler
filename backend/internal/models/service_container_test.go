package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func TestServiceContainer_TableName(t *testing.T) {
	sc := ServiceContainer{}
	assert.Equal(t, "service_containers", sc.TableName())
}

func TestServiceContainer_GetDisplayName(t *testing.T) {
	tests := []struct {
		name      string
		container *ServiceContainer
		expected  string
	}{
		{
			name: "with container name and version",
			container: &ServiceContainer{
				Container: Container{
					Name: "nginx",
				},
				ContainerVersion: ContainerVersion{
					Version: "v1.2.3",
				},
			},
			expected: "nginx:v1.2.3",
		},
		{
			name: "empty container name and version",
			container: &ServiceContainer{
				Container: Container{
					Name: "",
				},
				ContainerVersion: ContainerVersion{
					Version: "",
				},
			},
			expected: "Unknown Container",
		},
		{
			name: "only container name",
			container: &ServiceContainer{
				Container: Container{
					Name: "postgres",
				},
				ContainerVersion: ContainerVersion{
					Version: "",
				},
			},
			expected: "Unknown Container",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.container.GetDisplayName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestServiceContainer_IsConfigured(t *testing.T) {
	tests := []struct {
		name      string
		container *ServiceContainer
		expected  bool
	}{
		{
			name: "nil override vars",
			container: &ServiceContainer{
				OverrideVars: nil,
			},
			expected: false,
		},
		{
			name: "empty override vars",
			container: &ServiceContainer{
				OverrideVars: datatypes.JSON("{}"),
			},
			expected: false,
		},
		{
			name: "with override vars",
			container: &ServiceContainer{
				OverrideVars: datatypes.JSON(`{"key": "value"}`),
			},
			expected: true,
		},
		{
			name: "invalid json",
			container: &ServiceContainer{
				OverrideVars: datatypes.JSON("invalid-json"),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.container.IsConfigured()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestServiceContainer_GetEffectiveVariables(t *testing.T) {
	tests := []struct {
		name      string
		container *ServiceContainer
		expected  map[string]interface{}
	}{
		{
			name: "no variables",
			container: &ServiceContainer{
				ContainerVersion: ContainerVersion{
					Variables: nil,
				},
				OverrideVars: nil,
			},
			expected: map[string]interface{}{},
		},
		{
			name: "only container version variables",
			container: &ServiceContainer{
				ContainerVersion: ContainerVersion{
					Variables: datatypes.JSON(`{"port": 8080, "name": "app"}`),
				},
				OverrideVars: nil,
			},
			expected: map[string]interface{}{
				"port": float64(8080), // JSON unmarshal gives float64 for numbers
				"name": "app",
			},
		},
		{
			name: "only override variables",
			container: &ServiceContainer{
				ContainerVersion: ContainerVersion{
					Variables: nil,
				},
				OverrideVars: datatypes.JSON(`{"env": "production"}`),
			},
			expected: map[string]interface{}{
				"env": "production",
			},
		},
		{
			name: "merged variables with override",
			container: &ServiceContainer{
				ContainerVersion: ContainerVersion{
					Variables: datatypes.JSON(`{"port": 8080, "name": "app", "env": "development"}`),
				},
				OverrideVars: datatypes.JSON(`{"env": "production", "replicas": 3}`),
			},
			expected: map[string]interface{}{
				"port":     float64(8080),
				"name":     "app",
				"env":      "production", // overridden
				"replicas": float64(3),   // added
			},
		},
		{
			name: "invalid container variables",
			container: &ServiceContainer{
				ContainerVersion: ContainerVersion{
					Variables: datatypes.JSON("invalid-json"),
				},
				OverrideVars: datatypes.JSON(`{"env": "production"}`),
			},
			expected: map[string]interface{}{
				"env": "production",
			},
		},
		{
			name: "invalid override variables",
			container: &ServiceContainer{
				ContainerVersion: ContainerVersion{
					Variables: datatypes.JSON(`{"port": 8080}`),
				},
				OverrideVars: datatypes.JSON("invalid-json"),
			},
			expected: map[string]interface{}{
				"port": float64(8080),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.container.GetEffectiveVariables()
			assert.Equal(t, tt.expected, result)
		})
	}
}