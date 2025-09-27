import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import {
  EyeIcon,
  PencilIcon,
  TrashIcon,
  CubeIcon,
  ClockIcon,
  UserIcon,
  LinkIcon,
} from '@heroicons/react/24/outline';
import { Module } from '../../types/module';
import { StatusBadge, getModuleStatus } from '../common/StatusBadge';
import { useAuth } from '../../hooks/useAuth';

interface ModuleCardProps {
  module: Module;
  onEdit?: (module: Module) => void;
  onDelete?: (module: Module) => void;
  className?: string;
  showActions?: boolean;
}

export const ModuleCard: React.FC<ModuleCardProps> = ({
  module,
  onEdit,
  onDelete,
  className = '',
  showActions = true,
}) => {
  const { isDeveloper } = useAuth();
  const [isDeleting, setIsDeleting] = useState(false);

  const status = getModuleStatus(module);
  const versionCount = module.versions?.length || 0;
  const publishedVersions = module.versions?.filter((v) => v.published).length || 0;

  const handleEdit = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (onEdit) {
      onEdit(module);
    }
  };

  const handleDelete = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();

    if (!onDelete || isDeleting) return;

    const confirmed = window.confirm(
      `Are you sure you want to delete module "${module.name}"?\n\nThis action cannot be undone.`
    );

    if (confirmed) {
      setIsDeleting(true);
      try {
        await onDelete(module);
      } catch (error) {
        console.error('Failed to delete module:', error);
      } finally {
        setIsDeleting(false);
      }
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const cardClasses = `
    group relative bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700
    hover:border-gray-300 dark:hover:border-gray-600 hover:shadow-md transition-all duration-200
    ${status === 'deleted' ? 'opacity-60' : ''}
    ${className}
  `.trim();

  return (
    <Link to={`/modules/${module.id}`} className="block">
      <div className={cardClasses}>
        {/* Header */}
        <div className="p-6 pb-4">
          <div className="flex items-start justify-between">
            <div className="flex items-start space-x-3 min-w-0 flex-1">
              <div className="flex-shrink-0">
                <CubeIcon className="h-8 w-8 text-blue-500 dark:text-blue-400" />
              </div>
              <div className="min-w-0 flex-1">
                <h3 className="text-lg font-semibold text-gray-900 dark:text-white truncate group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors">
                  {module.name}
                </h3>
                {module.description && (
                  <p className="mt-1 text-sm text-gray-600 dark:text-gray-400 line-clamp-2">
                    {module.description}
                  </p>
                )}
              </div>
            </div>

            <div className="flex items-center space-x-2 ml-4">
              <StatusBadge status={status} size="sm" />
              {showActions && isDeveloper && status !== 'deleted' && (
                <div className="opacity-0 group-hover:opacity-100 transition-opacity flex items-center space-x-1">
                  <button
                    onClick={handleEdit}
                    className="p-1 text-gray-400 hover:text-blue-500 dark:hover:text-blue-400 transition-colors"
                    title="Edit module"
                  >
                    <PencilIcon className="h-4 w-4" />
                  </button>
                  <button
                    onClick={handleDelete}
                    disabled={isDeleting}
                    className="p-1 text-gray-400 hover:text-red-500 dark:hover:text-red-400 transition-colors disabled:opacity-50"
                    title="Delete module"
                  >
                    <TrashIcon className="h-4 w-4" />
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Content */}
        <div className="px-6 pb-4">
          <div className="flex items-center space-x-4 text-sm text-gray-500 dark:text-gray-400">
            {module.author && (
              <div className="flex items-center space-x-1">
                <UserIcon className="h-4 w-4" />
                <span>{module.author}</span>
              </div>
            )}

            {module.repository && (
              <div className="flex items-center space-x-1">
                <LinkIcon className="h-4 w-4" />
                <span className="truncate max-w-32">{module.repository}</span>
              </div>
            )}
          </div>

          {/* Version info */}
          <div className="mt-3 flex items-center justify-between">
            <div className="flex items-center space-x-4 text-sm">
              <div className="text-gray-600 dark:text-gray-400">
                <span className="font-medium">{versionCount}</span> version
                {versionCount !== 1 ? 's' : ''}
              </div>
              {publishedVersions > 0 && (
                <div className="text-blue-600 dark:text-blue-400">
                  <span className="font-medium">{publishedVersions}</span> published
                </div>
              )}
            </div>

            <div className="flex items-center space-x-1 text-xs text-gray-500 dark:text-gray-400">
              <ClockIcon className="h-3 w-3" />
              <span>Updated {formatDate(module.updated_at)}</span>
            </div>
          </div>
        </div>

        {/* Footer actions (visible on mobile) */}
        {showActions && (
          <div className="px-6 py-3 bg-gray-50 dark:bg-gray-700/50 rounded-b-lg border-t border-gray-200 dark:border-gray-600 sm:hidden">
            <div className="flex items-center justify-between">
              <Link
                to={`/modules/${module.id}`}
                className="inline-flex items-center space-x-1 text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300"
              >
                <EyeIcon className="h-4 w-4" />
                <span>View Details</span>
              </Link>

              {isDeveloper && status !== 'deleted' && (
                <div className="flex items-center space-x-3">
                  <button
                    onClick={handleEdit}
                    className="inline-flex items-center space-x-1 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300"
                  >
                    <PencilIcon className="h-4 w-4" />
                    <span>Edit</span>
                  </button>
                  <button
                    onClick={handleDelete}
                    disabled={isDeleting}
                    className="inline-flex items-center space-x-1 text-sm text-red-600 dark:text-red-400 hover:text-red-700 dark:hover:text-red-300 disabled:opacity-50"
                  >
                    <TrashIcon className="h-4 w-4" />
                    <span>{isDeleting ? 'Deleting...' : 'Delete'}</span>
                  </button>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Loading overlay */}
        {isDeleting && (
          <div className="absolute inset-0 bg-white dark:bg-gray-800 bg-opacity-75 dark:bg-opacity-75 flex items-center justify-center rounded-lg">
            <div className="flex items-center space-x-2 text-gray-600 dark:text-gray-400">
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-current"></div>
              <span className="text-sm">Deleting...</span>
            </div>
          </div>
        )}
      </div>
    </Link>
  );
};

export default ModuleCard;
