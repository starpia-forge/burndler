import { useState, useEffect } from 'react';
import {
  ContainerConfiguration,
  CreateContainerConfigurationRequest,
} from '../../../services/configurationService';
import { isValidSemanticVersion } from '../../../utils/versionCompatibility';

interface ConfigurationFormProps {
  mode: 'create' | 'edit';
  initialData?: ContainerConfiguration;
  existingNames: string[];
  onSubmit: (data: CreateContainerConfigurationRequest) => Promise<void>;
  onCancel: () => void;
  isSubmitting: boolean;
}

export function ConfigurationForm({
  mode,
  initialData,
  existingNames,
  onSubmit,
  onCancel,
  isSubmitting,
}: ConfigurationFormProps) {
  const [formData, setFormData] = useState({
    name: initialData?.name || '',
    minimum_version: initialData?.minimum_version || 'v1.0.0',
    description: initialData?.description || '',
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [touched, setTouched] = useState<Record<string, boolean>>({});

  useEffect(() => {
    if (initialData) {
      setFormData({
        name: initialData.name,
        minimum_version: initialData.minimum_version,
        description: initialData.description || '',
      });
    }
  }, [initialData]);

  const validateField = (name: string, value: string): string => {
    switch (name) {
      case 'name':
        if (!value.trim()) {
          return 'Configuration name is required';
        }
        if (mode === 'create' && existingNames.includes(value.trim())) {
          return 'A configuration with this name already exists';
        }
        if (!/^[a-zA-Z0-9_-]+$/.test(value)) {
          return 'Name can only contain letters, numbers, hyphens, and underscores';
        }
        if (value.length < 2) {
          return 'Name must be at least 2 characters';
        }
        if (value.length > 50) {
          return 'Name must not exceed 50 characters';
        }
        return '';

      case 'minimum_version':
        if (!value.trim()) {
          return 'Minimum version is required';
        }
        if (!isValidSemanticVersion(value)) {
          return 'Invalid semantic version format (e.g., v1.0.0 or 1.0.0)';
        }
        return '';

      case 'description':
        if (value.length > 500) {
          return 'Description must not exceed 500 characters';
        }
        return '';

      default:
        return '';
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));

    // Clear error when user starts typing
    if (errors[name]) {
      setErrors((prev) => ({ ...prev, [name]: '' }));
    }
  };

  const handleBlur = (e: React.FocusEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setTouched((prev) => ({ ...prev, [name]: true }));

    const error = validateField(name, value);
    if (error) {
      setErrors((prev) => ({ ...prev, [name]: error }));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validate all fields
    const newErrors: Record<string, string> = {};
    Object.keys(formData).forEach((key) => {
      const error = validateField(key, formData[key as keyof typeof formData]);
      if (error) {
        newErrors[key] = error;
      }
    });

    setErrors(newErrors);
    setTouched({ name: true, minimum_version: true, description: true });

    if (Object.keys(newErrors).length > 0) {
      return;
    }

    try {
      await onSubmit({
        name: formData.name.trim(),
        minimum_version: formData.minimum_version.trim(),
        description: formData.description.trim() || undefined,
      });
    } catch (error) {
      // Error handling done by parent component
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Name Field */}
      <div>
        <label htmlFor="name" className="block text-sm font-medium text-gray-700">
          Configuration Name <span className="text-red-500">*</span>
        </label>
        <input
          type="text"
          id="name"
          name="name"
          value={formData.name}
          onChange={handleChange}
          onBlur={handleBlur}
          disabled={mode === 'edit' || isSubmitting}
          className={`mt-1 block w-full rounded-md shadow-sm sm:text-sm ${
            errors.name && touched.name
              ? 'border-red-300 focus:border-red-500 focus:ring-red-500'
              : 'border-gray-300 focus:border-blue-500 focus:ring-blue-500'
          } ${mode === 'edit' ? 'bg-gray-50 cursor-not-allowed' : ''}`}
          placeholder="e.g., default, production, development"
        />
        {mode === 'edit' && (
          <p className="mt-1 text-xs text-gray-500">Configuration name cannot be changed</p>
        )}
        {errors.name && touched.name && <p className="mt-1 text-sm text-red-600">{errors.name}</p>}
        <p className="mt-1 text-xs text-gray-500">
          Unique identifier for this configuration template
        </p>
      </div>

      {/* Minimum Version Field */}
      <div>
        <label htmlFor="minimum_version" className="block text-sm font-medium text-gray-700">
          Minimum Version <span className="text-red-500">*</span>
        </label>
        <input
          type="text"
          id="minimum_version"
          name="minimum_version"
          value={formData.minimum_version}
          onChange={handleChange}
          onBlur={handleBlur}
          disabled={isSubmitting}
          className={`mt-1 block w-full rounded-md shadow-sm sm:text-sm ${
            errors.minimum_version && touched.minimum_version
              ? 'border-red-300 focus:border-red-500 focus:ring-red-500'
              : 'border-gray-300 focus:border-blue-500 focus:ring-blue-500'
          }`}
          placeholder="v1.0.0"
        />
        {errors.minimum_version && touched.minimum_version && (
          <p className="mt-1 text-sm text-red-600">{errors.minimum_version}</p>
        )}
        <p className="mt-1 text-xs text-gray-500">
          Container versions must be ≥ this version to use this configuration
        </p>
      </div>

      {/* Description Field */}
      <div>
        <label htmlFor="description" className="block text-sm font-medium text-gray-700">
          Description <span className="text-gray-400">(optional)</span>
        </label>
        <textarea
          id="description"
          name="description"
          rows={3}
          value={formData.description}
          onChange={handleChange}
          onBlur={handleBlur}
          disabled={isSubmitting}
          className={`mt-1 block w-full rounded-md shadow-sm sm:text-sm ${
            errors.description && touched.description
              ? 'border-red-300 focus:border-red-500 focus:ring-red-500'
              : 'border-gray-300 focus:border-blue-500 focus:ring-blue-500'
          }`}
          placeholder="Brief description of this configuration template..."
        />
        {errors.description && touched.description && (
          <p className="mt-1 text-sm text-red-600">{errors.description}</p>
        )}
        <div className="mt-1 flex items-center justify-between">
          <p className="text-xs text-gray-500">Optional description for this configuration</p>
          <p className="text-xs text-gray-400">{formData.description.length}/500</p>
        </div>
      </div>

      {/* Info Box */}
      <div className="rounded-md bg-blue-50 p-4 border border-blue-200">
        <div className="flex">
          <div className="flex-shrink-0">
            <svg className="h-5 w-5 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
              <path
                fillRule="evenodd"
                d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                clipRule="evenodd"
              />
            </svg>
          </div>
          <div className="ml-3 flex-1">
            <h3 className="text-sm font-medium text-blue-800">About Configurations</h3>
            <div className="mt-2 text-sm text-blue-700">
              <ul className="list-disc list-inside space-y-1">
                <li>Configurations can be reused across multiple container versions</li>
                <li>Only versions that meet the minimum version requirement can use this config</li>
                <li>UI Schema and Dependency Rules can be added later in advanced settings</li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      {/* Actions */}
      <div className="flex items-center justify-end gap-3 pt-4 border-t border-gray-200">
        <button
          type="button"
          onClick={onCancel}
          disabled={isSubmitting}
          className="px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Cancel
        </button>
        <button
          type="submit"
          disabled={isSubmitting}
          className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2"
        >
          {isSubmitting && (
            <svg className="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
              <circle
                className="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                strokeWidth="4"
              />
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              />
            </svg>
          )}
          {mode === 'create' ? 'Create Configuration' : 'Save Changes'}
        </button>
      </div>
    </form>
  );
}
