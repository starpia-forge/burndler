package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestContainerConfiguration_IsCompatibleWithVersion tests semantic version compatibility
func TestContainerConfiguration_IsCompatibleWithVersion(t *testing.T) {
	tests := []struct {
		name           string
		minimumVersion string
		targetVersion  string
		expected       bool
	}{
		// Exact match cases
		{
			name:           "exact match v1.0.0",
			minimumVersion: "v1.0.0",
			targetVersion:  "v1.0.0",
			expected:       true,
		},
		{
			name:           "exact match v2.5.3",
			minimumVersion: "v2.5.3",
			targetVersion:  "v2.5.3",
			expected:       true,
		},

		// Higher version cases (should be compatible)
		{
			name:           "higher patch version",
			minimumVersion: "v1.0.0",
			targetVersion:  "v1.0.1",
			expected:       true,
		},
		{
			name:           "higher minor version",
			minimumVersion: "v1.0.0",
			targetVersion:  "v1.1.0",
			expected:       true,
		},
		{
			name:           "higher major version",
			minimumVersion: "v1.0.0",
			targetVersion:  "v2.0.0",
			expected:       true,
		},
		{
			name:           "much higher version",
			minimumVersion: "v1.0.0",
			targetVersion:  "v3.5.2",
			expected:       true,
		},

		// Lower version cases (should not be compatible)
		{
			name:           "lower patch version",
			minimumVersion: "v1.0.1",
			targetVersion:  "v1.0.0",
			expected:       false,
		},
		{
			name:           "lower minor version",
			minimumVersion: "v1.1.0",
			targetVersion:  "v1.0.0",
			expected:       false,
		},
		{
			name:           "lower major version",
			minimumVersion: "v2.0.0",
			targetVersion:  "v1.9.9",
			expected:       false,
		},
		{
			name:           "much lower version",
			minimumVersion: "v3.0.0",
			targetVersion:  "v1.0.0",
			expected:       false,
		},

		// Without 'v' prefix
		{
			name:           "versions without v prefix - compatible",
			minimumVersion: "1.0.0",
			targetVersion:  "1.1.0",
			expected:       true,
		},
		{
			name:           "versions without v prefix - not compatible",
			minimumVersion: "1.1.0",
			targetVersion:  "1.0.0",
			expected:       false,
		},

		// Mixed formats
		{
			name:           "minimum with v, target without",
			minimumVersion: "v1.0.0",
			targetVersion:  "1.1.0",
			expected:       true,
		},
		{
			name:           "minimum without v, target with",
			minimumVersion: "1.0.0",
			targetVersion:  "v1.1.0",
			expected:       true,
		},

		// Edge cases with different patch/minor combinations
		{
			name:           "v1.0.9 vs v1.1.0 (higher minor wins)",
			minimumVersion: "v1.0.9",
			targetVersion:  "v1.1.0",
			expected:       true,
		},
		{
			name:           "v1.9.9 vs v2.0.0 (higher major wins)",
			minimumVersion: "v1.9.9",
			targetVersion:  "v2.0.0",
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &ContainerConfiguration{
				MinimumVersion: tt.minimumVersion,
			}

			result := config.IsCompatibleWithVersion(tt.targetVersion)
			assert.Equal(t, tt.expected, result,
				"IsCompatibleWithVersion(%s) with MinimumVersion=%s should return %v",
				tt.targetVersion, tt.minimumVersion, tt.expected)
		})
	}
}

// TestContainerConfiguration_IsCompatibleWithVersion_InvalidVersions tests error handling
func TestContainerConfiguration_IsCompatibleWithVersion_InvalidVersions(t *testing.T) {
	tests := []struct {
		name           string
		minimumVersion string
		targetVersion  string
		expected       bool
	}{
		{
			name:           "empty minimum version",
			minimumVersion: "",
			targetVersion:  "v1.0.0",
			expected:       false,
		},
		{
			name:           "empty target version",
			minimumVersion: "v1.0.0",
			targetVersion:  "",
			expected:       false,
		},
		{
			name:           "invalid minimum version format",
			minimumVersion: "invalid",
			targetVersion:  "v1.0.0",
			expected:       false,
		},
		{
			name:           "invalid target version format",
			minimumVersion: "v1.0.0",
			targetVersion:  "invalid",
			expected:       false,
		},
		{
			name:           "incomplete version - only major",
			minimumVersion: "v1",
			targetVersion:  "v1.0.0",
			expected:       false,
		},
		{
			name:           "incomplete version - major.minor only",
			minimumVersion: "v1.0",
			targetVersion:  "v1.0.0",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &ContainerConfiguration{
				MinimumVersion: tt.minimumVersion,
			}

			result := config.IsCompatibleWithVersion(tt.targetVersion)
			assert.Equal(t, tt.expected, result,
				"Invalid version should return false")
		})
	}
}

// TestContainerVersion_CanUseConfiguration tests the reverse direction
func TestContainerVersion_CanUseConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		versionString  string
		minimumVersion string
		expected       bool
	}{
		{
			name:           "version meets minimum requirement",
			versionString:  "v1.1.0",
			minimumVersion: "v1.0.0",
			expected:       true,
		},
		{
			name:           "version below minimum requirement",
			versionString:  "v1.0.0",
			minimumVersion: "v1.1.0",
			expected:       false,
		},
		{
			name:           "version exactly meets requirement",
			versionString:  "v1.0.0",
			minimumVersion: "v1.0.0",
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := &ContainerVersion{
				Version: tt.versionString,
			}

			config := &ContainerConfiguration{
				MinimumVersion: tt.minimumVersion,
			}

			result := version.CanUseConfiguration(config)
			assert.Equal(t, tt.expected, result,
				"Version %s should %s use configuration with MinimumVersion %s",
				tt.versionString,
				map[bool]string{true: "be able to", false: "not be able to"}[tt.expected],
				tt.minimumVersion)
		})
	}
}
