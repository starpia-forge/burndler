import React from 'react';
import { Service } from '../../types/service';
import { ServiceCard } from './ServiceCard';
import { PlusIcon } from '@heroicons/react/24/outline';
import { useAuth } from '../../hooks/useAuth';

interface ServiceListProps {
  services: Service[];
  loading?: boolean;
  onEdit?: (service: Service) => void;
  onCreate?: () => void;
  onDelete?: (service: Service) => Promise<void>;
  className?: string;
}

export const ServiceList: React.FC<ServiceListProps> = ({
  services,
  loading = false,
  onEdit,
  onCreate,
  onDelete,
  className = '',
}) => {
  const { isDeveloper } = useAuth();

  if (loading) {
    return (
      <div className={`space-y-4 ${className}`}>
        {[...Array(3)].map((_, index) => (
          <div
            key={index}
            className="h-32 bg-card rounded-lg border border-border animate-pulse"
          />
        ))}
      </div>
    );
  }

  if (services.length === 0) {
    return (
      <div className={`text-center py-12 ${className}`}>
        <div className="max-w-md mx-auto">
          <div className="h-16 w-16 mx-auto mb-4 bg-muted rounded-full flex items-center justify-center">
            <PlusIcon className="h-8 w-8 text-muted-foreground" />
          </div>
          <h3 className="text-lg font-medium text-foreground mb-2">
            No services found
          </h3>
          <p className="text-muted-foreground mb-6">
            {isDeveloper
              ? 'Get started by creating your first service.'
              : 'No services are available to view.'}
          </p>
          {isDeveloper && onCreate && (
            <button
              onClick={onCreate}
              className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              <PlusIcon className="h-4 w-4 mr-2" />
              Create Service
            </button>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className={`space-y-4 ${className}`}>
      {services.map((service) => (
        <ServiceCard
          key={service.id}
          service={service}
          onEdit={onEdit}
          onDelete={onDelete}
          showActions={true}
        />
      ))}
    </div>
  );
};

export default ServiceList;