import React, { useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { Module } from '../types/module';
import { useModules } from '../hooks/useModules';
import ModuleList from '../components/modules/ModuleList';

const ModulesPage: React.FC = () => {
  const navigate = useNavigate();

  const {
    modules,
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
    deleteModule,
  } = useModules({
    autoFetch: true,
    initialFilters: {
      page: 1,
      page_size: 12, // Show more modules on the main page
      active: undefined,
      author: '',
      show_deleted: false,
      published_only: false,
      search: '',
    },
  });

  // Navigation handlers
  const handleCreateModule = useCallback(() => {
    navigate('/modules/create');
  }, [navigate]);

  const handleEditModule = useCallback(
    (module: Module) => {
      navigate(`/modules/${module.id}/edit`);
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
      setFilters((prev) => ({ ...prev, page_size: pageSize, page: 1 }));
    },
    [setFilters]
  );

  // Delete handler
  const handleDeleteModule = useCallback(
    async (module: Module) => {
      await deleteModule(module.id);
    },
    [deleteModule]
  );

  // Refresh handler
  const handleRefresh = useCallback(() => {
    refetch();
  }, [refetch]);

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <ModuleList
          modules={modules}
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
          onCreateModule={handleCreateModule}
          onEditModule={handleEditModule}
          onDeleteModule={handleDeleteModule}
          onRefresh={handleRefresh}
        />
      </div>
    </div>
  );
};

export default ModulesPage;
