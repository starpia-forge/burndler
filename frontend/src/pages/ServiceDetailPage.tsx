import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { ArrowLeftIcon, ExclamationTriangleIcon } from '@heroicons/react/24/outline';
import { ServiceDetail, ServiceForm, ContainerSelector } from '../components/services';
import { Service, UpdateServiceRequest, ServiceContainerFormState } from '../types/service';
import { useAuth } from '../hooks/useAuth';
import serviceService from '../services/serviceService';

type EditMode = 'details' | 'containers';

export const ServiceDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { isDeveloper } = useAuth();

  const [service, setService] = useState<Service | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [editMode, setEditMode] = useState<EditMode | null>(null);
  const [isUpdating, setIsUpdating] = useState(false);
  const [selectedContainers, setSelectedContainers] = useState<ServiceContainerFormState[]>([]);

  const serviceId = id ? parseInt(id, 10) : null;

  // Load service data
  useEffect(() => {
    if (!serviceId) {
      setError('Invalid service ID');
      setLoading(false);
      return;
    }

    loadService();
  }, [serviceId]);

  const loadService = async () => {
    if (!serviceId) return;

    try {
      setLoading(true);
      setError(null);
      const serviceData = await serviceService.getService(serviceId, true); // Include containers
      setService(serviceData);

      // Convert service containers to form state for editing
      const containerFormData: ServiceContainerFormState[] =
        serviceData.containers?.map((sc) => ({
          container_id: sc.container_id,
          container_version: sc.container_version,
          variables: sc.variables || {},
          order: sc.order,
        })) || [];
      setSelectedContainers(containerFormData);
    } catch (error: any) {
      console.error('Failed to load service:', error);
      if (error.status === 404) {
        setError('Service not found');
      } else {
        setError(error.message || 'Failed to load service');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateService = async (data: UpdateServiceRequest) => {
    if (!serviceId || !service) return;

    setIsUpdating(true);
    try {
      const updatedService = await serviceService.updateService(serviceId, data);
      setService(updatedService);
      setEditMode(null);
    } catch (error) {
      console.error('Failed to update service:', error);
      throw error;
    } finally {
      setIsUpdating(false);
    }
  };

  const handleDeleteService = async () => {
    if (!serviceId) return;

    try {
      await serviceService.deleteService(serviceId);
      navigate('/services');
    } catch (error) {
      console.error('Failed to delete service:', error);
      throw error;
    }
  };

  const handleAddContainer = () => {
    setEditMode('containers');
  };

  const handleEditContainer = (_containerId: number) => {
    // In a real implementation, this would open a specific container edit modal
    setEditMode('containers');
  };

  const handleRemoveContainer = async (containerId: number) => {
    if (!serviceId) return;

    try {
      await serviceService.removeContainerFromService(serviceId, containerId);
      await loadService(); // Reload to get updated data
    } catch (error) {
      console.error('Failed to remove container:', error);
      throw error;
    }
  };

  const handleContainerSelectionChange = (containers: ServiceContainerFormState[]) => {
    setSelectedContainers(containers);
  };

  const handleSaveContainers = async () => {
    if (!serviceId || !service) return;

    setIsUpdating(true);
    try {
      // For simplicity, we'll need to implement bulk container updates
      // In a real implementation, this would involve adding/updating/removing containers
      // based on the differences between current and selected containers

      // For now, just reload the service
      await loadService();
      setEditMode(null);
    } catch (error) {
      console.error('Failed to save containers:', error);
      throw error;
    } finally {
      setIsUpdating(false);
    }
  };

  const handleValidateService = async () => {
    if (!serviceId) return;

    try {
      const result = await serviceService.validateService(serviceId);
      console.log('Validation result:', result);
      // Show validation results to user
    } catch (error) {
      console.error('Failed to validate service:', error);
    }
  };

  const handleBuildService = async () => {
    if (!serviceId) return;

    try {
      const result = await serviceService.buildService(serviceId);
      console.log('Build result:', result);
      // Show build results to user or redirect to build page
    } catch (error) {
      console.error('Failed to build service:', error);
    }
  };

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center space-x-3">
          <div className="h-6 w-6 bg-muted rounded animate-pulse" />
          <div className="h-6 bg-muted rounded w-32 animate-pulse" />
        </div>
        <div className="h-64 bg-muted rounded animate-pulse" />
        <div className="h-96 bg-muted rounded animate-pulse" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div className="flex items-center space-x-3">
          <Link
            to="/services"
            className="inline-flex items-center text-sm text-muted-foreground hover:text-foreground"
          >
            <ArrowLeftIcon className="h-4 w-4 mr-1" />
            Back to Services
          </Link>
        </div>

        <div className="bg-red-50 border border-red-200 rounded-lg p-6">
          <div className="flex">
            <ExclamationTriangleIcon className="h-5 w-5 text-red-400" />
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">Error</h3>
              <div className="mt-2 text-sm text-red-700">
                <p>{error}</p>
              </div>
              <div className="mt-4">
                <button
                  onClick={loadService}
                  className="text-sm bg-red-100 text-red-800 rounded-md px-3 py-1 hover:bg-red-200"
                >
                  Try again
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!service) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">Service not found</p>
        <Link
          to="/services"
          className="mt-4 inline-flex items-center text-blue-600 hover:text-blue-700"
        >
          <ArrowLeftIcon className="h-4 w-4 mr-1" />
          Back to Services
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <div className="flex items-center space-x-3">
        <Link
          to="/services"
          className="inline-flex items-center text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeftIcon className="h-4 w-4 mr-1" />
          Back to Services
        </Link>
        <span className="text-muted-foreground">/</span>
        <span className="text-sm font-medium text-foreground">{service.name}</span>
      </div>

      {/* Edit Service Details Modal */}
      {editMode === 'details' && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-background rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
            <ServiceForm
              service={service}
              onSubmit={handleUpdateService}
              onCancel={() => setEditMode(null)}
              loading={isUpdating}
              title="Edit Service"
              submitLabel="Update Service"
            />
          </div>
        </div>
      )}

      {/* Edit Containers Modal */}
      {editMode === 'containers' && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-background rounded-lg shadow-xl max-w-4xl w-full mx-4 max-h-[90vh] overflow-y-auto">
            <div className="p-6">
              <div className="flex items-center justify-between mb-6">
                <h2 className="text-xl font-semibold text-foreground">Edit Service Containers</h2>
                <button
                  onClick={() => setEditMode(null)}
                  className="text-muted-foreground hover:text-foreground"
                >
                  ✕
                </button>
              </div>

              <ContainerSelector
                onSelectionChange={handleContainerSelectionChange}
                initialSelection={selectedContainers}
                disabled={!isDeveloper}
              />

              <div className="flex items-center justify-end space-x-3 mt-6 pt-4 border-t border-border">
                <button
                  onClick={() => setEditMode(null)}
                  disabled={isUpdating}
                  className="px-4 py-2 border border-border rounded-md shadow-sm text-sm font-medium text-foreground bg-background hover:bg-muted focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
                >
                  Cancel
                </button>
                <button
                  onClick={handleSaveContainers}
                  disabled={isUpdating || !isDeveloper}
                  className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isUpdating ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Service Detail */}
      <ServiceDetail
        service={service}
        loading={isUpdating}
        onEdit={() => setEditMode('details')}
        onDelete={handleDeleteService}
        onAddContainer={handleAddContainer}
        onEditContainer={handleEditContainer}
        onRemoveContainer={handleRemoveContainer}
        onValidate={handleValidateService}
        onBuild={handleBuildService}
      />
    </div>
  );
};

export default ServiceDetailPage;
