package models

import (
	"strconv"
	"strings"
)

// IsCompatibleWithVersion checks if this configuration can be used with the given version.
// Returns true if the target version is >= MinimumVersion.
func (cc *ContainerConfiguration) IsCompatibleWithVersion(targetVersion string) bool {
	if cc.MinimumVersion == "" || targetVersion == "" {
		return false
	}

	minVer, err := parseSemanticVersion(cc.MinimumVersion)
	if err != nil {
		return false
	}

	targetVer, err := parseSemanticVersion(targetVersion)
	if err != nil {
		return false
	}

	return compareVersions(targetVer, minVer) >= 0
}

// CanUseConfiguration checks if this version can use the given configuration.
// Returns true if this version >= configuration's MinimumVersion.
func (cv *ContainerVersion) CanUseConfiguration(config *ContainerConfiguration) bool {
	return config.IsCompatibleWithVersion(cv.Version)
}

// semanticVersion represents a parsed semantic version
type semanticVersion struct {
	major int
	minor int
	patch int
}

// parseSemanticVersion parses a semantic version string (e.g., "v1.2.3" or "1.2.3")
// into major, minor, and patch components.
func parseSemanticVersion(version string) (*semanticVersion, error) {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	// Split by '.'
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return nil, &versionParseError{version: version, reason: "must have exactly 3 parts (major.minor.patch)"}
	}

	// Parse each component
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, &versionParseError{version: version, reason: "invalid major version"}
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, &versionParseError{version: version, reason: "invalid minor version"}
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, &versionParseError{version: version, reason: "invalid patch version"}
	}

	return &semanticVersion{
		major: major,
		minor: minor,
		patch: patch,
	}, nil
}

// compareVersions compares two semantic versions.
// Returns:
//   - positive if v1 > v2
//   - zero if v1 == v2
//   - negative if v1 < v2
func compareVersions(v1, v2 *semanticVersion) int {
	// Compare major version
	if v1.major != v2.major {
		return v1.major - v2.major
	}

	// Compare minor version
	if v1.minor != v2.minor {
		return v1.minor - v2.minor
	}

	// Compare patch version
	return v1.patch - v2.patch
}

// versionParseError represents an error parsing a version string
type versionParseError struct {
	version string
	reason  string
}

func (e *versionParseError) Error() string {
	return "invalid version '" + e.version + "': " + e.reason
}
