import React, { useEffect, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { ArrowLeftIcon, CubeIcon } from '@heroicons/react/24/outline';
import { useAuth } from '../hooks/useAuth';
import { Container, ContainerVersion, UpdateVersionRequest } from '../types/container';
import ContainerVersionForm from '../components/containers/ContainerVersionForm';
import containerService from '../services/containerService';

const EditContainerVersionPage: React.FC = () => {
  const { id, version: versionParam } = useParams<{ id: string; version: string }>();
  const navigate = useNavigate();
  const { canCreateContainer } = useAuth();
  const { t } = useTranslation(['containers', 'common']);

  const containerId = id ? parseInt(id, 10) : 0;

  const [container, setContainer] = useState<Container | null>(null);
  const [version, setVersion] = useState<ContainerVersion | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  // Fetch container and version details
  useEffect(() => {
    const fetchData = async () => {
      if (!containerId || !versionParam) return;

      try {
        setLoading(true);
        setError(null);

        // Fetch container
        const containerData = await containerService.getContainer(containerId, false);
        setContainer(containerData);

        // Fetch version
        const versionData = await containerService.getVersion(containerId, versionParam);
        setVersion(versionData);

        // Check if version is published
        if (versionData.published) {
          setError(t('containers:cannotEditPublishedVersion'));
        }
      } catch (err: any) {
        setError(err.message || t('containers:failedToFetch'));
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [containerId, versionParam, t]);

  const handleSubmit = async (data: UpdateVersionRequest) => {
    if (!versionParam) return;

    try {
      setSubmitting(true);
      setError(null);
      await containerService.updateVersion(containerId, versionParam, data);
      navigate(`/containers/${containerId}`);
    } catch (err: any) {
      setError(err.message || t('containers:failedToUpdate'));
    } finally {
      setSubmitting(false);
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
  if (loading) {
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
  if (error && (!container || !version || version.published)) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
            <h3 className="text-lg font-medium text-red-800 dark:text-red-300 mb-2">
              {version?.published
                ? t('containers:cannotEditPublishedVersion')
                : t('containers:errorLoading')}
            </h3>
            <p className="text-red-700 dark:text-red-400">{error}</p>
            <div className="mt-4">
              <Link
                to={`/containers/${containerId}`}
                className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-red-600 hover:bg-red-700"
              >
                <ArrowLeftIcon className="h-4 w-4 mr-2" />
                {t('containers:backToContainer')}
              </Link>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!container || !version) return null;

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

          <div className="flex items-start space-x-4">
            <div className="flex-shrink-0">
              <CubeIcon className="h-8 w-8 text-blue-500 dark:text-blue-400" />
            </div>
            <div className="min-w-0 flex-1">
              <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
                {t('containers:editVersionTitle', { version: version.version })}
              </h1>
              <p className="mt-1 text-gray-600 dark:text-gray-400">
                {t('containers:editVersionDescription', { containerName: container.name })}
              </p>
            </div>
          </div>
        </div>

        {/* Form */}
        <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
          <ContainerVersionForm
            mode="edit"
            initialData={version}
            onSubmit={handleSubmit}
            onCancel={handleCancel}
            loading={submitting}
            error={error}
          />
        </div>
      </div>
    </div>
  );
};

export default EditContainerVersionPage;
