package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestModuleVersion_TableName(t *testing.T) {
	mv := ModuleVersion{}
	assert.Equal(t, "module_versions", mv.TableName())
}

func TestModuleVersion_Publish(t *testing.T) {
	mv := ModuleVersion{
		Version:   "v1.0.0",
		Published: false,
	}

	// Record time before publishing
	before := time.Now()

	mv.Publish()

	assert.True(t, mv.Published)
	assert.NotNil(t, mv.PublishedAt)
	assert.True(t, mv.PublishedAt.After(before))
}

func TestModuleVersion_CanModify(t *testing.T) {
	tests := []struct {
		name      string
		version   ModuleVersion
		canModify bool
	}{
		{
			name: "can modify unpublished version",
			version: ModuleVersion{
				Published: false,
			},
			canModify: true,
		},
		{
			name: "cannot modify published version",
			version: ModuleVersion{
				Published: true,
			},
			canModify: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version.CanModify()
			assert.Equal(t, tt.canModify, result)
		})
	}
}

func TestModuleVersion_GetFullName(t *testing.T) {
	mv := ModuleVersion{
		Version: "v1.0.0",
		Module: Module{
			Name: "webapp",
		},
	}

	result := mv.GetFullName()
	assert.Equal(t, "webapp:v1.0.0", result)
}