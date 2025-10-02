/**
 * Version Compatibility Utilities
 * Provides semantic version parsing and comparison for container configurations
 */

export interface SemanticVersion {
  major: number;
  minor: number;
  patch: number;
}

/**
 * Parse a semantic version string into components
 * Supports both "v1.2.3" and "1.2.3" formats
 */
export function parseSemanticVersion(versionStr: string): SemanticVersion | null {
  const cleaned = versionStr.replace(/^v/, '');
  const parts = cleaned.split('.');

  if (parts.length !== 3) {
    return null;
  }

  const [major, minor, patch] = parts.map(Number);

  if (isNaN(major) || isNaN(minor) || isNaN(patch)) {
    return null;
  }

  return { major, minor, patch };
}

/**
 * Check if a version string is valid semantic version format
 */
export function isValidSemanticVersion(versionStr: string): boolean {
  return parseSemanticVersion(versionStr) !== null;
}

/**
 * Compare two semantic versions
 * Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
 */
export function compareVersions(v1: SemanticVersion, v2: SemanticVersion): number {
  if (v1.major !== v2.major) {
    return v1.major - v2.major;
  }

  if (v1.minor !== v2.minor) {
    return v1.minor - v2.minor;
  }

  return v1.patch - v2.patch;
}

/**
 * Check if a version meets the minimum version requirement
 * @param versionStr - The version to check (e.g., "v1.5.0")
 * @param minimumVersionStr - The minimum required version (e.g., "v1.0.0")
 * @returns true if version >= minimumVersion
 */
export function isVersionCompatible(
  versionStr: string,
  minimumVersionStr: string
): boolean {
  const version = parseSemanticVersion(versionStr);
  const minVersion = parseSemanticVersion(minimumVersionStr);

  if (!version || !minVersion) {
    return false;
  }

  return compareVersions(version, minVersion) >= 0;
}

/**
 * Get compatibility status for a version and configuration
 */
export type CompatibilityStatus = 'assigned' | 'compatible' | 'incompatible';

export interface VersionCompatibilityInfo {
  status: CompatibilityStatus;
  message: string;
}

/**
 * Get detailed compatibility information
 */
export function getCompatibilityInfo(
  versionStr: string,
  configMinVersion: string,
  isAssigned: boolean
): VersionCompatibilityInfo {
  if (isAssigned) {
    return {
      status: 'assigned',
      message: 'Currently assigned to this version',
    };
  }

  const compatible = isVersionCompatible(versionStr, configMinVersion);

  if (compatible) {
    return {
      status: 'compatible',
      message: `Version ${versionStr} meets minimum requirement ${configMinVersion}`,
    };
  }

  return {
    status: 'incompatible',
    message: `Version ${versionStr} is below minimum requirement ${configMinVersion}`,
  };
}

/**
 * Format a semantic version for display
 */
export function formatVersion(versionStr: string): string {
  // Ensure version starts with 'v'
  return versionStr.startsWith('v') ? versionStr : `v${versionStr}`;
}
