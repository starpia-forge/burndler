package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_TableName(t *testing.T) {
	service := Service{}
	assert.Equal(t, "services", service.TableName())
}

func TestService_GetContainerCount(t *testing.T) {
	tests := []struct {
		name       string
		service    *Service
		expected   int
	}{
		{
			name: "empty service",
			service: &Service{
				ServiceContainers: []ServiceContainer{},
			},
			expected: 0,
		},
		{
			name: "all enabled containers",
			service: &Service{
				ServiceContainers: []ServiceContainer{
					{Enabled: true},
					{Enabled: true},
					{Enabled: true},
				},
			},
			expected: 3,
		},
		{
			name: "mixed enabled/disabled containers",
			service: &Service{
				ServiceContainers: []ServiceContainer{
					{Enabled: true},
					{Enabled: false},
					{Enabled: true},
				},
			},
			expected: 2,
		},
		{
			name: "all disabled containers",
			service: &Service{
				ServiceContainers: []ServiceContainer{
					{Enabled: false},
					{Enabled: false},
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := tt.service.GetContainerCount()
			assert.Equal(t, tt.expected, count)
		})
	}
}

func TestService_GetEnabledContainers(t *testing.T) {
	service := &Service{
		ServiceContainers: []ServiceContainer{
			{ID: 1, Enabled: true, Order: 1},
			{ID: 2, Enabled: false, Order: 2},
			{ID: 3, Enabled: true, Order: 3},
		},
	}

	enabled := service.GetEnabledContainers()
	assert.Len(t, enabled, 2)
	assert.Equal(t, uint(1), enabled[0].ID)
	assert.Equal(t, uint(3), enabled[1].ID)
}

func TestService_HasContainer(t *testing.T) {
	service := &Service{
		ServiceContainers: []ServiceContainer{
			{ContainerID: 1, Enabled: true},
			{ContainerID: 2, Enabled: false},
			{ContainerID: 3, Enabled: true},
		},
	}

	tests := []struct {
		name        string
		containerID uint
		expected    bool
	}{
		{
			name:        "has enabled container",
			containerID: 1,
			expected:    true,
		},
		{
			name:        "has disabled container",
			containerID: 2,
			expected:    false,
		},
		{
			name:        "does not have container",
			containerID: 99,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.HasContainer(tt.containerID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestService_CanBuild(t *testing.T) {
	tests := []struct {
		name     string
		service  *Service
		expected bool
	}{
		{
			name: "active service with containers",
			service: &Service{
				Active: true,
				ServiceContainers: []ServiceContainer{
					{Enabled: true},
				},
			},
			expected: true,
		},
		{
			name: "inactive service with containers",
			service: &Service{
				Active: false,
				ServiceContainers: []ServiceContainer{
					{Enabled: true},
				},
			},
			expected: false,
		},
		{
			name: "active service without containers",
			service: &Service{
				Active:            true,
				ServiceContainers: []ServiceContainer{},
			},
			expected: false,
		},
		{
			name: "active service with only disabled containers",
			service: &Service{
				Active: true,
				ServiceContainers: []ServiceContainer{
					{Enabled: false},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.service.CanBuild()
			assert.Equal(t, tt.expected, result)
		})
	}
}