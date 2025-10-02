import { useState } from 'react';
import { useContainerConfigurations } from '../../../hooks/useContainerConfigurations';
import { ContainerConfiguration } from '../../../services/configurationService';
import { ContainerVersion } from '../../../types/container';
import { ConfigurationListPanel } from './ConfigurationListPanel';
import { VersionAssignmentPanel } from './VersionAssignmentPanel';

interface ConfigurationsTabProps {
  containerId: string;
  versions: ContainerVersion[];
  onVersionUpdate: () => void;
}

export function ConfigurationsTab({ containerId, versions, onVersionUpdate }: ConfigurationsTabProps) {
  const [selectedConfig, setSelectedConfig] = useState<ContainerConfiguration | null>(null);

  const {
    configurations,
    loading,
    error,
    refetch,
    getVersionsUsingConfig,
  } = useContainerConfigurations({
    containerId,
    autoFetch: true,
  });

  // Placeholder handlers - will be implemented in Phase 4
  const handleCreateConfig = () => {
    console.log('Create configuration - modal will be implemented in Phase 4');
  };

  const handleEditConfig = (_config: ContainerConfiguration) => {
    console.log('Edit configuration - modal will be implemented in Phase 4');
  };

  const handleDeleteConfig = (_config: ContainerConfiguration) => {
    console.log('Delete configuration - confirmation will be implemented in Phase 4');
  };

  const handleAssignConfig = (_versionId: number, _configId: number) => {
    console.log('Assign configuration - API call will be implemented in Phase 4');
    // Will call containerService.updateVersion to set configuration_id
  };

  const handleUnassignConfig = (_versionId: number) => {
    console.log('Unassign configuration - API call will be implemented in Phase 4');
    // Will call containerService.updateVersion to set configuration_id to null
  };

  if (loading && configurations.length === 0) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-800">
        <p className="font-semibold">Error loading configurations</p>
        <p className="text-sm mt-1">{error}</p>
        <button
          onClick={refetch}
          className="mt-3 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Configuration Management</h2>
          <p className="text-sm text-gray-600 mt-1">
            Create and manage reusable configuration templates, then assign them to versions
          </p>
        </div>
      </div>

      {/* Split Panel Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-5 gap-6 min-h-[600px]">
        {/* Left Panel - Configuration List (40%) */}
        <div className="lg:col-span-2">
          <ConfigurationListPanel
            configurations={configurations}
            selectedConfig={selectedConfig}
            versions={versions}
            onSelectConfig={setSelectedConfig}
            onCreateConfig={handleCreateConfig}
            onEditConfig={handleEditConfig}
            onDeleteConfig={handleDeleteConfig}
            getVersionsUsingConfig={getVersionsUsingConfig}
          />
        </div>

        {/* Right Panel - Version Assignment (60%) */}
        <div className="lg:col-span-3">
          <VersionAssignmentPanel
            selectedConfig={selectedConfig}
            versions={versions}
            onAssignConfig={handleAssignConfig}
            onUnassignConfig={handleUnassignConfig}
          />
        </div>
      </div>
    </div>
  );
}
