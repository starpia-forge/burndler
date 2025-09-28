import { useState, useEffect, useCallback, useMemo } from 'react';
import containerService from '../services/containerService';
import {
  ContainerVersion,
  VersionFilters,
  ContainerVersionState,
  CreateVersionRequest,
  UpdateVersionRequest,
  ApiError,
} from '../types/container';

export interface UseContainerVersionsOptions {
  containerId: number;
  autoFetch?: boolean;
  initialFilters?: VersionFilters;
}

export interface UseContainerVersionsReturn {
  versions: ContainerVersion[];
  loading: boolean;
  error: string | null;
  filters: VersionFilters;
  refetch: () => Promise<void>;
  setFilters: (filters: VersionFilters) => void;
  createVersion: (data: CreateVersionRequest) => Promise<ContainerVersion>;
  updateVersion: (version: string, data: UpdateVersionRequest) => Promise<ContainerVersion>;
  publishVersion: (version: string) => Promise<ContainerVersion>;
  getVersion: (version: string) => Promise<ContainerVersion>;
  refreshVersion: (version: string) => Promise<void>;
  getLatestVersion: () => ContainerVersion | null;
  getPublishedVersions: () => ContainerVersion[];
  getDraftVersions: () => ContainerVersion[];
}

const DEFAULT_FILTERS: VersionFilters = {
  published_only: false,
};

