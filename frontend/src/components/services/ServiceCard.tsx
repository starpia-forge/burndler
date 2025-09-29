import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import {
  EyeIcon,
  PencilIcon,
  TrashIcon,
  Squares2X2Icon,
  ClockIcon,
  CubeIcon,
} from '@heroicons/react/24/outline';
import { Service } from '../../types/service';
import { StatusBadge, getServiceStatus } from '../common/StatusBadge';
import { useAuth } from '../../hooks/useAuth';

interface ServiceCardProps {
  service: Service;
  onEdit?: (service: Service) => void;
  onDelete?: (service: Service) => void;
  className?: string;
  showActions?: boolean;
}

export const ServiceCard: React.FC<ServiceCardProps> = ({
  service,
  onEdit,
  onDelete,
  className = '',
  showActions = true,
}) => {
  const { isDeveloper } = useAuth();
  const [isDeleting, setIsDeleting] = useState(false);

  const status = getServiceStatus(service);
  const containerCount = service.containers?.length || 0;

  const handleEdit = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (onEdit) {
      onEdit(service);
    }
  };

  const handleDelete = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();

    if (!onDelete || isDeleting) return;

    const confirmed = window.confirm(
      `Are you sure you want to delete service "${service.name}"?\n\nThis action cannot be undone.`
    );

    if (confirmed) {
      setIsDeleting(true);
      try {
        await onDelete(service);
      } catch (error) {
        console.error('Failed to delete service:', error);
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
    group relative bg-card rounded-lg border border-border
    hover:border-border/80 hover:shadow-md transition-all duration-200
    ${status === 'deleted' ? 'opacity-60' : ''}
    ${className}
  `.trim();

  return (
    <Link to={`/services/${service.id}`} className="block">
      <div className={cardClasses}>
        {/* Header */}
        <div className="p-6 pb-4">
          <div className="flex items-start justify-between">
            <div className="flex items-start space-x-3 min-w-0 flex-1">
              <div className="flex-shrink-0">
                <Squares2X2Icon className="h-8 w-8 text-blue-500" />
              </div>
              <div className="min-w-0 flex-1">
                <h3 className="text-lg font-semibold text-foreground truncate group-hover:text-blue-600 transition-colors">
                  {service.name}
                </h3>
                {service.description && (
                  <p className="mt-1 text-sm text-muted-foreground line-clamp-2">
                    {service.description}
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
                    className="p-1 text-muted-foreground hover:text-blue-500 transition-colors"
                    title="Edit service"
                  >
                    <PencilIcon className="h-4 w-4" />
                  </button>
                  <button
                    onClick={handleDelete}
                    disabled={isDeleting}
                    className="p-1 text-muted-foreground hover:text-red-500 transition-colors disabled:opacity-50"
                    title="Delete service"
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
          {/* Container info */}
          <div className="mt-3 flex items-center justify-between">
            <div className="flex items-center space-x-4 text-sm">
              <div className="text-muted-foreground flex items-center space-x-1">
                <CubeIcon className="h-4 w-4" />
                <span className="font-medium">{containerCount}</span>
                <span>container{containerCount !== 1 ? 's' : ''}</span>
              </div>
            </div>

            <div className="flex items-center space-x-1 text-xs text-muted-foreground">
              <ClockIcon className="h-3 w-3" />
              <span>Updated {formatDate(service.updated_at)}</span>
            </div>
          </div>
        </div>

        {/* Footer actions (visible on mobile) */}
        {showActions && (
          <div className="px-6 py-3 bg-muted/50 rounded-b-lg border-t border-border sm:hidden">
            <div className="flex items-center justify-between">
              <Link
                to={`/services/${service.id}`}
                className="inline-flex items-center space-x-1 text-sm text-blue-600 hover:text-blue-700"
              >
                <EyeIcon className="h-4 w-4" />
                <span>View Details</span>
              </Link>

              {isDeveloper && status !== 'deleted' && (
                <div className="flex items-center space-x-3">
                  <button
                    onClick={handleEdit}
                    className="inline-flex items-center space-x-1 text-sm text-muted-foreground hover:text-foreground"
                  >
                    <PencilIcon className="h-4 w-4" />
                    <span>Edit</span>
                  </button>
                  <button
                    onClick={handleDelete}
                    disabled={isDeleting}
                    className="inline-flex items-center space-x-1 text-sm text-red-600 hover:text-red-700 disabled:opacity-50"
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
          <div className="absolute inset-0 bg-card bg-opacity-75 flex items-center justify-center rounded-lg">
            <div className="flex items-center space-x-2 text-muted-foreground">
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-current"></div>
              <span className="text-sm">Deleting...</span>
            </div>
          </div>
        )}
      </div>
    </Link>
  );
};

export default ServiceCard;
