import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import containerService from '../services/containerService';
import { CreateVersionRequest, ContainerVersion } from '../types/container';

interface UseCreateContainerVersionReturn {
  createVersion: (
    containerId: number,
    data: CreateVersionRequest
  ) => Promise<ContainerVersion | null>;
  loading: boolean;
  error: string | null;
  clearError: () => void;
}

export function useCreateContainerVersion(): UseCreateContainerVersionReturn {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { t } = useTranslation(['containers', 'common']);

  const createVersion = async (
    containerId: number,
    data: CreateVersionRequest
  ): Promise<ContainerVersion | null> => {
    try {
      setLoading(true);
      setError(null);

      const version = await containerService.createVersion(containerId, data);
      return version;
    } catch (error: any) {
      let errorMessage = t('containers:createVersionFailed');

      // Handle specific error cases
      if (error.status === 400) {
        if (error.message?.includes('already exists')) {
          errorMessage = t('containers:versionAlreadyExists', { version: data.version });
        } else if (error.message?.includes('validation')) {
          errorMessage = t('containers:validationFailed');
        } else {
          errorMessage = error.message || errorMessage;
        }
      } else if (error.status === 401) {
        errorMessage = t('containers:unauthorized');
      } else if (error.status === 403) {
        errorMessage = t('containers:accessDenied');
      } else if (error.status === 404) {
        errorMessage = t('containers:containerNotFound');
      } else if (error.status === 409) {
        errorMessage = t('containers:versionAlreadyExists', { version: data.version });
      } else if (error.status >= 500) {
        errorMessage = t('containers:serverError');
      } else if (error.name === 'NetworkError') {
        errorMessage = t('containers:networkError');
      }

      setError(errorMessage);
      return null;
    } finally {
      setLoading(false);
    }
  };

  const clearError = () => {
    setError(null);
  };

  return {
    createVersion,
    loading,
    error,
    clearError,
  };
}
