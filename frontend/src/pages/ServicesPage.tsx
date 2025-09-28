import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  PlusIcon,
  MagnifyingGlassIcon,
  ArrowPathIcon,
} from '@heroicons/react/24/outline';
import { ServiceList, ServiceForm } from '../components/services';
import { Service, CreateServiceRequest, UpdateServiceRequest } from '../types/service';
import { useServices } from '../hooks/useServices';
import { useAuth } from '../hooks/useAuth';
import serviceService from '../services/serviceService';

export const ServicesPage: React.FC = () => {
  const navigate = useNavigate();
  const { isDeveloper } = useAuth();
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [isCreating, setIsCreating] = useState(false);

  const {
    services,
    initialLoading,
    isRefreshing,
    error,
    pagination,
    filters,
    updateFilter,
    clearFilters,
    refetch,
    deleteService,
  } = useServices();

  const handleCreateService = async (data: CreateServiceRequest | UpdateServiceRequest) => {
    setIsCreating(true);
    try {
      const newService = await serviceService.createService(data as CreateServiceRequest);
      setShowCreateForm(false);
      // Navigate to the new service's detail page
      navigate(`/services/${newService.id}`);
    } catch (error) {
      console.error('Failed to create service:', error);
      throw error; // Let the form handle the error
    } finally {
      setIsCreating(false);
    }
  };

  const handleEditService = (service: Service) => {
    navigate(`/services/${service.id}/edit`);
  };

  const handleDeleteService = async (service: Service) => {
    try {
      await deleteService(service.id);
    } catch (error) {
      console.error('Failed to delete service:', error);
      throw error;
    }
  };

  const handleSearchChange = (value: string) => {
    updateFilter('search', value);
  };

  const handleActiveFilterChange = (value: string) => {
    if (value === 'all') {
      updateFilter('active', undefined);
    } else {
      updateFilter('active', value === 'true');
    }
  };

  const handlePageChange = (page: number) => {
    updateFilter('page', page);
  };

  const handleRefresh = () => {
    refetch();
  };

  const hasActiveFilters = filters.search || filters.active !== undefined;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Services</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Manage containerized services and their configurations
          </p>
        </div>

        <div className="flex items-center space-x-3">
          <button
            onClick={handleRefresh}
            disabled={isRefreshing}
            className="inline-flex items-center px-3 py-2 border border-border rounded-md shadow-sm text-sm font-medium text-foreground bg-background hover:bg-muted focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
          >
            <ArrowPathIcon className={`h-4 w-4 mr-2 ${isRefreshing ? 'animate-spin' : ''}`} />
            Refresh
          </button>

          {isDeveloper && (
            <button
              onClick={() => setShowCreateForm(true)}
              className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              <PlusIcon className="h-4 w-4 mr-2" />
              Create Service
            </button>
          )}
        </div>
      </div>

      {/* Filters */}
      <div className="bg-card rounded-lg border border-border p-4">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-sm font-medium text-foreground">Filters</h3>
          {hasActiveFilters && (
            <button
              onClick={clearFilters}
              className="text-xs text-blue-600 hover:text-blue-700"
            >
              Clear all
            </button>
          )}
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {/* Search */}
          <div>
            <label className="block text-xs font-medium text-muted-foreground mb-1">
              Search
            </label>
            <div className="relative">
              <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <input
                type="text"
                value={filters.search || ''}
                onChange={(e) => handleSearchChange(e.target.value)}
                placeholder="Search services..."
                className="w-full pl-10 pr-3 py-2 border border-border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-background text-foreground"
              />
            </div>
          </div>

          {/* Status Filter */}
          <div>
            <label className="block text-xs font-medium text-muted-foreground mb-1">
              Status
            </label>
            <select
              value={filters.active === undefined ? 'all' : filters.active.toString()}
              onChange={(e) => handleActiveFilterChange(e.target.value)}
              className="w-full px-3 py-2 border border-border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-background text-foreground"
            >
              <option value="all">All Services</option>
              <option value="true">Active</option>
              <option value="false">Inactive</option>
            </select>
          </div>

          {/* Page Size */}
          <div>
            <label className="block text-xs font-medium text-muted-foreground mb-1">
              Items per page
            </label>
            <select
              value={filters.page_size || 10}
              onChange={(e) => updateFilter('page_size', parseInt(e.target.value))}
              className="w-full px-3 py-2 border border-border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-background text-foreground"
            >
              <option value={5}>5</option>
              <option value={10}>10</option>
              <option value={20}>20</option>
              <option value={50}>50</option>
            </select>
          </div>
        </div>
      </div>

      {/* Error State */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="flex">
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">Error</h3>
              <div className="mt-2 text-sm text-red-700">
                <p>{error}</p>
              </div>
              <div className="mt-4">
                <button
                  onClick={handleRefresh}
                  className="text-sm bg-red-100 text-red-800 rounded-md px-3 py-1 hover:bg-red-200"
                >
                  Try again
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Services List */}
      <ServiceList
        services={services}
        loading={initialLoading}
        onEdit={handleEditService}
        onCreate={() => setShowCreateForm(true)}
        onDelete={handleDeleteService}
      />

      {/* Pagination */}
      {pagination && pagination.total_pages > 1 && (
        <div className="flex items-center justify-between border-t border-border pt-4">
          <div className="text-sm text-muted-foreground">
            Showing {((pagination.page - 1) * pagination.page_size) + 1} to{' '}
            {Math.min(pagination.page * pagination.page_size, pagination.total)} of{' '}
            {pagination.total} results
          </div>
          <div className="flex items-center space-x-2">
            <button
              onClick={() => handlePageChange(pagination.page - 1)}
              disabled={pagination.page <= 1}
              className="px-3 py-1 border border-border rounded-md text-sm text-foreground bg-background hover:bg-muted disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Previous
            </button>

            <div className="flex items-center space-x-1">
              {Array.from({ length: Math.min(5, pagination.total_pages) }, (_, i) => {
                const page = i + 1;
                const isCurrentPage = page === pagination.page;
                return (
                  <button
                    key={page}
                    onClick={() => handlePageChange(page)}
                    className={`px-3 py-1 border rounded-md text-sm ${
                      isCurrentPage
                        ? 'border-blue-500 bg-blue-50 text-blue-600'
                        : 'border-border text-foreground bg-background hover:bg-muted'
                    }`}
                  >
                    {page}
                  </button>
                );
              })}
            </div>

            <button
              onClick={() => handlePageChange(pagination.page + 1)}
              disabled={pagination.page >= pagination.total_pages}
              className="px-3 py-1 border border-border rounded-md text-sm text-foreground bg-background hover:bg-muted disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Next
            </button>
          </div>
        </div>
      )}

      {/* Create Service Modal */}
      {showCreateForm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-background rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
            <ServiceForm
              onSubmit={handleCreateService}
              onCancel={() => setShowCreateForm(false)}
              loading={isCreating}
              title="Create New Service"
              submitLabel="Create Service"
            />
          </div>
        </div>
      )}
    </div>
  );
};

export default ServicesPage;