import React, { useEffect, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { ArrowLeftIcon } from '@heroicons/react/24/outline';
import { useAuth } from '../hooks/useAuth';
import { Container } from '../types/container';
import { useUpdateContainer } from '../hooks/useUpdateContainer';
import { UpdateContainerRequest } from '../types/container';
import ContainerForm from '../components/containers/ContainerForm';
import containerService from '../services/containerService';

const EditContainerPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { canCreateContainer } = useAuth();
  const { t } = useTranslation(['containers', 'common']);

  const containerId = id ? parseInt(id, 10) : 0;
  const { updateContainer, loading, error } = useUpdateContainer();

  const [container, setContainer] = useState<Container | null>(null);
  const [containerLoading, setContainerLoading] = useState(true);
  const [containerError, setContainerError] = useState<string | null>(null);

  // Fetch container details
  useEffect(() => {
    const fetchContainer = async () => {
      if (!containerId) return;

      try {
        setContainerLoading(true);
        setContainerError(null);
        const containerData = await containerService.getContainer(containerId, false);
        setContainer(containerData);
      } catch (error: any) {
        if (error.status === 404) {
          setContainerError(t('containers:containerNotFound'));
        } else {
          setContainerError(error.message || t('containers:failedToFetch'));
        }
      } finally {
        setContainerLoading(false);
      }
    };

    fetchContainer();
  }, [containerId, t]);

  const handleSubmit = async (data: UpdateContainerRequest) => {
    const updatedContainer = await updateContainer(containerId, data);
    if (updatedContainer) {
      navigate(`/containers/${containerId}`);
    }
  };

  const handleCancel = () => {
    navigate(`/containers/${containerId}`);
  };

  // Access control
  if (!canCreateContainer) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
            <h3 className="text-lg font-medium text-red-800 dark:text-red-300 mb-2">
              {t('common:accessDenied')}
            </h3>
            <p className="text-red-700 dark:text-red-400">{t('containers:developerRequired')}</p>
            <div className="mt-4">
              <Link
                to="/containers"
                className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-red-600 hover:bg-red-700"
              >
                <ArrowLeftIcon className="h-4 w-4 mr-2" />
                {t('containers:backToContainers')}
              </Link>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Loading state
  if (containerLoading) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="animate-pulse">
            <div className="h-8 bg-gray-300 dark:bg-gray-600 rounded w-1/4 mb-6"></div>
            <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
              <div className="h-6 bg-gray-300 dark:bg-gray-600 rounded w-1/3 mb-4"></div>
              <div className="h-4 bg-gray-300 dark:bg-gray-600 rounded w-2/3 mb-2"></div>
              <div className="h-4 bg-gray-300 dark:bg-gray-600 rounded w-1/2"></div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Error state
  if (containerError && !container) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
            <h3 className="text-lg font-medium text-red-800 dark:text-red-300 mb-2">
              {t('containers:errorLoading')}
            </h3>
            <p className="text-red-700 dark:text-red-400">{containerError}</p>
            <div className="mt-4">
              <Link
                to="/containers"
                className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-red-600 hover:bg-red-700"
              >
                <ArrowLeftIcon className="h-4 w-4 mr-2" />
                {t('containers:backToContainers')}
              </Link>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!container) return null;

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center space-x-4 mb-4">
            <Link
              to={`/containers/${containerId}`}
              className="inline-flex items-center text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300"
            >
              <ArrowLeftIcon className="h-4 w-4 mr-1" />
              {t('containers:backToContainer')}
            </Link>
          </div>

          <div className="mb-6">
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
              {t('containers:editContainerTitle')}
            </h1>
            <p className="mt-1 text-gray-600 dark:text-gray-400">
              {t('containers:editContainerDescription', { containerName: container.name })}
            </p>
          </div>
        </div>

        {/* Error display */}
        {error && (
          <div className="mb-6 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
            <p className="text-red-700 dark:text-red-400">{error}</p>
          </div>
        )}

        {/* Form */}
        <ContainerForm
          onSubmit={handleSubmit}
          onCancel={handleCancel}
          loading={loading}
          initialData={container}
          isEditMode={true}
          title={t('containers:editContainerTitle')}
          submitLabel={t('containers:updateContainer')}
        />
      </div>
    </div>
  );
};

export default EditContainerPage;
