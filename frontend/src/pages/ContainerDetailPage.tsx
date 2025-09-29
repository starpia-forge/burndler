import React, { useEffect, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  ArrowLeftIcon,
  PencilIcon,
  TrashIcon,
  PlusIcon,
  EyeIcon,
  ClockIcon,
  UserIcon,
  LinkIcon,
  CubeIcon,
} from '@heroicons/react/24/outline';
import { Container, ContainerVersion } from '../types/container';
import {
  StatusBadge,
  getContainerStatus,
  getContainerVersionStatus,
} from '../components/common/StatusBadge';
import { useAuth } from '../hooks/useAuth';
import containerService from '../services/containerService';
import { useContainerVersions } from '../hooks/useContainerVersions';
import { useConfirmationModal } from '../hooks/useConfirmationModal';
import ConfirmationModal from '../components/common/ConfirmationModal';

const ContainerDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { isDeveloper } = useAuth();
  const { t } = useTranslation(['containers', 'common']);

  const [container, setContainer] = useState<Container | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const containerId = id ? parseInt(id, 10) : 0;

  const {
    versions,
    loading: versionsLoading,
    error: versionsError,
    refetch: refetchVersions,
    publishVersion,
  } = useContainerVersions({
    containerId,
    autoFetch: true,
  });

  const confirmationModal = useConfirmationModal();

  // Fetch container details
  useEffect(() => {
    const fetchContainer = async () => {
      if (!containerId) return;

      try {
        setLoading(true);
        setError(null);
        const containerData = await containerService.getContainer(containerId, false);
        setContainer(containerData);
      } catch (error: any) {
        setError(error.message || t('containers:failedToFetch'));
      } finally {
        setLoading(false);
      }
    };

    fetchContainer();
  }, [containerId]);

  const handleEdit = () => {
    navigate(`/containers/${containerId}/edit`);
  };

  const handleDelete = async () => {
    if (!container) return;

    const placeholder = '__CONTAINER_NAME__';
    const translatedTemplate = t('containers:confirmDelete', { name: placeholder });
    const parts = translatedTemplate.split(placeholder);

    const deleteMessage = (
      <>
        {parts[0]}
        <span className="font-bold">{container.name}</span>
        {parts[1]}
      </>
    );

    confirmationModal.openModal({
      title: t('containers:deleteContainerTitle'),
      message: deleteMessage,
      confirmLabel: t('containers:deleteContainer'),
      variant: 'danger',
      onConfirm: async () => {
        try {
          await containerService.deleteContainer(containerId);
          navigate('/containers');
        } catch (error: any) {
          setError(error.message || t('containers:failedToDelete'));
          throw error; // Re-throw to keep modal open on error
        }
      },
    });
  };

  const handleCreateVersion = () => {
    navigate(`/containers/${containerId}/versions/create`);
  };

  const handlePublishVersion = async (version: ContainerVersion) => {
    try {
      await publishVersion(version.version);
    } catch (error: any) {
      setError(error.message || t('containers:failedToPublish'));
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

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

  if (error && !container) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
            <h3 className="text-lg font-medium text-red-800 dark:text-red-300 mb-2">
              {t('containers:errorLoading')}
            </h3>
            <p className="text-red-700 dark:text-red-400">{error}</p>
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

  const status = getContainerStatus(container);
  const publishedVersions = versions.filter((v) => v.published);

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <Link
                to="/containers"
                className="inline-flex items-center text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300"
              >
                <ArrowLeftIcon className="h-4 w-4 mr-1" />
                {t('containers:backToContainers')}
              </Link>
            </div>

            {isDeveloper && status !== 'deleted' && (
              <div className="flex items-center space-x-3">
                <button
                  onClick={handleEdit}
                  className="inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700"
                >
                  <PencilIcon className="h-4 w-4 mr-2" />
                  {t('containers:editContainer')}
                </button>
                <button
                  onClick={handleDelete}
                  className="inline-flex items-center px-4 py-2 border border-red-300 dark:border-red-600 rounded-md shadow-sm text-sm font-medium text-red-700 dark:text-red-300 bg-white dark:bg-gray-800 hover:bg-red-50 dark:hover:bg-red-900/20"
                >
                  <TrashIcon className="h-4 w-4 mr-2" />
                  {t('containers:deleteContainer')}
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Error display */}
        {error && (
          <div className="mb-6 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
            <p className="text-red-700 dark:text-red-400">{error}</p>
          </div>
        )}

        {/* Container Details */}
        <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 mb-8">
          <div className="p-6">
            <div className="flex items-start justify-between">
              <div className="flex items-start space-x-4">
                <div className="flex-shrink-0">
                  <CubeIcon className="h-12 w-12 text-blue-500 dark:text-blue-400" />
                </div>
                <div className="min-w-0 flex-1">
                  <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
                    {container.name}
                  </h1>
                  {container.description && (
                    <p className="mt-2 text-gray-600 dark:text-gray-400">{container.description}</p>
                  )}
                </div>
              </div>
              <StatusBadge status={status} size="md" />
            </div>

            <div className="mt-6 grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="flex items-center space-x-2 text-sm text-gray-600 dark:text-gray-400">
                <UserIcon className="h-4 w-4" />
                <span>
                  {t('containers:author')}: {container.author || t('containers:unknown')}
                </span>
              </div>

              {container.repository && (
                <div className="flex items-center space-x-2 text-sm text-gray-600 dark:text-gray-400">
                  <LinkIcon className="h-4 w-4" />
                  <span className="truncate">
                    {t('containers:repository')}: {container.repository}
                  </span>
                </div>
              )}

              <div className="flex items-center space-x-2 text-sm text-gray-600 dark:text-gray-400">
                <ClockIcon className="h-4 w-4" />
                <span>
                  {t('containers:updated')}: {formatDate(container.updated_at)}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Versions Section */}
        <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
          <div className="p-6 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                  {t('containers:versions')}
                </h2>
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  {t('containers:totalVersions', {
                    count: versions.length,
                    published: publishedVersions.length,
                  })}
                </p>
              </div>

              {isDeveloper && status !== 'deleted' && (
                <button
                  onClick={handleCreateVersion}
                  className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
                >
                  <PlusIcon className="h-4 w-4 mr-2" />
                  {t('containers:createVersion')}
                </button>
              )}
            </div>
          </div>

          <div className="p-6">
            {versionsLoading ? (
              <div className="space-y-4">
                {Array.from({ length: 3 }).map((_, index) => (
                  <div
                    key={index}
                    className="animate-pulse border border-gray-200 dark:border-gray-700 rounded-lg p-4"
                  >
                    <div className="h-4 bg-gray-300 dark:bg-gray-600 rounded w-1/4 mb-2"></div>
                    <div className="h-3 bg-gray-300 dark:bg-gray-600 rounded w-3/4"></div>
                  </div>
                ))}
              </div>
            ) : versions.length > 0 ? (
              <div className="space-y-4">
                {versions.map((version) => (
                  <div
                    key={version.id}
                    className="border border-gray-200 dark:border-gray-700 rounded-lg p-4 hover:border-gray-300 dark:hover:border-gray-600 transition-colors"
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-3">
                          <h3 className="text-lg font-medium text-gray-900 dark:text-white">
                            {version.version}
                          </h3>
                          <StatusBadge status={getContainerVersionStatus(version)} size="sm" />
                        </div>
                        <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
                          {t('containers:created')} {formatDate(version.created_at)}
                          {version.published_at && (
                            <span>
                              {' '}
                              • {t('containers:published')} {formatDate(version.published_at)}
                            </span>
                          )}
                        </p>
                      </div>

                      <div className="flex items-center space-x-2">
                        <Link
                          to={`/containers/${containerId}/versions/${version.version}`}
                          className="inline-flex items-center px-3 py-1 border border-gray-300 dark:border-gray-600 rounded-md text-xs font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700"
                        >
                          <EyeIcon className="h-3 w-3 mr-1" />
                          {t('containers:view')}
                        </Link>

                        {isDeveloper && !version.published && (
                          <button
                            onClick={() => handlePublishVersion(version)}
                            className="inline-flex items-center px-3 py-1 border border-transparent rounded-md text-xs font-medium text-white bg-blue-600 hover:bg-blue-700"
                          >
                            {t('containers:publish')}
                          </button>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8">
                <p className="text-gray-500 dark:text-gray-400">{t('containers:noVersionsYet')}</p>
                {isDeveloper && status !== 'deleted' && (
                  <button
                    onClick={handleCreateVersion}
                    className="mt-4 inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
                  >
                    <PlusIcon className="h-4 w-4 mr-2" />
                    {t('containers:createFirstVersion')}
                  </button>
                )}
              </div>
            )}

            {versionsError && (
              <div className="text-center py-4">
                <p className="text-red-600 dark:text-red-400">{versionsError}</p>
                <button
                  onClick={refetchVersions}
                  className="mt-2 text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300"
                >
                  {t('containers:tryAgain')}
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Confirmation Modal */}
      <ConfirmationModal
        isOpen={confirmationModal.isOpen}
        onClose={confirmationModal.closeModal}
        onConfirm={confirmationModal.handleConfirm}
        title={confirmationModal.title}
        message={confirmationModal.message}
        confirmLabel={confirmationModal.confirmLabel}
        cancelLabel={confirmationModal.cancelLabel}
        variant={confirmationModal.variant}
        isLoading={confirmationModal.isLoading}
      />
    </div>
  );
};

export default ContainerDetailPage;
