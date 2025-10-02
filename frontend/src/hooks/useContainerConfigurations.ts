import { useState, useEffect, useCallback } from 'react';
import {
  listContainerConfigurations,
  createContainerConfiguration,
  updateContainerConfiguration,
  deleteContainerConfiguration,
  ContainerConfiguration,
  CreateContainerConfigurationRequest,
  UpdateContainerConfigurationRequest,
} from '../services/configurationService';
import { ContainerVersion } from '../types/container';
import { isVersionCompatible } from '../utils/versionCompatibility';

export interface UseContainerConfigurationsOptions {
  containerId: string;
  autoFetch?: boolean;
}

export interface UseContainerConfigurationsReturn {
  configurations: ContainerConfiguration[];
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
  createConfig: (data: CreateContainerConfigurationRequest) => Promise<ContainerConfiguration>;
  updateConfig: (
    name: string,
    data: UpdateContainerConfigurationRequest
  ) => Promise<ContainerConfiguration>;
  deleteConfig: (name: string) => Promise<void>;
  getConfigForVersion: (versionId: number) => ContainerConfiguration | null;
  getCompatibleVersions: (configName: string, versions: ContainerVersion[]) => ContainerVersion[];
  isConfigInUse: (configId: number, versions: ContainerVersion[]) => boolean;
  getVersionsUsingConfig: (configId: number, versions: ContainerVersion[]) => ContainerVersion[];
}

export function useContainerConfigurations({
  containerId,
  autoFetch = true,
}: UseContainerConfigurationsOptions): UseContainerConfigurationsReturn {
  const [configurations, setConfigurations] = useState<ContainerConfiguration[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchConfigurations = useCallback(async () => {
    if (!containerId) return;

    try {
      setLoading(true);
      setError(null);
      const data = await listContainerConfigurations(containerId);
      setConfigurations(data);
    } catch (err: any) {
      const errorMessage =
        err.response?.data?.error || err.message || 'Failed to load configurations';
      setError(errorMessage);
      console.error('Failed to fetch configurations:', err);
    } finally {
      setLoading(false);
    }
  }, [containerId]);

  useEffect(() => {
    if (autoFetch) {
      fetchConfigurations();
    }
  }, [autoFetch, fetchConfigurations]);

  const createConfig = useCallback(
    async (data: CreateContainerConfigurationRequest): Promise<ContainerConfiguration> => {
      try {
        setError(null);
        const newConfig = await createContainerConfiguration(containerId, data);
        setConfigurations((prev) => [...prev, newConfig]);
        return newConfig;
      } catch (err: any) {
        const errorMessage =
          err.response?.data?.error || err.message || 'Failed to create configuration';
        setError(errorMessage);
        throw err;
      }
    },
    [containerId]
  );

  const updateConfig = useCallback(
    async (
      name: string,
      data: UpdateContainerConfigurationRequest
    ): Promise<ContainerConfiguration> => {
      try {
        setError(null);
        const updatedConfig = await updateContainerConfiguration(containerId, name, data);
        setConfigurations((prev) => prev.map((c) => (c.name === name ? updatedConfig : c)));
        return updatedConfig;
      } catch (err: any) {
        const errorMessage =
          err.response?.data?.error || err.message || 'Failed to update configuration';
        setError(errorMessage);
        throw err;
      }
    },
    [containerId]
  );

  const deleteConfig = useCallback(
    async (name: string): Promise<void> => {
      try {
        setError(null);
        await deleteContainerConfiguration(containerId, name);
        setConfigurations((prev) => prev.filter((c) => c.name !== name));
      } catch (err: any) {
        const errorMessage =
          err.response?.data?.error || err.message || 'Failed to delete configuration';
        setError(errorMessage);
        throw err;
      }
    },
    [containerId]
  );

  const getConfigForVersion = useCallback(
    (versionId: number): ContainerConfiguration | null => {
      return configurations.find((c) => c.id === versionId) || null;
    },
    [configurations]
  );

  const getCompatibleVersions = useCallback(
    (configName: string, versions: ContainerVersion[]): ContainerVersion[] => {
      const config = configurations.find((c) => c.name === configName);
      if (!config) return [];

      return versions.filter((v) => isVersionCompatible(v.version, config.minimum_version));
    },
    [configurations]
  );

  const isConfigInUse = useCallback((configId: number, versions: ContainerVersion[]): boolean => {
    return versions.some((v) => v.configuration_id === configId);
  }, []);

  const getVersionsUsingConfig = useCallback(
    (configId: number, versions: ContainerVersion[]): ContainerVersion[] => {
      return versions.filter((v) => v.configuration_id === configId);
    },
    []
  );

  return {
    configurations,
    loading,
    error,
    refetch: fetchConfigurations,
    createConfig,
    updateConfig,
    deleteConfig,
    getConfigForVersion,
    getCompatibleVersions,
    isConfigInUse,
    getVersionsUsingConfig,
  };
}
