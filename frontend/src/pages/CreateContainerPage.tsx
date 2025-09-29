import React from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { ArrowLeftIcon } from '@heroicons/react/24/outline';
import { useAuth } from '../hooks/useAuth';
import ContainerForm from '../components/containers/ContainerForm';
import { useCreateContainer } from '../hooks/useCreateContainer';
import { CreateContainerRequest, UpdateContainerRequest } from '../types/container';

const CreateContainerPage: React.FC = () => {
  const navigate = useNavigate();
  const { isDeveloper } = useAuth();
  const { t } = useTranslation(['containers', 'common']);
  const { createContainer, loading, error } = useCreateContainer();

  const handleSubmit = async (data: CreateContainerRequest | UpdateContainerRequest) => {
    // Type guard to ensure we have CreateContainerRequest
    if (!('name' in data)) {
      throw new Error('Invalid data for container creation');
    }
    const container = await createContainer(data);
    if (container) {
      navigate(`/containers/${container.id}`);
    }
  };

  const handleCancel = () => {
    navigate('/containers');
  };

  if (!isDeveloper) {
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

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Breadcrumb */}
        <div className="mb-8">
          <nav className="flex items-center space-x-2 text-sm text-gray-500 dark:text-gray-400">
            <Link to="/containers" className="hover:text-gray-700 dark:hover:text-gray-300">
              {t('containers:title')}
            </Link>
            <span className="mx-2">/</span>
            <span className="text-gray-900 dark:text-white">{t('containers:createContainer')}</span>
          </nav>
        </div>

        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center space-x-4">
            <Link
              to="/containers"
              className="inline-flex items-center text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300"
            >
              <ArrowLeftIcon className="h-4 w-4 mr-1" />
              {t('containers:backToContainers')}
            </Link>
          </div>
          <h1 className="mt-4 text-2xl font-bold text-gray-900 dark:text-white">
            {t('containers:createNewContainer')}
          </h1>
          <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
            {t('containers:createContainerDescription')}
          </p>
        </div>

        {/* Error display */}
        {error && (
          <div className="mb-6 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
            <p className="text-red-700 dark:text-red-400">{error}</p>
          </div>
        )}

        {/* Container Form */}
        <div className="max-w-2xl">
          <ContainerForm
            onSubmit={handleSubmit}
            onCancel={handleCancel}
            loading={loading}
            title={t('containers:createNewContainer')}
            submitLabel={t('containers:createContainer')}
          />
        </div>
      </div>
    </div>
  );
};

export default CreateContainerPage;
