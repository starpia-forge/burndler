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
