import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import containerService from '../services/containerService';
import { UpdateContainerRequest, Container } from '../types/container';

interface UseUpdateContainerReturn {
  updateContainer: (id: number, data: UpdateContainerRequest) => Promise<Container | null>;
  loading: boolean;
  error: string | null;
  clearError: () => void;
}

export function useUpdateContainer(): UseUpdateContainerReturn {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { t } = useTranslation(['containers', 'common']);

  const updateContainer = async (
    id: number,
    data: UpdateContainerRequest
  ): Promise<Container | null> => {
    try {
      setLoading(true);
      setError(null);

      const container = await containerService.updateContainer(id, data);
      return container;
    } catch (error: any) {
      let errorMessage = t('containers:updateFailed');

      // Handle specific error cases
      if (error.status === 400) {
        if (error.message?.includes('already exists')) {
          errorMessage = t('containers:containerNameExists');
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
        errorMessage = t('containers:containerNameExists');
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
    updateContainer,
    loading,
    error,
    clearError,
  };
}
