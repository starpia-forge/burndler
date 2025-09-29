import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { CreateContainerRequest } from '../../types/container';

interface ContainerFormProps {
  onSubmit: (data: CreateContainerRequest) => Promise<void>;
  onCancel: () => void;
  loading?: boolean;
  title?: string;
  submitLabel?: string;
}

export const ContainerForm: React.FC<ContainerFormProps> = ({
  onSubmit,
  onCancel,
  loading = false,
  title,
  submitLabel,
}) => {
  const [formData, setFormData] = useState<CreateContainerRequest>({
    name: '',
    description: '',
    author: '',
    repository: '',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});
  const { t } = useTranslation(['containers', 'common']);

  const finalTitle = title || t('containers:createContainer');
  const finalSubmitLabel = submitLabel || t('containers:createContainer');

  const validateForm = () => {
    const newErrors: Record<string, string> = {};

    // Name validation (required, 1-100 chars)
    if (!formData.name.trim()) {
      newErrors.name = t('containers:nameRequired');
    } else if (formData.name.length < 1) {
      newErrors.name = t('containers:nameMinLength');
    } else if (formData.name.length > 100) {
      newErrors.name = t('containers:nameMaxLength');
    }

    // Description validation (optional, max 500 chars)
    if (formData.description && formData.description.length > 500) {
      newErrors.description = t('containers:descriptionMaxLength');
    }

    // Author validation (optional, max 100 chars)
    if (formData.author && formData.author.length > 100) {
      newErrors.author = t('containers:authorMaxLength');
    }

    // Repository validation (optional, max 200 chars)
    if (formData.repository && formData.repository.length > 200) {
      newErrors.repository = t('containers:repositoryMaxLength');
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
      await onSubmit({
        name: formData.name.trim(),
        description: formData.description.trim(),
        author: formData.author.trim(),
        repository: formData.repository.trim(),
      });
    } catch (error) {
      console.error('Failed to create container:', error);
    }
  };

  const handleChange = (field: keyof CreateContainerRequest, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: '' }));
    }
  };

  return (
    <div className="max-w-2xl mx-auto">
      <div className="bg-card rounded-lg border border-border p-6">
        <h2 className="text-xl font-semibold text-foreground mb-6">{finalTitle}</h2>

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Container Name */}
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-foreground mb-2">
              {t('containers:containerName')}{' '}
              <span className="text-red-500">{t('containers:required')}</span>
            </label>
            <input
              type="text"
              id="name"
              value={formData.name}
              onChange={(e) => handleChange('name', e.target.value)}
              disabled={loading}
              className="w-full px-3 py-2 border border-border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-background text-foreground disabled:bg-muted disabled:text-muted-foreground"
              placeholder={t('containers:enterContainerName')}
            />
            {errors.name && <p className="mt-1 text-sm text-red-600">{errors.name}</p>}
            <p className="mt-1 text-sm text-muted-foreground">
              {t('containers:nameCharacterCount', { count: formData.name.length, max: 100 })}
            </p>
          </div>

          {/* Description */}
          <div>
            <label htmlFor="description" className="block text-sm font-medium text-foreground mb-2">
              {t('common:description')}
            </label>
            <textarea
              id="description"
              value={formData.description}
              onChange={(e) => handleChange('description', e.target.value)}
              disabled={loading}
              rows={4}
              className="w-full px-3 py-2 border border-border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-background text-foreground resize-vertical"
              placeholder={t('containers:enterDescription')}
            />
            {errors.description && (
              <p className="mt-1 text-sm text-red-600">{errors.description}</p>
            )}
            <p className="mt-1 text-sm text-muted-foreground">
              {t('containers:descriptionCharacterCount', {
                count: formData.description.length,
                max: 500,
              })}
            </p>
          </div>

          {/* Author */}
          <div>
            <label htmlFor="author" className="block text-sm font-medium text-foreground mb-2">
              {t('containers:author')}
            </label>
            <input
              type="text"
              id="author"
              value={formData.author}
              onChange={(e) => handleChange('author', e.target.value)}
              disabled={loading}
              className="w-full px-3 py-2 border border-border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-background text-foreground disabled:bg-muted disabled:text-muted-foreground"
              placeholder={t('containers:enterAuthor')}
            />
            {errors.author && <p className="mt-1 text-sm text-red-600">{errors.author}</p>}
            <p className="mt-1 text-sm text-muted-foreground">
              {t('containers:authorCharacterCount', { count: formData.author.length, max: 100 })}
            </p>
          </div>

          {/* Repository */}
          <div>
            <label htmlFor="repository" className="block text-sm font-medium text-foreground mb-2">
              {t('containers:repository')}
            </label>
            <input
              type="text"
              id="repository"
              value={formData.repository}
              onChange={(e) => handleChange('repository', e.target.value)}
              disabled={loading}
              className="w-full px-3 py-2 border border-border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-background text-foreground disabled:bg-muted disabled:text-muted-foreground"
              placeholder={t('containers:enterRepository')}
            />
            {errors.repository && <p className="mt-1 text-sm text-red-600">{errors.repository}</p>}
            <p className="mt-1 text-sm text-muted-foreground">
              {t('containers:repositoryCharacterCount', {
                count: formData.repository.length,
                max: 200,
              })}
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
              {t('common:cancel')}
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? t('containers:creating') : finalSubmitLabel}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default ContainerForm;
