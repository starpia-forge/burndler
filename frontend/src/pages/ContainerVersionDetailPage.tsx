import React, { useEffect, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  ArrowLeftIcon,
  PencilIcon,
  CubeIcon,
  ClockIcon,
  CodeBracketIcon,
  Cog6ToothIcon,
  DocumentIcon,
  LinkIcon,
} from '@heroicons/react/24/outline';
import { Container, ContainerVersion } from '../types/container';
import {
  StatusBadge,
  getContainerVersionStatus,
} from '../components/common/StatusBadge';
import { useAuth } from '../hooks/useAuth';
import containerService from '../services/containerService';
import { useContainerVersions } from '../hooks/useContainerVersions';

const ContainerVersionDetailPage: React.FC = () => {
  const { id, version } = useParams<{ id: string; version: string }>();
  const navigate = useNavigate();
  const { canCreateContainer } = useAuth();
  const { t } = useTranslation(['containers', 'common']);

  const containerId = id ? parseInt(id, 10) : 0;

  const [container, setContainer] = useState<Container | null>(null);
  const [versionData, setVersionData] = useState<ContainerVersion | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const { publishVersion } = useContainerVersions({
    containerId,
    autoFetch: false,
  });

  // Fetch container and version data
  useEffect(() => {
    const fetchData = async () => {
      if (!containerId || !version) return;

      try {
        setLoading(true);
        setError(null);

        // Fetch both container and version data in parallel
        const [containerData, versionResponse] = await Promise.all([
          containerService.getContainer(containerId, false),
          containerService.getVersion(containerId, version),
        ]);

        setContainer(containerData);
        setVersionData(versionResponse);
      } catch (error: any) {
        if (error.status === 404) {
          setError(t('containers:versionNotFound'));
        } else {
          setError(error.message || t('containers:failedToFetch'));
        }
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [containerId, version, t]);

  const handlePublishVersion = async () => {
    if (!versionData) return;

    try {
      await publishVersion(versionData.version);
      // Refresh version data
      const updatedVersion = await containerService.getVersion(containerId, version!);
      setVersionData(updatedVersion);
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

  const formatJson = (obj: any) => {
    return JSON.stringify(obj, null, 2);
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

  if (error) {
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

  if (!container || !versionData) return null;

  const status = getContainerVersionStatus(versionData);

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <Link
                to={`/containers/${containerId}`}
                className="inline-flex items-center text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300"
              >
                <ArrowLeftIcon className="h-4 w-4 mr-1" />
                {t('containers:backToContainer')}
              </Link>
            </div>

            {canCreateContainer && (
              <div className="flex items-center space-x-3">
                {!versionData.published && (
                  <button
                    onClick={handlePublishVersion}
                    className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
                  >
                    {t('containers:publishVersion')}
                  </button>
                )}
              </div>
            )}
          </div>
        </div>

        {/* Version Header */}
        <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 mb-8">
          <div className="p-6">
            <div className="flex items-start justify-between">
              <div className="flex items-start space-x-4">
                <div className="flex-shrink-0">
                  <CubeIcon className="h-12 w-12 text-blue-500 dark:text-blue-400" />
                </div>
                <div className="min-w-0 flex-1">
                  <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
                    {container.name} - {versionData.version}
                  </h1>
                  <p className="mt-1 text-gray-600 dark:text-gray-400">
                    {t('containers:versionDetails')}
                  </p>
                </div>
              </div>
              <StatusBadge status={status} size="md" />
            </div>

            <div className="mt-6 grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="flex items-center space-x-2 text-sm text-gray-600 dark:text-gray-400">
                <ClockIcon className="h-4 w-4" />
                <span>
                  {t('containers:created')}: {formatDate(versionData.created_at)}
                </span>
              </div>

              {versionData.published_at && (
                <div className="flex items-center space-x-2 text-sm text-gray-600 dark:text-gray-400">
                  <ClockIcon className="h-4 w-4" />
                  <span>
                    {t('containers:published')}: {formatDate(versionData.published_at)}
                  </span>
                </div>
              )}

              <div className="flex items-center space-x-2 text-sm text-gray-600 dark:text-gray-400">
                <ClockIcon className="h-4 w-4" />
                <span>
                  {t('containers:updated')}: {formatDate(versionData.updated_at)}
                </span>
              </div>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Content */}
          <div className="lg:col-span-2 space-y-8">
            {/* Docker Compose Content */}
            <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
              <div className="p-6 border-b border-gray-200 dark:border-gray-700">
                <div className="flex items-center space-x-2">
                  <CodeBracketIcon className="h-5 w-5 text-gray-400" />
                  <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                    {t('containers:dockerComposeContent')}
                  </h2>
                </div>
              </div>
              <div className="p-6">
                <pre className="bg-gray-50 dark:bg-gray-900 rounded-md p-4 overflow-x-auto text-sm font-mono text-gray-900 dark:text-gray-100 border border-gray-200 dark:border-gray-700">
                  <code>{versionData.compose_content}</code>
                </pre>
              </div>
            </div>

            {/* Variables */}
            {Object.keys(versionData.variables).length > 0 && (
              <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
                <div className="p-6 border-b border-gray-200 dark:border-gray-700">
                  <div className="flex items-center space-x-2">
                    <Cog6ToothIcon className="h-5 w-5 text-gray-400" />
                    <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
                      {t('containers:variables')}
                    </h2>
                  </div>
                </div>
                <div className="p-6">
                  <pre className="bg-gray-50 dark:bg-gray-900 rounded-md p-4 overflow-x-auto text-sm font-mono text-gray-900 dark:text-gray-100 border border-gray-200 dark:border-gray-700">
                    <code>{formatJson(versionData.variables)}</code>
                  </pre>
                </div>
              </div>
            )}
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Resource Paths */}
            {versionData.resource_paths.length > 0 && (
              <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
                <div className="p-6 border-b border-gray-200 dark:border-gray-700">
                  <div className="flex items-center space-x-2">
                    <DocumentIcon className="h-5 w-5 text-gray-400" />
                    <h3 className="text-md font-semibold text-gray-900 dark:text-white">
                      {t('containers:resourcePaths')}
                    </h3>
                  </div>
                </div>
                <div className="p-6">
                  <ul className="space-y-2">
                    {versionData.resource_paths.map((path, index) => (
                      <li
                        key={index}
                        className="text-sm text-gray-600 dark:text-gray-400 font-mono bg-gray-50 dark:bg-gray-900 px-3 py-2 rounded border border-gray-200 dark:border-gray-700"
                      >
                        {path}
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            )}

            {/* Dependencies */}
            {Object.keys(versionData.dependencies).length > 0 && (
              <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
                <div className="p-6 border-b border-gray-200 dark:border-gray-700">
                  <div className="flex items-center space-x-2">
                    <LinkIcon className="h-5 w-5 text-gray-400" />
                    <h3 className="text-md font-semibold text-gray-900 dark:text-white">
                      {t('containers:dependencies')}
                    </h3>
                  </div>
                </div>
                <div className="p-6">
                  <pre className="bg-gray-50 dark:bg-gray-900 rounded-md p-4 overflow-x-auto text-sm font-mono text-gray-900 dark:text-gray-100 border border-gray-200 dark:border-gray-700">
                    <code>{formatJson(versionData.dependencies)}</code>
                  </pre>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default ContainerVersionDetailPage;