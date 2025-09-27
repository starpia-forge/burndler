import React, { useEffect, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
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
import { Module, ModuleVersion } from '../types/module';
import { StatusBadge, getModuleStatus, getVersionStatus } from '../components/common/StatusBadge';
import { useAuth } from '../hooks/useAuth';
import moduleService from '../services/moduleService';
import { useModuleVersions } from '../hooks/useModuleVersions';

const ModuleDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { isDeveloper } = useAuth();

  const [module, setModule] = useState<Module | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const moduleId = id ? parseInt(id, 10) : 0;

  const {
    versions,
    loading: versionsLoading,
    error: versionsError,
    refetch: refetchVersions,
    publishVersion,
  } = useModuleVersions({
    moduleId,
    autoFetch: true,
  });

  // Fetch module details
  useEffect(() => {
    const fetchModule = async () => {
      if (!moduleId) return;

      try {
        setLoading(true);
        setError(null);
        const moduleData = await moduleService.getModule(moduleId, false);
        setModule(moduleData);
      } catch (error: any) {
        setError(error.message || 'Failed to fetch module');
      } finally {
        setLoading(false);
      }
    };

    fetchModule();
  }, [moduleId]);

  const handleEdit = () => {
    navigate(`/modules/${moduleId}/edit`);
  };

  const handleDelete = async () => {
    if (!module) return;

    const confirmed = window.confirm(
      `Are you sure you want to delete module "${module.name}"?\n\nThis action cannot be undone.`
    );

    if (confirmed) {
      try {
        await moduleService.deleteModule(moduleId);
        navigate('/modules');
      } catch (error: any) {
        setError(error.message || 'Failed to delete module');
      }
    }
  };

  const handleCreateVersion = () => {
    navigate(`/modules/${moduleId}/versions/create`);
  };

  const handlePublishVersion = async (version: ModuleVersion) => {
    try {
      await publishVersion(version.version);
    } catch (error: any) {
      setError(error.message || 'Failed to publish version');
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

  if (error && !module) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
            <h3 className="text-lg font-medium text-red-800 dark:text-red-300 mb-2">
              Error loading module
            </h3>
            <p className="text-red-700 dark:text-red-400">{error}</p>
            <div className="mt-4">
              <Link
                to="/modules"
                className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-red-600 hover:bg-red-700"
              >
                <ArrowLeftIcon className="h-4 w-4 mr-2" />
                Back to Modules
              </Link>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!module) return null;

  const status = getModuleStatus(module);
  const publishedVersions = versions.filter((v) => v.published);

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <Link
                to="/modules"
                className="inline-flex items-center text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300"
              >
                <ArrowLeftIcon className="h-4 w-4 mr-1" />
                Back to Modules
              </Link>
            </div>

            {isDeveloper && status !== 'deleted' && (
              <div className="flex items-center space-x-3">
                <button
                  onClick={handleEdit}
                  className="inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700"
                >
                  <PencilIcon className="h-4 w-4 mr-2" />
                  Edit Module
                </button>
                <button
                  onClick={handleDelete}
                  className="inline-flex items-center px-4 py-2 border border-red-300 dark:border-red-600 rounded-md shadow-sm text-sm font-medium text-red-700 dark:text-red-300 bg-white dark:bg-gray-800 hover:bg-red-50 dark:hover:bg-red-900/20"
                >
                  <TrashIcon className="h-4 w-4 mr-2" />
                  Delete Module
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

        {/* Module Details */}
        <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 mb-8">
          <div className="p-6">
            <div className="flex items-start justify-between">
              <div className="flex items-start space-x-4">
                <div className="flex-shrink-0">
                  <CubeIcon className="h-12 w-12 text-blue-500 dark:text-blue-400" />
                </div>
                <div className="min-w-0 flex-1">
                  <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
                    {module.name}
                  </h1>
                  {module.description && (
                    <p className="mt-2 text-gray-600 dark:text-gray-400">{module.description}</p>
                  )}
                </div>
              </div>
              <StatusBadge status={status} size="md" />
            </div>

            <div className="mt-6 grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="flex items-center space-x-2 text-sm text-gray-600 dark:text-gray-400">
                <UserIcon className="h-4 w-4" />
                <span>Author: {module.author || 'Unknown'}</span>
              </div>

              {module.repository && (
                <div className="flex items-center space-x-2 text-sm text-gray-600 dark:text-gray-400">
                  <LinkIcon className="h-4 w-4" />
                  <span className="truncate">Repository: {module.repository}</span>
                </div>
              )}

              <div className="flex items-center space-x-2 text-sm text-gray-600 dark:text-gray-400">
                <ClockIcon className="h-4 w-4" />
                <span>Updated: {formatDate(module.updated_at)}</span>
              </div>
            </div>
          </div>
        </div>

        {/* Versions Section */}
        <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
          <div className="p-6 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-lg font-semibold text-gray-900 dark:text-white">Versions</h2>
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  {versions.length} total versions ({publishedVersions.length} published)
                </p>
              </div>

              {isDeveloper && status !== 'deleted' && (
                <button
                  onClick={handleCreateVersion}
                  className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
                >
                  <PlusIcon className="h-4 w-4 mr-2" />
                  Create Version
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
                          <StatusBadge status={getVersionStatus(version)} size="sm" />
                        </div>
                        <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">
                          Created {formatDate(version.created_at)}
                          {version.published_at && (
                            <span> • Published {formatDate(version.published_at)}</span>
                          )}
                        </p>
                      </div>

                      <div className="flex items-center space-x-2">
                        <Link
                          to={`/modules/${moduleId}/versions/${version.version}`}
                          className="inline-flex items-center px-3 py-1 border border-gray-300 dark:border-gray-600 rounded-md text-xs font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700"
                        >
                          <EyeIcon className="h-3 w-3 mr-1" />
                          View
                        </Link>

                        {isDeveloper && !version.published && (
                          <button
                            onClick={() => handlePublishVersion(version)}
                            className="inline-flex items-center px-3 py-1 border border-transparent rounded-md text-xs font-medium text-white bg-blue-600 hover:bg-blue-700"
                          >
                            Publish
                          </button>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center py-8">
                <p className="text-gray-500 dark:text-gray-400">No versions created yet.</p>
                {isDeveloper && status !== 'deleted' && (
                  <button
                    onClick={handleCreateVersion}
                    className="mt-4 inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700"
                  >
                    <PlusIcon className="h-4 w-4 mr-2" />
                    Create First Version
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
                  Try again
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default ModuleDetailPage;