export function useContainerVersions(options: UseContainerVersionsOptions): UseContainerVersionsReturn {
  const { containerId, autoFetch = true, initialFilters = {} } = options;

  const [state, setState] = useState<ContainerVersionState>({
    versions: [],
    loading: false,
    error: null,
  });

  const [filters, setFilters] = useState<VersionFilters>({
    ...DEFAULT_FILTERS,
    ...initialFilters,
  });

  // Fetch versions based on current filters
  const fetchVersions = useCallback(async () => {
    if (!containerId) return;

    setState((prev) => ({ ...prev, loading: true, error: null }));

    try {
      const response = await containerService.listVersions(containerId, filters);
      setState((prev) => ({
        ...prev,
        versions: response.data,
        loading: false,
      }));
    } catch (error: any) {
      const apiError = error as ApiError;
      setState((prev) => ({
        ...prev,
        error: apiError.message || 'Failed to fetch versions',
        loading: false,
      }));
    }
  }, [containerId, filters]);

  // Refetch with current filters
  const refetch = useCallback(async () => {
    await fetchVersions();
  }, [fetchVersions]);

  // Create new version
  const createVersion = useCallback(
    async (data: CreateVersionRequest): Promise<ContainerVersion> => {
      try {
        setState((prev) => ({ ...prev, loading: true, error: null }));

        const newVersion = await containerService.createVersion(containerId, data);

        // Add to local state
        setState((prev) => ({
          ...prev,
          versions: [newVersion, ...prev.versions],
          loading: false,
        }));

        return newVersion;
      } catch (error: any) {
        const apiError = error as ApiError;
        setState((prev) => ({
          ...prev,
          error: apiError.message || 'Failed to create version',
          loading: false,
        }));
        throw error;
      }
    },
    [containerId]
  );

  // Update existing version
  const updateVersion = useCallback(
    async (version: string, data: UpdateVersionRequest): Promise<ContainerVersion> => {
      try {
        setState((prev) => ({ ...prev, loading: true, error: null }));

        const updatedVersion = await containerService.updateVersion(containerId, version, data);

        // Update in local state
        setState((prev) => ({
          ...prev,
          versions: prev.versions.map((v) => (v.version === version ? updatedVersion : v)),
          loading: false,
        }));

        return updatedVersion;
      } catch (error: any) {
        const apiError = error as ApiError;
        setState((prev) => ({
          ...prev,
          error: apiError.message || 'Failed to update version',
          loading: false,
        }));
        throw error;
      }
    },
    [containerId]
  );

  // Publish version
  const publishVersion = useCallback(
    async (version: string): Promise<ContainerVersion> => {
      try {
        setState((prev) => ({ ...prev, loading: true, error: null }));

        const publishedVersion = await containerService.publishVersion(containerId, version);

        // Update in local state
        setState((prev) => ({
          ...prev,
          versions: prev.versions.map((v) => (v.version === version ? publishedVersion : v)),
          loading: false,
        }));

        return publishedVersion;
      } catch (error: any) {
        const apiError = error as ApiError;
        setState((prev) => ({
          ...prev,
          error: apiError.message || 'Failed to publish version',
          loading: false,
        }));
        throw error;
      }
    },
    [containerId]
  );

  // Get specific version
  const getVersion = useCallback(
    async (version: string): Promise<ContainerVersion> => {
      try {
        const versionData = await containerService.getVersion(containerId, version);
        return versionData;
      } catch (error: any) {
        const apiError = error as ApiError;
        setState((prev) => ({
          ...prev,
          error: apiError.message || 'Failed to get version',
        }));
        throw error;
      }
    },
    [containerId]
  );

  // Refresh single version
  const refreshVersion = useCallback(
    async (version: string) => {
      try {
        const updatedVersion = await containerService.getVersion(containerId, version);
        setState((prev) => ({
          ...prev,
          versions: prev.versions.map((v) => (v.version === version ? updatedVersion : v)),
        }));
      } catch (error: any) {
        const apiError = error as ApiError;
        setState((prev) => ({
          ...prev,
          error: apiError.message || 'Failed to refresh version',
        }));
      }
    },
    [containerId]
  );

  // Get latest version (by semantic version)
  const getLatestVersion = useCallback((): ContainerVersion | null => {
    if (state.versions.length === 0) return null;

    // Sort versions by semantic version descending
    const sortedVersions = [...state.versions].sort((a, b) => {
      const aVersion = a.version.replace(/^v/, '');
      const bVersion = b.version.replace(/^v/, '');

      // Simple semantic version comparison
      const aParts = aVersion.split('.').map(Number);
      const bParts = bVersion.split('.').map(Number);

      for (let i = 0; i < Math.max(aParts.length, bParts.length); i++) {
        const aPart = aParts[i] || 0;
        const bPart = bParts[i] || 0;

        if (aPart > bPart) return -1;
        if (aPart < bPart) return 1;
      }

      return 0;
    });

    return sortedVersions[0];
  }, [state.versions]);

  // Get published versions only
  const getPublishedVersions = useCallback((): ContainerVersion[] => {
    return state.versions.filter((version) => version.published);
  }, [state.versions]);

  // Get draft versions only
  const getDraftVersions = useCallback((): ContainerVersion[] => {
    return state.versions.filter((version) => !version.published);
  }, [state.versions]);

  // Auto-fetch on mount and when filters change
  useEffect(() => {
    if (autoFetch && containerId) {
      fetchVersions();
    }
  }, [autoFetch, containerId, fetchVersions]);

  // Memoized return value to prevent unnecessary re-renders
  const returnValue = useMemo(
    () => ({
      versions: state.versions,
      loading: state.loading,
      error: state.error,
      filters,
      refetch,
      setFilters,
      createVersion,
      updateVersion,
      publishVersion,
      getVersion,
      refreshVersion,
      getLatestVersion,
      getPublishedVersions,
      getDraftVersions,
    }),
    [
      state.versions,
      state.loading,
      state.error,
      filters,
      refetch,
      createVersion,
      updateVersion,
      publishVersion,
      getVersion,
      refreshVersion,
      getLatestVersion,
      getPublishedVersions,
      getDraftVersions,
    ]
  );

  return returnValue;
}

// Helper hook for version validation
export function useVersionValidation() {
  const validateSemVer = useCallback((version: string): { isValid: boolean; error?: string } => {
    if (!version.trim()) {
      return { isValid: false, error: 'Version is required' };
    }

    const cleanVersion = version.replace(/^v/, '');
    const semverRegex =
      /^(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$/;

    if (!semverRegex.test(cleanVersion)) {
      return {
        isValid: false,
        error: 'Version must follow semantic versioning format (e.g., 1.0.0)',
      };
    }

    return { isValid: true };
  }, []);

  const validateCompose = useCallback((compose: string): { isValid: boolean; error?: string } => {
    if (!compose.trim()) {
      return { isValid: false, error: 'Compose content is required' };
    }

    // Basic YAML validation (check for basic structure)
    try {
      const lines = compose.split('\n');
      const hasServices = lines.some((line) => line.trim().startsWith('services:'));

      if (!hasServices) {
        return { isValid: false, error: 'Compose file must contain a services section' };
      }

      return { isValid: true };
    } catch (error) {
      return { isValid: false, error: 'Invalid YAML format' };
    }
  }, []);

  return {
    validateSemVer,
    validateCompose,
  };
}