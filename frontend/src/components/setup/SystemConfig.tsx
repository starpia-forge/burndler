import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { CogIcon, BuildingOfficeIcon } from '@heroicons/react/24/outline';
import { useSetupWizardContext } from '../../contexts/SetupWizardContext';

interface SystemConfigProps {
  onConfigComplete: () => void;
}

export default function SystemConfig({ onConfigComplete }: SystemConfigProps) {
  const { t } = useTranslation('setup');
  const { setSystemConfig } = useSetupWizardContext();
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    companyName: '',
    systemSettings: {
      default_namespace: 'burndler',
      max_concurrent_builds: '3',
      storage_retention_days: '30',
      auto_cleanup_enabled: 'true',
      notification_email: '',
    },
  });
  const [formError, setFormError] = useState('');

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;

    if (name === 'companyName') {
      setFormData({
        ...formData,
        companyName: value,
      });
    } else {
      setFormData({
        ...formData,
        systemSettings: {
          ...formData.systemSettings,
          [name]: value,
        },
      });
    }
    setFormError('');
  };

  const validateForm = () => {
    if (!formData.companyName.trim()) {
      setFormError(t('configStep.errors.companyRequired'));
      return false;
    }
    if (!formData.systemSettings.default_namespace.trim()) {
      setFormError(t('configStep.errors.namespaceRequired'));
      return false;
    }
    if (!/^[a-z0-9-]+$/.test(formData.systemSettings.default_namespace)) {
      setFormError(t('configStep.errors.namespaceInvalid'));
      return false;
    }
    if (
      formData.systemSettings.notification_email &&
      !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.systemSettings.notification_email)
    ) {
      setFormError(t('configStep.errors.invalidEmail'));
      return false;
    }
    return true;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setLoading(true);

    // Save configuration data to context instead of calling API
    setSystemConfig({
      companyName: formData.companyName,
      systemSettings: formData.systemSettings,
    });

    // Simulate a brief loading state for better UX
    setTimeout(() => {
      setLoading(false);
      onConfigComplete();
    }, 500);
  };

  return (
    <div className="p-8">
      <div className="text-center mb-8">
        <CogIcon className="h-16 w-16 text-primary-600 mx-auto mb-4" />
        <h2 className="text-2xl font-bold text-foreground mb-2">{t('configStep.title')}</h2>
        <p className="text-muted-foreground">{t('configStep.description')}</p>
      </div>

      {formError && (
        <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4 mb-6">
          <div className="text-sm text-destructive">{formError}</div>
        </div>
      )}

      <form onSubmit={handleSubmit} className="max-w-2xl mx-auto space-y-8">
        {/* Company Information */}
        <div className="bg-muted/50 rounded-lg p-6 border border-border">
          <div className="flex items-center mb-4">
            <BuildingOfficeIcon className="h-6 w-6 text-muted-foreground mr-2" />
            <h3 className="text-lg font-medium text-foreground">
              {t('configStep.companyInformation')}
            </h3>
          </div>

          <div>
            <label htmlFor="companyName" className="block text-sm font-medium text-foreground mb-2">
              {t('configStep.companyNameLabel')}
            </label>
            <input
              type="text"
              id="companyName"
              name="companyName"
              value={formData.companyName}
              onChange={handleInputChange}
              required
              className="w-full px-3 py-2 bg-background border border-input rounded-md text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              placeholder={t('configStep.companyNamePlaceholder')}
            />
          </div>
        </div>

        {/* System Settings */}
        <div className="bg-muted/50 rounded-lg p-6 border border-border">
          <div className="flex items-center mb-4">
            <CogIcon className="h-6 w-6 text-muted-foreground mr-2" />
            <h3 className="text-lg font-medium text-foreground">
              {t('configStep.systemSettings')}
            </h3>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label
                htmlFor="default_namespace"
                className="block text-sm font-medium text-foreground mb-2"
              >
                {t('configStep.defaultNamespaceLabel')}
              </label>
              <input
                type="text"
                id="default_namespace"
                name="default_namespace"
                value={formData.systemSettings.default_namespace}
                onChange={handleInputChange}
                required
                className="w-full px-3 py-2 bg-background border border-input rounded-md text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                placeholder={t('configStep.defaultNamespacePlaceholder')}
              />
              <p className="text-xs text-muted-foreground mt-1">
                {t('configStep.defaultNamespaceHelp')}
              </p>
            </div>

            <div>
              <label
                htmlFor="max_concurrent_builds"
                className="block text-sm font-medium text-foreground mb-2"
              >
                {t('configStep.maxBuildsLabel')}
              </label>
              <select
                id="max_concurrent_builds"
                name="max_concurrent_builds"
                value={formData.systemSettings.max_concurrent_builds}
                onChange={handleInputChange}
                className="w-full px-3 py-2 bg-background border border-input rounded-md text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              >
                <option value="1">1</option>
                <option value="2">2</option>
                <option value="3">3</option>
                <option value="5">5</option>
                <option value="10">10</option>
              </select>
              <p className="text-xs text-muted-foreground mt-1">{t('configStep.maxBuildsHelp')}</p>
            </div>

            <div>
              <label
                htmlFor="storage_retention_days"
                className="block text-sm font-medium text-foreground mb-2"
              >
                {t('configStep.retentionLabel')}
              </label>
              <input
                type="number"
                id="storage_retention_days"
                name="storage_retention_days"
                value={formData.systemSettings.storage_retention_days}
                onChange={handleInputChange}
                min="1"
                max="365"
                className="w-full px-3 py-2 bg-background border border-input rounded-md text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              />
              <p className="text-xs text-muted-foreground mt-1">{t('configStep.retentionHelp')}</p>
            </div>

            <div>
              <label
                htmlFor="auto_cleanup_enabled"
                className="block text-sm font-medium text-foreground mb-2"
              >
                {t('configStep.autoCleanupLabel')}
              </label>
              <select
                id="auto_cleanup_enabled"
                name="auto_cleanup_enabled"
                value={formData.systemSettings.auto_cleanup_enabled}
                onChange={handleInputChange}
                className="w-full px-3 py-2 bg-background border border-input rounded-md text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              >
                <option value="true">{t('configStep.enabled')}</option>
                <option value="false">{t('configStep.disabled')}</option>
              </select>
              <p className="text-xs text-muted-foreground mt-1">
                {t('configStep.autoCleanupHelp')}
              </p>
            </div>
          </div>

          <div className="mt-6">
            <label
              htmlFor="notification_email"
              className="block text-sm font-medium text-foreground mb-2"
            >
              {t('configStep.notificationEmailLabel')}
            </label>
            <input
              type="email"
              id="notification_email"
              name="notification_email"
              value={formData.systemSettings.notification_email}
              onChange={handleInputChange}
              className="w-full px-3 py-2 bg-background border border-input rounded-md text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              placeholder={t('configStep.notificationEmailPlaceholder')}
            />
            <p className="text-xs text-muted-foreground mt-1">
              {t('configStep.notificationEmailHelp')}
            </p>
          </div>
        </div>

        <div className="bg-primary/10 border border-primary/20 rounded-lg p-4">
          <div className="text-sm text-primary-700">
            <p className="font-medium mb-2">{t('configStep.configSummaryTitle')}</p>
            <ul className="list-disc list-inside space-y-1">
              <li>{t('configStep.configSummaryItems.companyProfile')}</li>
              <li>{t('configStep.configSummaryItems.namespace')}</li>
              <li>{t('configStep.configSummaryItems.buildSystem')}</li>
              <li>{t('configStep.configSummaryItems.cleanup')}</li>
            </ul>
          </div>
        </div>

        <div className="flex justify-end">
          <button
            type="submit"
            disabled={loading}
            className="bg-primary-600 text-white px-8 py-2 rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? t('configStep.savingConfiguration') : t('configStep.saveConfiguration')}
          </button>
        </div>
      </form>
    </div>
  );
}
