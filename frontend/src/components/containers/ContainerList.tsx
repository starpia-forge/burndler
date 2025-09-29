import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { PlusIcon } from '@heroicons/react/24/outline';
import { Container, ContainerFilters as ContainerFiltersType } from '../../types/container';
import { useAuth } from '../../hooks/useAuth';
import ContainerCard from './ContainerCard';
import ContainerFilters from './ContainerFilters';
import Pagination from '../common/Pagination';

interface ContainerListProps {
  containers: Container[];
  loading: boolean;
  initialLoading?: boolean;
  isRefreshing?: boolean;
  error: string | null;
  pagination: any;
  filters: ContainerFiltersType;
  onFiltersChange: (filters: ContainerFiltersType) => void;
  onClearFilters: () => void;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
  onCreateContainer?: () => void;
  onEditContainer?: (container: Container) => void;
  onDeleteContainer?: (container: Container) => Promise<void>;
  onRefresh?: () => void;
  className?: string;
}

export const ContainerList: React.FC<ContainerListProps> = ({
  containers,
  loading,
  initialLoading = false,
  isRefreshing = false,
  error,
  pagination,
  filters,
  onFiltersChange,
  onClearFilters,
  onPageChange,
  onPageSizeChange,
  onCreateContainer,
  onEditContainer,
  onDeleteContainer,
  onRefresh,
  className = '',
}) => {
  const { isDeveloper } = useAuth();
  const { t } = useTranslation(['containers', 'common']);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  const handleDeleteContainer = async (container: Container) => {
    if (!onDeleteContainer) return;

    try {
      setDeleteError(null);
      await onDeleteContainer(container);
    } catch (error: any) {
      setDeleteError(error.message || t('containers:failedToDelete'));
      throw error; // Re-throw for component error handling
    }
  };

  // Loading state - only show skeleton on initial load
  if (initialLoading && containers.length === 0) {
    return (
      <div className={`space-y-6 ${className}`}>
        <ContainerFilters
          filters={filters}
          onFiltersChange={onFiltersChange}
          onClearFilters={onClearFilters}
          loading={loading}
        />

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {Array.from({ length: 6 }).map((_, index) => (
            <div
              key={index}
              className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6 animate-pulse"
            >
              <div className="flex items-start space-x-3">
                <div className="w-8 h-8 bg-gray-300 dark:bg-gray-600 rounded"></div>
                <div className="flex-1">
                  <div className="h-5 bg-gray-300 dark:bg-gray-600 rounded w-3/4 mb-2"></div>
                  <div className="h-4 bg-gray-300 dark:bg-gray-600 rounded w-full mb-1"></div>
                  <div className="h-4 bg-gray-300 dark:bg-gray-600 rounded w-2/3"></div>
                </div>
              </div>
              <div className="mt-4 flex items-center justify-between">
                <div className="h-4 bg-gray-300 dark:bg-gray-600 rounded w-24"></div>
                <div className="h-4 bg-gray-300 dark:bg-gray-600 rounded w-20"></div>
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
            {t('containers:title')}
          </h1>
          <p className="text-gray-600 dark:text-gray-400">{t('containers:description')}</p>
        </div>

        <div className="flex items-center space-x-3">
          {onRefresh && (
            <button
              onClick={onRefresh}
              disabled={loading || isRefreshing}
              className="inline-flex items-center px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
            >
              {isRefreshing ? t('containers:refreshing') : t('containers:refresh')}
            </button>
          )}

          {isDeveloper && onCreateContainer && (
            <button
              onClick={onCreateContainer}
              disabled={loading}
              className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
            >
              <PlusIcon className="h-4 w-4 mr-2" />
              {t('containers:createContainer')}
            </button>
          )}
        </div>
      </div>

      {/* Filters */}
      <ContainerFilters
        filters={filters}
        onFiltersChange={onFiltersChange}
        onClearFilters={onClearFilters}
        loading={loading}
      />

      {/* Error display */}
      {(error || deleteError) && (
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
          <div className="flex">
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800 dark:text-red-300">
                {t('containers:error')}
              </h3>
              <div className="mt-2 text-sm text-red-700 dark:text-red-400">
                {error || deleteError}
              </div>
              {deleteError && (
                <button
                  onClick={() => setDeleteError(null)}
                  className="mt-2 text-sm text-red-600 dark:text-red-400 hover:text-red-500 dark:hover:text-red-300"
                >
                  {t('containers:dismiss')}
                </button>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Container grid */}
      {containers.length > 0 ? (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {containers.map((container) => (
              <ContainerCard
                key={container.id}
                container={container}
                onEdit={onEditContainer}
                onDelete={handleDeleteContainer}
                className={isRefreshing ? 'opacity-75 pointer-events-none' : ''}
              />
            ))}
          </div>

          {/* Loading overlay for refresh operations */}
          {isRefreshing && (
            <div className="relative">
              <div className="absolute inset-0 bg-white dark:bg-gray-900 bg-opacity-50 dark:bg-opacity-50 flex items-center justify-center z-10">
                <div className="flex items-center space-x-2 text-gray-600 dark:text-gray-400">
                  <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-current"></div>
                  <span>{t('containers:refreshingContainers')}</span>
                </div>
              </div>
            </div>
          )}

          {/* Pagination */}
          {pagination && pagination.total > 0 && (
            <div className="flex justify-center">
              <Pagination
                pagination={pagination}
                onPageChange={onPageChange}
                onPageSizeChange={onPageSizeChange}
              />
            </div>
          )}
        </>
      ) : !loading ? (
        /* Empty state */
        <div className="text-center py-12">
          <svg
            className="mx-auto h-12 w-12 text-gray-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            aria-hidden="true"
          >
            <path
              vectorEffect="non-scaling-stroke"
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4"
            />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-white">
            {t('containers:noContainers')}
          </h3>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {filters.search || filters.author || filters.active !== undefined
              ? t('containers:tryAdjustingFilters')
              : t('containers:createYourFirst')}
          </p>
          <div className="mt-6">
            {filters.search || filters.author || filters.active !== undefined ? (
              <button
                onClick={onClearFilters}
                className="inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {t('containers:clearFilters')}
              </button>
            ) : isDeveloper && onCreateContainer ? (
              <button
                onClick={onCreateContainer}
                className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <PlusIcon className="h-4 w-4 mr-2" />
                {t('containers:createYourFirst')}
              </button>
            ) : null}
          </div>
        </div>
      ) : null}

      {/* Results summary */}
      {pagination && pagination.total > 0 && (
        <div className="text-center text-sm text-gray-500 dark:text-gray-400">
          {t('containers:showingResults', { showing: containers.length, total: pagination.total })}
        </div>
      )}
    </div>
  );
};

export default ContainerList;
