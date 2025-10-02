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
  const response = await api.get<ServiceContainerConfigurationResponse>(
    `/services/${serviceId}/containers/${containerId}/configuration`
  );
  return response.data;
};

/**
 * Save configuration values for a service container
 */
export const saveServiceContainerConfiguration = async (
  serviceId: string,
  containerId: string,
  values: ConfigurationValues
): Promise<SaveConfigurationResponse> => {
  const response = await api.put<SaveConfigurationResponse>(
    `/services/${serviceId}/containers/${containerId}/configuration`,
    {
      configuration_values: values,
    }
  );
  return response.data;
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
  ui_schema?: UISchema;
  dependency_rules?: DependencyRule[];
}

export interface UpdateContainerConfigurationRequest {
  minimum_version?: string;
  ui_schema?: UISchema;
  dependency_rules?: DependencyRule[];
}

/**
 * List all configurations for a container
 */
export const listContainerConfigurations = async (
  containerId: string
): Promise<ContainerConfiguration[]> => {
  const response = await api.get<ContainerConfiguration[]>(
    `/containers/${containerId}/configurations`
  );
  return response.data;
};

/**
 * Create a new configuration for a container
 */
export const createContainerConfiguration = async (
  containerId: string,
  data: CreateContainerConfigurationRequest
): Promise<ContainerConfiguration> => {
  const response = await api.post<ContainerConfiguration>(
    `/containers/${containerId}/configurations`,
    data
  );
  return response.data;
};

/**
 * Get a specific configuration by name
 */
export const getContainerConfiguration = async (
  containerId: string,
  name: string
): Promise<ContainerConfiguration> => {
  const response = await api.get<ContainerConfiguration>(
    `/containers/${containerId}/configurations/${name}`
  );
  return response.data;
};

/**
 * Update a configuration
 */
export const updateContainerConfiguration = async (
  containerId: string,
  name: string,
  data: UpdateContainerConfigurationRequest
): Promise<ContainerConfiguration> => {
  const response = await api.put<ContainerConfiguration>(
    `/containers/${containerId}/configurations/${name}`,
    data
  );
  return response.data;
};

/**
 * Delete a configuration
 */
export const deleteContainerConfiguration = async (
  containerId: string,
  name: string
): Promise<{ message: string }> => {
  const response = await api.delete<{ message: string }>(
    `/containers/${containerId}/configurations/${name}`
  );
  return response.data;
};
