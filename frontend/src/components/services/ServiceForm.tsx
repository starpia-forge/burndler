import React, { useState, useEffect } from 'react';
import { Service, CreateServiceRequest, UpdateServiceRequest } from '../../types/service';

interface ServiceFormProps {
  service?: Service;
  onSubmit: (data: CreateServiceRequest | UpdateServiceRequest) => Promise<void>;
  onCancel: () => void;
  loading?: boolean;
  submitLabel?: string;
  title?: string;
}

export const ServiceForm: React.FC<ServiceFormProps> = ({
  service,
  onSubmit,
  onCancel,
  loading = false,
  submitLabel,
  title,
}) => {
  const [formData, setFormData] = useState({
    name: '',
    description: '',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  const isEditing = !!service;
  const finalTitle = title || (isEditing ? 'Edit Service' : 'Create Service');
  const finalSubmitLabel = submitLabel || (isEditing ? 'Update Service' : 'Create Service');

  useEffect(() => {
    if (service) {
      setFormData({
        name: service.name,
        description: service.description,
      });
    }
  }, [service]);

  const validateForm = () => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Service name is required';
    } else if (formData.name.length < 2) {
      newErrors.name = 'Service name must be at least 2 characters';
    } else if (formData.name.length > 100) {
      newErrors.name = 'Service name must be less than 100 characters';
    }

    if (formData.description.length > 500) {
      newErrors.description = 'Description must be less than 500 characters';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    try {
      if (isEditing) {
        await onSubmit({
          description: formData.description,
          active: true, // Keep service active by default
        } as UpdateServiceRequest);
      } else {
        await onSubmit({
          name: formData.name.trim(),
          description: formData.description.trim(),
        } as CreateServiceRequest);
      }
    } catch (error) {
      console.error('Failed to save service:', error);
    }
  };

  const handleChange = (field: string, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: '' }));
    }
  };

  return (
    <div className="max-w-2xl mx-auto">
      <div className="bg-card rounded-lg border border-border p-6">
        <h2 className="text-xl font-semibold text-foreground mb-6">{finalTitle}</h2>

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Service Name */}
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-foreground mb-2">
              Service Name {!isEditing && <span className="text-red-500">*</span>}
            </label>
            <input
              type="text"
              id="name"
              value={formData.name}
              onChange={(e) => handleChange('name', e.target.value)}
              disabled={isEditing || loading}
              className="w-full px-3 py-2 border border-border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-background text-foreground disabled:bg-muted disabled:text-muted-foreground"
              placeholder="Enter service name"
            />
            {errors.name && (
              <p className="mt-1 text-sm text-red-600">{errors.name}</p>
            )}
            {isEditing && (
              <p className="mt-1 text-sm text-muted-foreground">
                Service name cannot be changed after creation
              </p>
            )}
          </div>

          {/* Description */}
          <div>
            <label htmlFor="description" className="block text-sm font-medium text-foreground mb-2">
              Description
            </label>
            <textarea
              id="description"
              value={formData.description}
              onChange={(e) => handleChange('description', e.target.value)}
              disabled={loading}
              rows={4}
              className="w-full px-3 py-2 border border-border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-background text-foreground resize-vertical"
              placeholder="Enter service description (optional)"
            />
            {errors.description && (
              <p className="mt-1 text-sm text-red-600">{errors.description}</p>
            )}
            <p className="mt-1 text-sm text-muted-foreground">
              {formData.description.length}/500 characters
            </p>
          </div>

          {/* Form Actions */}
          <div className="flex items-center justify-end space-x-3 pt-4 border-t border-border">
            <button
              type="button"
              onClick={onCancel}
              disabled={loading}
              className="px-4 py-2 border border-border rounded-md shadow-sm text-sm font-medium text-foreground bg-background hover:bg-muted focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? 'Saving...' : finalSubmitLabel}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default ServiceForm;