import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Container, CreateContainerRequest } from '../types/container';
import containerService from '../services/containerService';

interface UseCreateContainerReturn {
  createContainer: (data: CreateContainerRequest) => Promise<Container | null>;
  loading: boolean;
  error: string | null;
}

export const useCreateContainer = (): UseCreateContainerReturn => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { t } = useTranslation(['containers', 'common']);

  const mapApiError = (err: any): string => {
    // Handle specific API error responses
    if (err.response?.data?.error) {
      const errorCode = err.response.data.error;
      const message = err.response.data.message;

      switch (errorCode) {
        case 'MODULE_EXISTS':
        case 'CONTAINER_EXISTS':
          return t('containers:containerAlreadyExists', { name: extractNameFromMessage(message) });
        case 'VALIDATION_FAILED':
          return t('containers:validationFailed');
        case 'INSUFFICIENT_PERMISSIONS':
          return t('containers:developerRequired');
        default:
          return message || t('containers:createFailed');
      }
    }

    // Handle network errors
    if (err.code === 'NETWORK_ERROR' || !err.response) {
      return t('common:networkError');
    }

    // Handle other HTTP errors
    if (err.response?.status) {
      switch (err.response.status) {
        case 400:
          return t('containers:validationFailed');
        case 401:
          return t('common:unauthorized');
        case 403:
          return t('containers:developerRequired');
        case 409:
          return t('containers:containerNameExists');
        case 500:
          return t('common:serverError');
        default:
          return t('containers:createFailed');
      }
    }

    return t('containers:createFailed');
  };

  const extractNameFromMessage = (message: string): string => {
    // Extract container name from error message like "container with name 'nginx' already exists"
    const match = message.match(/name '([^']+)'/);
    return match ? match[1] : '';
  };

  const createContainer = async (data: CreateContainerRequest): Promise<Container | null> => {
    setLoading(true);
    setError(null);

    try {
      const container = await containerService.createContainer(data);
      return container;
    } catch (err: any) {
      const errorMessage = mapApiError(err);
      setError(errorMessage);
      console.error('Failed to create container:', err);
      return null;
    } finally {
      setLoading(false);
    }
  };

  return {
    createContainer,
    loading,
    error,
  };
};
