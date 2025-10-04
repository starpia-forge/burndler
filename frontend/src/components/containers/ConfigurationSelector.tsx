import React, { useState, useEffect } from 'react';
import { PlusIcon, XMarkIcon } from '@heroicons/react/24/outline';
import { useContainerConfigurations } from '../../hooks/useContainerConfigurations';
import { CreateContainerConfigurationRequest } from '../../services/configurationService';
import { isValidSemanticVersion } from '../../utils/versionCompatibility';

interface ConfigurationSelectorProps {
  containerId: string;
  value: number | null;
  onChange: (configId: number | null) => void;
  disabled?: boolean;
}

export const ConfigurationSelector: React.FC<ConfigurationSelectorProps> = ({
  containerId,
  value,
  onChange,
  disabled = false,
}) => {
  const { configurations, loading, createConfig } = useContainerConfigurations({
    containerId,
    autoFetch: true,
  });

  const [showNewForm, setShowNewForm] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    minimum_version: 'v1.0.0',
    description: '',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [creating, setCreating] = useState(false);

  // Reset form when showing/hiding
  useEffect(() => {
    if (!showNewForm) {
      setFormData({ name: '', minimum_version: 'v1.0.0', description: '' });
      setErrors({});
    }
  }, [showNewForm]);

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Configuration name is required';
    } else if (!/^[a-zA-Z0-9_-]+$/.test(formData.name)) {
      newErrors.name = 'Name can only contain letters, numbers, hyphens, and underscores';
    } else if (configurations.some((c) => c.name === formData.name.trim())) {
      newErrors.name = 'A configuration with this name already exists';
    }

    if (!formData.minimum_version.trim()) {
      newErrors.minimum_version = 'Minimum version is required';
    } else if (!isValidSemanticVersion(formData.minimum_version)) {
      newErrors.minimum_version = 'Invalid semantic version format (e.g., v1.0.0 or 1.0.0)';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleCreateNew = async () => {
    if (!validateForm()) return;

    try {
      setCreating(true);
      const data: CreateContainerConfigurationRequest = {
        name: formData.name.trim(),
        minimum_version: formData.minimum_version.trim(),
        description: formData.description.trim() || undefined,
      };
      const newConfig = await createConfig(data);
      onChange(newConfig.id);
      setShowNewForm(false);
    } catch (err: any) {
      setErrors({ general: err.message || 'Failed to create configuration' });
    } finally {
      setCreating(false);
    }
  };

  const handleSelect = (configId: string) => {
    if (configId === 'new') {
      setShowNewForm(true);
    } else if (configId === '') {
      onChange(null);
      setShowNewForm(false);
    } else {
      onChange(parseInt(configId));
      setShowNewForm(false);
    }
  };

  return (
    <div className="space-y-4">
      {/* Dropdown Selection */}
      <div>
        <label
          htmlFor="configuration"
          className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1"
        >
          Configuration Template
        </label>
        <select
          id="configuration"
          value={showNewForm ? 'new' : value?.toString() || ''}
          onChange={(e) => handleSelect(e.target.value)}
          disabled={disabled || loading}
          className="block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-800 dark:text-white disabled:opacity-50"
        >
          <option value="">No Configuration</option>
          {configurations.map((config) => (
            <option key={config.id} value={config.id}>
              {config.name} (≥ {config.minimum_version})
            </option>
          ))}
          <option value="new">+ Create New Configuration</option>
        </select>
        <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
          Optional: Select a configuration template to apply to this version
        </p>
      </div>

      {/* Inline Create Form */}
      {showNewForm && (
        <div className="border border-blue-300 dark:border-blue-600 rounded-lg p-4 bg-blue-50 dark:bg-blue-900/20">
          <div className="flex items-center justify-between mb-4">
            <h4 className="text-sm font-semibold text-gray-900 dark:text-white">
              Create New Configuration
            </h4>
            <button
              type="button"
              onClick={() => setShowNewForm(false)}
              className="text-gray-400 hover:text-gray-500 dark:hover:text-gray-300"
              disabled={creating}
            >
              <XMarkIcon className="h-5 w-5" />
            </button>
          </div>

          {errors.general && (
            <div className="mb-4 p-2 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded text-sm text-red-600 dark:text-red-400">
              {errors.general}
            </div>
          )}

          <div className="space-y-3">
            <div>
              <label
                htmlFor="new-config-name"
                className="block text-xs font-medium text-gray-700 dark:text-gray-300"
              >
                Name <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                id="new-config-name"
                value={formData.name}
                onChange={(e) => {
                  setFormData({ ...formData, name: e.target.value });
                  if (errors.name) setErrors({ ...errors, name: '' });
                }}
                placeholder="e.g., default, production"
                disabled={creating}
                className={`mt-1 block w-full px-2 py-1.5 text-sm border rounded shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:text-white ${
                  errors.name
                    ? 'border-red-300 dark:border-red-600'
                    : 'border-gray-300 dark:border-gray-600'
                }`}
              />
              {errors.name && (
                <p className="mt-1 text-xs text-red-600 dark:text-red-400">{errors.name}</p>
              )}
            </div>

            <div>
              <label
                htmlFor="new-config-version"
                className="block text-xs font-medium text-gray-700 dark:text-gray-300"
              >
                Minimum Version <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                id="new-config-version"
                value={formData.minimum_version}
                onChange={(e) => {
                  setFormData({ ...formData, minimum_version: e.target.value });
                  if (errors.minimum_version) setErrors({ ...errors, minimum_version: '' });
                }}
                placeholder="v1.0.0"
                disabled={creating}
                className={`mt-1 block w-full px-2 py-1.5 text-sm border rounded shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:text-white ${
                  errors.minimum_version
                    ? 'border-red-300 dark:border-red-600'
                    : 'border-gray-300 dark:border-gray-600'
                }`}
              />
              {errors.minimum_version && (
                <p className="mt-1 text-xs text-red-600 dark:text-red-400">
                  {errors.minimum_version}
                </p>
              )}
            </div>

            <div>
              <label
                htmlFor="new-config-desc"
                className="block text-xs font-medium text-gray-700 dark:text-gray-300"
              >
                Description <span className="text-gray-400">(optional)</span>
              </label>
              <input
                type="text"
                id="new-config-desc"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                placeholder="Brief description..."
                disabled={creating}
                className="mt-1 block w-full px-2 py-1.5 text-sm border border-gray-300 dark:border-gray-600 rounded shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:text-white"
              />
            </div>

            <button
              type="button"
              onClick={handleCreateNew}
              disabled={creating}
              className="w-full inline-flex items-center justify-center gap-2 px-3 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
            >
              {creating ? (
                <>
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
                  Creating...
                </>
              ) : (
                <>
                  <PlusIcon className="h-4 w-4" />
                  Create Configuration
                </>
              )}
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
