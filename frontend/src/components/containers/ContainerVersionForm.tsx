import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  CreateVersionRequest,
  UpdateVersionRequest,
  ContainerVersion,
} from '../../types/container';
import { ConfigurationSelector } from './ConfigurationSelector';

interface ContainerVersionFormProps {
  mode?: 'create' | 'edit';
  initialData?: ContainerVersion;
  containerId?: string;
  onSubmit: (data: CreateVersionRequest | UpdateVersionRequest) => void;
  onCancel: () => void;
  loading?: boolean;
  error?: string | null;
}

interface FormState {
  version: string;
  compose: string;
  variables?: Record<string, any>;
  resource_paths?: string[];
  dependencies?: Record<string, string>;
  configuration_id?: number | null;
}

const ContainerVersionForm: React.FC<ContainerVersionFormProps> = ({
  mode = 'create',
  initialData,
  containerId,
  onSubmit,
  onCancel,
  loading = false,
  error = null,
}) => {
  const { t } = useTranslation(['containers', 'common']);

  const [formData, setFormData] = useState<FormState>({
    version: initialData?.version || '',
    compose:
      initialData?.compose_content ||
      `version: '3.8'

services:
  # Add your services here
  example:
    image: nginx:alpine
    ports:
      - "80:80"
`,
    variables: initialData?.variables || {},
    resource_paths: initialData?.resource_paths || [],
    dependencies: initialData?.dependencies || {},
    configuration_id: initialData?.configuration_id || null,
  });

  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});

  const validateForm = (): boolean => {
    const errors: Record<string, string> = {};

    // Version validation (only for create mode)
    if (mode === 'create') {
      if (!formData.version?.trim()) {
        errors.version = t('containers:versionRequired');
      } else if (formData.version.length > 50) {
        errors.version = t('containers:versionMaxLength');
      } else if (!/^[a-zA-Z0-9._-]+$/.test(formData.version)) {
        errors.version = t('containers:versionInvalidFormat');
      }
    }

    // Compose validation
    if (!formData.compose?.trim()) {
      errors.compose = t('containers:composeRequired');
    } else if (formData.compose.length > 50000) {
      errors.compose = t('containers:composeMaxLength');
    }

    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (validateForm()) {
      if (mode === 'create') {
        const submitData: CreateVersionRequest = {
          version: formData.version.trim(),
          compose: formData.compose.trim(),
          variables: formData.variables,
          resource_paths: formData.resource_paths,
          dependencies: formData.dependencies,
          configuration_id: formData.configuration_id,
        };
        onSubmit(submitData);
      } else {
        const submitData: UpdateVersionRequest = {
          compose: formData.compose.trim(),
          variables: formData.variables,
          resource_paths: formData.resource_paths,
          dependencies: formData.dependencies,
          configuration_id: formData.configuration_id,
        };
        onSubmit(submitData);
      }
    }
  };

  const handleVersionChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setFormData((prev) => ({ ...prev, version: value }));

    // Clear validation error when user types
    if (validationErrors.version) {
      setValidationErrors((prev) => ({ ...prev, version: '' }));
    }
  };

  const handleComposeChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value;
    setFormData((prev) => ({ ...prev, compose: value }));

    // Clear validation error when user types
    if (validationErrors.compose) {
      setValidationErrors((prev) => ({ ...prev, compose: '' }));
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Error display */}
      {error && (
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
          <p className="text-red-700 dark:text-red-400">{error}</p>
        </div>
      )}

      {/* Version */}
      <div>
        <label
          htmlFor="version"
          className="block text-sm font-medium text-gray-700 dark:text-gray-300"
        >
          {t('containers:versionNumber')} <span className="text-red-500">*</span>
        </label>
        <div className="mt-1">
          <input
            type="text"
            id="version"
            value={formData.version}
            onChange={handleVersionChange}
            placeholder={t('containers:enterVersionNumber')}
            className={`block w-full px-3 py-2 border rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-800 dark:text-white dark:placeholder-gray-500 ${
              mode === 'edit'
                ? 'bg-gray-100 dark:bg-gray-700 cursor-not-allowed'
                : validationErrors.version
                  ? 'border-red-300 dark:border-red-600'
                  : 'border-gray-300 dark:border-gray-600'
            }`}
            disabled={loading || mode === 'edit'}
          />
          {mode === 'edit' && (
            <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
              Version number cannot be changed after creation
            </p>
          )}
          {validationErrors.version && (
            <p className="mt-1 text-sm text-red-600 dark:text-red-400">
              {validationErrors.version}
            </p>
          )}
          {mode === 'create' && (
            <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {t('containers:versionCharacterCount', {
                count: formData.version?.length || 0,
                max: 50,
              })}
            </p>
          )}
        </div>
      </div>

      {/* Configuration */}
      {(mode === 'edit' && initialData) || (mode === 'create' && containerId) ? (
        <ConfigurationSelector
          containerId={
            mode === 'edit' && initialData ? initialData.container_id.toString() : containerId || ''
          }
          value={formData.configuration_id || null}
          onChange={(configId) => setFormData((prev) => ({ ...prev, configuration_id: configId }))}
          disabled={loading}
        />
      ) : null}

      {/* Docker Compose YAML */}
      <div>
        <label
          htmlFor="compose"
          className="block text-sm font-medium text-gray-700 dark:text-gray-300"
        >
          {t('containers:dockerComposeYaml')} <span className="text-red-500">*</span>
        </label>
        <div className="mt-1">
          <textarea
            id="compose"
            rows={20}
            value={formData.compose}
            onChange={handleComposeChange}
            placeholder={t('containers:enterDockerCompose')}
            className={`block w-full px-3 py-2 border rounded-md shadow-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-800 dark:text-white dark:placeholder-gray-500 font-mono text-sm ${
              validationErrors.compose
                ? 'border-red-300 dark:border-red-600'
                : 'border-gray-300 dark:border-gray-600'
            }`}
            disabled={loading}
          />
          {validationErrors.compose && (
            <p className="mt-1 text-sm text-red-600 dark:text-red-400">
              {validationErrors.compose}
            </p>
          )}
          <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
            {t('containers:composeCharacterCount', {
              count: formData.compose.length,
              max: 50000,
            })}
          </p>
        </div>
      </div>

      {/* Form Actions */}
      <div className="flex items-center justify-end space-x-3 pt-6">
        <button
          type="button"
          onClick={onCancel}
          disabled={loading}
          className="inline-flex items-center px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
        >
          {t('common:cancel')}
        </button>
        <button
          type="submit"
          disabled={loading}
          className="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
        >
          {loading
            ? mode === 'edit'
              ? t('containers:updating')
              : t('containers:creating')
            : mode === 'edit'
              ? t('containers:updateVersion')
              : t('containers:createVersion')}
        </button>
      </div>
    </form>
  );
};

export default ContainerVersionForm;
