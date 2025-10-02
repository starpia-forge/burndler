import api from './api';
import {
  UISchema,
  ConfigurationValues,
  DependencyRule,
  ContainerFile,
  ContainerAsset,
} from '../types/configuration';

export interface ServiceContainerConfigurationResponse {
  ui_schema: UISchema;
  dependency_rules: DependencyRule[];
  current_values: ConfigurationValues;
  files: ContainerFile[];
  assets: ContainerAsset[];
}

export interface SaveConfigurationRequest {
  configuration_values: ConfigurationValues;
}

export interface SaveConfigurationResponse {
  message: string;
  config: {
    id: number;
    service_id: number;
    container_id: number;
    configuration_values: ConfigurationValues;
  };
}

/**
 * Fetch configuration schema and current values for a service container
 */
export const getServiceContainerConfiguration = async (
  serviceId: string,
  containerId: string
): Promise<ServiceContainerConfigurationResponse> => {
  return await api.get<ServiceContainerConfigurationResponse>(
    `/services/${serviceId}/containers/${containerId}/configuration`
  );
};

/**
 * Save configuration values for a service container
 */
export const saveServiceContainerConfiguration = async (
  serviceId: string,
  containerId: string,
  values: ConfigurationValues
): Promise<SaveConfigurationResponse> => {
  return await api.put<SaveConfigurationResponse>(
    `/services/${serviceId}/containers/${containerId}/configuration`,
    {
      configuration_values: values,
    }
  );
};

/**
 * Validate configuration values against dependency rules
 */
export const validateConfiguration = async (
  serviceId: string,
  containerId: string,
  values: ConfigurationValues
): Promise<{ valid: boolean; errors: Array<{ field: string; message: string; rule: string }> }> => {
  const response = await api.post(`/services/${serviceId}/containers/${containerId}/validate`, {
    values,
  });
  return response.data;
};

/**
 * Export all service configurations as JSON
 */
export const exportServiceConfiguration = async (serviceId: string): Promise<Blob> => {
  const response = await api.get(`/services/${serviceId}/configuration/export`, {
    responseType: 'blob',
  });
  return response.data;
};

/**
 * Import service configurations from JSON file
 */
export const importServiceConfiguration = async (
  serviceId: string,
  data: unknown
): Promise<{ message: string; imported: number; skipped?: string[] }> => {
  const response = await api.post(`/services/${serviceId}/configuration/import`, data);
  return response.data;
};

// ============================================================================
// Container-Level Configuration Management (Phase 6)
// ============================================================================

export interface ContainerConfiguration {
  id: number;
  container_id: number;
  name: string;
  minimum_version: string;
  description?: string;
  ui_schema?: UISchema;
  dependency_rules?: DependencyRule[];
  files?: ContainerFile[];
  assets?: ContainerAsset[];
  created_at?: string;
  updated_at?: string;
}

export interface CreateContainerConfigurationRequest {
  name: string;
  minimum_version: string;
  description?: string;
  ui_schema?: UISchema;
  dependency_rules?: DependencyRule[];
}

export interface UpdateContainerConfigurationRequest {
  minimum_version?: string;
  description?: string;
  ui_schema?: UISchema;
  dependency_rules?: DependencyRule[];
}

/**
 * List all configurations for a container
 */
export const listContainerConfigurations = async (
  containerId: string
): Promise<ContainerConfiguration[]> => {
  return await api.get<ContainerConfiguration[]>(`/containers/${containerId}/configurations`);
};

/**
 * Create a new configuration for a container
 */
export const createContainerConfiguration = async (
  containerId: string,
  data: CreateContainerConfigurationRequest
): Promise<ContainerConfiguration> => {
  return await api.post<ContainerConfiguration>(`/containers/${containerId}/configurations`, data);
};

/**
 * Get a specific configuration by name
 */
export const getContainerConfiguration = async (
  containerId: string,
  name: string
): Promise<ContainerConfiguration> => {
  return await api.get<ContainerConfiguration>(`/containers/${containerId}/configurations/${name}`);
};

/**
 * Update a configuration
 */
export const updateContainerConfiguration = async (
  containerId: string,
  name: string,
  data: UpdateContainerConfigurationRequest
): Promise<ContainerConfiguration> => {
  return await api.put<ContainerConfiguration>(
    `/containers/${containerId}/configurations/${name}`,
    data
  );
};

/**
 * Delete a configuration
 */
export const deleteContainerConfiguration = async (
  containerId: string,
  name: string
): Promise<{ message: string }> => {
  return await api.delete<{ message: string }>(`/containers/${containerId}/configurations/${name}`);
};
