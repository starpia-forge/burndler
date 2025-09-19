import { useState } from 'react';
import { CogIcon, BuildingOfficeIcon } from '@heroicons/react/24/outline';
import { useSetupWizardContext } from '../../contexts/SetupWizardContext';

interface SystemConfigProps {
  onConfigComplete: () => void;
}

export default function SystemConfig({ onConfigComplete }: SystemConfigProps) {
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
      setFormError('Company name is required');
      return false;
    }
    if (!formData.systemSettings.default_namespace.trim()) {
      setFormError('Default namespace is required');
      return false;
    }
    if (!/^[a-z0-9-]+$/.test(formData.systemSettings.default_namespace)) {
      setFormError('Namespace can only contain lowercase letters, numbers, and hyphens');
      return false;
    }
    if (
      formData.systemSettings.notification_email &&
      !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.systemSettings.notification_email)
    ) {
      setFormError('Please enter a valid notification email address');
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
        <CogIcon className="h-16 w-16 text-blue-600 mx-auto mb-4" />
        <h2 className="text-2xl font-bold text-gray-900 mb-2">System Configuration</h2>
        <p className="text-gray-600">
          Configure your organization and system settings to complete the setup.
        </p>
      </div>

      {formError && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
          <div className="text-sm text-red-700">{formError}</div>
        </div>
      )}

      <form onSubmit={handleSubmit} className="max-w-2xl mx-auto space-y-8">
        {/* Company Information */}
        <div className="bg-gray-50 rounded-lg p-6">
          <div className="flex items-center mb-4">
            <BuildingOfficeIcon className="h-6 w-6 text-gray-400 mr-2" />
            <h3 className="text-lg font-medium text-gray-900">Company Information</h3>
          </div>

          <div>
            <label htmlFor="companyName" className="block text-sm font-medium text-gray-700 mb-2">
              Company Name *
            </label>
            <input
              type="text"
              id="companyName"
              name="companyName"
              value={formData.companyName}
              onChange={handleInputChange}
              required
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
              placeholder="Your Company Name"
            />
          </div>
        </div>

        {/* System Settings */}
        <div className="bg-gray-50 rounded-lg p-6">
          <div className="flex items-center mb-4">
            <CogIcon className="h-6 w-6 text-gray-400 mr-2" />
            <h3 className="text-lg font-medium text-gray-900">System Settings</h3>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label
                htmlFor="default_namespace"
                className="block text-sm font-medium text-gray-700 mb-2"
              >
                Default Namespace *
              </label>
              <input
                type="text"
                id="default_namespace"
                name="default_namespace"
                value={formData.systemSettings.default_namespace}
                onChange={handleInputChange}
                required
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                placeholder="burndler"
              />
              <p className="text-xs text-gray-500 mt-1">Used for container and service prefixes</p>
            </div>

            <div>
              <label
                htmlFor="max_concurrent_builds"
                className="block text-sm font-medium text-gray-700 mb-2"
              >
                Max Concurrent Builds
              </label>
              <select
                id="max_concurrent_builds"
                name="max_concurrent_builds"
                value={formData.systemSettings.max_concurrent_builds}
                onChange={handleInputChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="1">1</option>
                <option value="2">2</option>
                <option value="3">3</option>
                <option value="5">5</option>
                <option value="10">10</option>
              </select>
              <p className="text-xs text-gray-500 mt-1">Number of simultaneous package builds</p>
            </div>

            <div>
              <label
                htmlFor="storage_retention_days"
                className="block text-sm font-medium text-gray-700 mb-2"
              >
                Storage Retention (Days)
              </label>
              <input
                type="number"
                id="storage_retention_days"
                name="storage_retention_days"
                value={formData.systemSettings.storage_retention_days}
                onChange={handleInputChange}
                min="1"
                max="365"
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
              />
              <p className="text-xs text-gray-500 mt-1">How long to keep build artifacts</p>
            </div>

            <div>
              <label
                htmlFor="auto_cleanup_enabled"
                className="block text-sm font-medium text-gray-700 mb-2"
              >
                Auto Cleanup
              </label>
              <select
                id="auto_cleanup_enabled"
                name="auto_cleanup_enabled"
                value={formData.systemSettings.auto_cleanup_enabled}
                onChange={handleInputChange}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="true">Enabled</option>
                <option value="false">Disabled</option>
              </select>
              <p className="text-xs text-gray-500 mt-1">Automatically clean up old builds</p>
            </div>
          </div>

          <div className="mt-6">
            <label
              htmlFor="notification_email"
              className="block text-sm font-medium text-gray-700 mb-2"
            >
              Notification Email
            </label>
            <input
              type="email"
              id="notification_email"
              name="notification_email"
              value={formData.systemSettings.notification_email}
              onChange={handleInputChange}
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
              placeholder="notifications@company.com"
            />
            <p className="text-xs text-gray-500 mt-1">Email for system notifications (optional)</p>
          </div>
        </div>

        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <div className="text-sm text-blue-700">
            <p className="font-medium mb-2">Configuration Summary:</p>
            <ul className="list-disc list-inside space-y-1">
              <li>Company profile will be set up with your organization name</li>
              <li>System will use the specified namespace for all deployments</li>
              <li>Build system will be configured with your performance settings</li>
              <li>Cleanup policies will be applied to manage storage</li>
            </ul>
          </div>
        </div>

        <div className="flex justify-end">
          <button
            type="submit"
            disabled={loading}
            className="bg-blue-600 text-white px-8 py-2 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? 'Saving Configuration...' : 'Save Configuration'}
          </button>
        </div>
      </form>
    </div>
  );
}
