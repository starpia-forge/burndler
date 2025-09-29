import React, { useState } from 'react';
import {
  PencilIcon,
  TrashIcon,
  PlusIcon,
  CubeIcon,
  ClockIcon,
  UserIcon,
} from '@heroicons/react/24/outline';
import { Service } from '../../types/service';
import { StatusBadge, getServiceStatus } from '../common/StatusBadge';
import { useAuth } from '../../hooks/useAuth';

interface ServiceDetailProps {
  service: Service;
  loading?: boolean;
  onEdit?: () => void;
  onDelete?: () => Promise<void>;
  onAddContainer?: () => void;
  onEditContainer?: (containerId: number) => void;
  onRemoveContainer?: (containerId: number) => Promise<void>;
  onValidate?: () => Promise<void>;
  onBuild?: () => Promise<void>;
}

export const ServiceDetail: React.FC<ServiceDetailProps> = ({
  service,
  loading = false,
  onEdit,
  onDelete,
  onAddContainer,
  onEditContainer,
  onRemoveContainer,
  onValidate,
  onBuild,
}) => {
  const { isDeveloper } = useAuth();
  const [isDeleting, setIsDeleting] = useState(false);
  const [removingContainerIds, setRemovingContainerIds] = useState<Set<number>>(new Set());

  const status = getServiceStatus(service);
  const containerCount = service.containers?.length || 0;

  const handleDelete = async () => {
    if (!onDelete || isDeleting) return;

    const confirmed = window.confirm(
      `Are you sure you want to delete service "${service.name}"?\n\nThis action cannot be undone.`
    );

    if (confirmed) {
      setIsDeleting(true);
      try {
        await onDelete();
      } catch (error) {
        console.error('Failed to delete service:', error);
      } finally {
        setIsDeleting(false);
      }
    }
  };

  const handleRemoveContainer = async (containerId: number) => {
    if (!onRemoveContainer) return;

    const confirmed = window.confirm(
      'Are you sure you want to remove this container from the service?'
    );

    if (confirmed) {
      setRemovingContainerIds((prev) => new Set(prev).add(containerId));
      try {
        await onRemoveContainer(containerId);
      } catch (error) {
        console.error('Failed to remove container:', error);
      } finally {
        setRemovingContainerIds((prev) => {
          const newSet = new Set(prev);
          newSet.delete(containerId);
          return newSet;
        });
      }
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (loading) {
    return (
      <div className="animate-pulse">
        <div className="h-8 bg-muted rounded w-1/3 mb-4"></div>
        <div className="h-32 bg-muted rounded mb-6"></div>
        <div className="h-64 bg-muted rounded"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="bg-card rounded-lg border border-border p-6">
        <div className="flex items-start justify-between">
          <div className="flex items-start space-x-4 min-w-0 flex-1">
            <div className="flex-shrink-0">
              <div className="h-12 w-12 bg-blue-100 rounded-lg flex items-center justify-center">
                <CubeIcon className="h-6 w-6 text-blue-600" />
              </div>
            </div>
            <div className="min-w-0 flex-1">
              <h1 className="text-2xl font-bold text-foreground truncate">{service.name}</h1>
              {service.description && (
                <p className="mt-2 text-muted-foreground">{service.description}</p>
              )}
              <div className="mt-3 flex items-center space-x-4 text-sm text-muted-foreground">
                <div className="flex items-center space-x-1">
                  <UserIcon className="h-4 w-4" />
                  <span>ID: {service.user_id}</span>
                </div>
                <div className="flex items-center space-x-1">
                  <ClockIcon className="h-4 w-4" />
                  <span>Updated {formatDate(service.updated_at)}</span>
                </div>
              </div>
            </div>
          </div>

          <div className="flex items-center space-x-3 ml-4">
            <StatusBadge status={status} />
            {isDeveloper && status !== 'deleted' && (
              <div className="flex items-center space-x-2">
                {onEdit && (
                  <button
                    onClick={onEdit}
                    className="p-2 text-muted-foreground hover:text-blue-500 hover:bg-blue-50 rounded-md transition-colors"
                    title="Edit service"
                  >
                    <PencilIcon className="h-5 w-5" />
                  </button>
                )}
                {onDelete && (
                  <button
                    onClick={handleDelete}
                    disabled={isDeleting}
                    className="p-2 text-muted-foreground hover:text-red-500 hover:bg-red-50 rounded-md transition-colors disabled:opacity-50"
                    title="Delete service"
                  >
                    <TrashIcon className="h-5 w-5" />
                  </button>
                )}
              </div>
            )}
          </div>
        </div>

        {/* Actions */}
        {isDeveloper && status !== 'deleted' && (
          <div className="mt-6 pt-4 border-t border-border">
            <div className="flex items-center space-x-3">
              {onValidate && (
                <button
                  onClick={onValidate}
                  className="inline-flex items-center px-3 py-2 border border-border rounded-md shadow-sm text-sm font-medium text-foreground bg-background hover:bg-muted focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                >
                  Validate Service
                </button>
              )}
              {onBuild && (
                <button
                  onClick={onBuild}
                  className="inline-flex items-center px-3 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                >
                  Build Service
                </button>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Containers Section */}
      <div className="bg-card rounded-lg border border-border p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-foreground">Containers ({containerCount})</h2>
          {isDeveloper && status !== 'deleted' && onAddContainer && (
            <button
              onClick={onAddContainer}
              className="inline-flex items-center px-3 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              <PlusIcon className="h-4 w-4 mr-2" />
              Add Container
            </button>
          )}
        </div>

        {containerCount === 0 ? (
          <div className="text-center py-8">
            <CubeIcon className="mx-auto h-12 w-12 text-muted-foreground" />
            <h3 className="mt-2 text-sm font-medium text-foreground">No containers</h3>
            <p className="mt-1 text-sm text-muted-foreground">
              {isDeveloper
                ? 'Get started by adding a container to this service.'
                : 'This service has no containers configured.'}
            </p>
          </div>
        ) : (
          <div className="space-y-3">
            {service.containers?.map((serviceContainer) => (
              <div
                key={serviceContainer.id}
                className="flex items-center justify-between p-4 border border-border rounded-lg"
              >
                <div className="flex items-center space-x-3 min-w-0 flex-1">
                  <CubeIcon className="h-5 w-5 text-blue-500 flex-shrink-0" />
                  <div className="min-w-0 flex-1">
                    <div className="flex items-center space-x-2">
                      <span className="text-sm font-medium text-foreground">
                        {serviceContainer.container?.name ||
                          `Container ${serviceContainer.container_id}`}
                      </span>
                      <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800">
                        v{serviceContainer.container_version}
                      </span>
                    </div>
                    {serviceContainer.container?.description && (
                      <p className="text-sm text-muted-foreground truncate">
                        {serviceContainer.container.description}
                      </p>
                    )}
                    <div className="text-xs text-muted-foreground">
                      Order: {serviceContainer.order}
                    </div>
                  </div>
                </div>

                {isDeveloper && status !== 'deleted' && (
                  <div className="flex items-center space-x-2 ml-4">
                    {onEditContainer && (
                      <button
                        onClick={() => onEditContainer(serviceContainer.id)}
                        className="p-1 text-muted-foreground hover:text-blue-500 transition-colors"
                        title="Edit container"
                      >
                        <PencilIcon className="h-4 w-4" />
                      </button>
                    )}
                    <button
                      onClick={() => handleRemoveContainer(serviceContainer.id)}
                      disabled={removingContainerIds.has(serviceContainer.id)}
                      className="p-1 text-muted-foreground hover:text-red-500 transition-colors disabled:opacity-50"
                      title="Remove container"
                    >
                      <TrashIcon className="h-4 w-4" />
                    </button>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Loading overlay */}
      {isDeleting && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-card rounded-lg p-6 flex items-center space-x-3">
            <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-blue-600"></div>
            <span className="text-foreground">Deleting service...</span>
          </div>
        </div>
      )}
    </div>
  );
};

export default ServiceDetail;
