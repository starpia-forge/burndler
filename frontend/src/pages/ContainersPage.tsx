import React, { useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { Container } from '../types/container';
import { useContainers } from '../hooks/useContainers';
import ContainerList from '../components/containers/ContainerList';

const ContainersPage: React.FC = () => {
  const navigate = useNavigate();

  const {
    containers,
    loading,
    initialLoading,
    isRefreshing,
    error,
    pagination,
    filters,
    refetch,
    setFilters,
    updateFilter,
    clearFilters,
    deleteContainer,
  } = useContainers({
    autoFetch: true,
    initialFilters: {
      page: 1,
      page_size: 12, // Show more containers on the main page
      active: undefined,
      author: '',
      show_deleted: false,
      published_only: false,
      search: '',
    },
  });

  // Navigation handlers
  const handleCreateContainer = useCallback(() => {
    navigate('/containers/create');
  }, [navigate]);

  const handleEditContainer = useCallback(
    (container: Container) => {
      navigate(`/containers/${container.id}/edit`);
    },
    [navigate]
  );

  // Filter handlers
  const handleFiltersChange = useCallback(
    (newFilters: any) => {
      setFilters(newFilters);
    },
    [setFilters]
  );

  const handleClearFilters = useCallback(() => {
    clearFilters();
  }, [clearFilters]);

  // Pagination handlers
  const handlePageChange = useCallback(
    (page: number) => {
      updateFilter('page', page);
    },
    [updateFilter]
  );

  const handlePageSizeChange = useCallback(
    (pageSize: number) => {
      setFilters({ ...filters, page_size: pageSize, page: 1 });
    },
    [filters, setFilters]
  );

  // Delete handler
  const handleDeleteContainer = useCallback(
    async (container: Container) => {
      await deleteContainer(container.id);
    },
    [deleteContainer]
  );

  // Refresh handler
  const handleRefresh = useCallback(() => {
    refetch();
  }, [refetch]);

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <ContainerList
          containers={containers}
          loading={loading}
          initialLoading={initialLoading}
          isRefreshing={isRefreshing}
          error={error}
          pagination={pagination}
          filters={filters}
          onFiltersChange={handleFiltersChange}
          onClearFilters={handleClearFilters}
          onPageChange={handlePageChange}
          onPageSizeChange={handlePageSizeChange}
          onCreateContainer={handleCreateContainer}
          onEditContainer={handleEditContainer}
          onDeleteContainer={handleDeleteContainer}
          onRefresh={handleRefresh}
        />
      </div>
    </div>
  );
};

export default ContainersPage;
