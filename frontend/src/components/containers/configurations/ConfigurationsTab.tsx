import { useState } from 'react';
import { useContainerConfigurations } from '../../../hooks/useContainerConfigurations';
import { ContainerConfiguration } from '../../../services/configurationService';
import { ContainerVersion } from '../../../types/container';

interface ConfigurationsTabProps {
  containerId: string;
  versions: ContainerVersion[];
  onVersionUpdate: () => void;
}

export function ConfigurationsTab({ containerId, versions, onVersionUpdate: _onVersionUpdate }: ConfigurationsTabProps) {
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
          <h2 className="text-2xl font-bold text-gray-900">Configurations</h2>
          <p className="text-sm text-gray-600 mt-1">
            Manage configuration templates and assign them to versions
          </p>
        </div>
      </div>

      {/* Split Panel Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
        {/* Left Panel - Configuration List (40%) */}
        <div className="lg:col-span-2">
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Configuration List</h3>

            {configurations.length === 0 ? (
              <div className="text-center py-12">
                <svg
                  className="mx-auto h-12 w-12 text-gray-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                  />
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                  />
                </svg>
                <h3 className="mt-4 text-sm font-medium text-gray-900">No configurations</h3>
                <p className="mt-2 text-sm text-gray-500">
                  Create your first configuration template to get started.
                </p>
                <button className="mt-6 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
                  Create Configuration
                </button>
              </div>
            ) : (
              <div className="space-y-3">
                {configurations.map((config) => {
                  const versionsUsing = getVersionsUsingConfig(config.id, versions);
                  const inUse = versionsUsing.length > 0;
                  const isSelected = selectedConfig?.id === config.id;

                  return (
                    <div
                      key={config.id}
                      onClick={() => setSelectedConfig(config)}
                      className={`p-4 rounded-lg border-2 cursor-pointer transition-all ${
                        isSelected
                          ? 'border-blue-500 bg-blue-50'
                          : inUse
                          ? 'border-blue-200 bg-white hover:border-blue-300'
                          : 'border-gray-200 bg-white hover:border-gray-300'
                      }`}
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <h4 className="font-semibold text-gray-900">{config.name}</h4>
                          <div className="mt-1 flex items-center gap-2">
                            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800">
                              Min: {config.minimum_version}
                            </span>
                            {inUse && (
                              <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800">
                                {versionsUsing.length} version{versionsUsing.length !== 1 ? 's' : ''}
                              </span>
                            )}
                          </div>
                        </div>
                      </div>
                      {config.description && (
                        <p className="mt-2 text-sm text-gray-600">{config.description}</p>
                      )}
                      <div className="mt-3 flex gap-2">
                        <button className="px-3 py-1 text-sm text-blue-600 hover:text-blue-700 font-medium">
                          Edit
                        </button>
                        <button
                          className={`px-3 py-1 text-sm font-medium ${
                            inUse
                              ? 'text-gray-400 cursor-not-allowed'
                              : 'text-red-600 hover:text-red-700'
                          }`}
                          disabled={inUse}
                        >
                          Delete
                        </button>
                      </div>
                    </div>
                  );
                })}
                <button className="w-full mt-4 px-4 py-2 border-2 border-dashed border-gray-300 rounded-lg text-gray-600 hover:border-blue-400 hover:text-blue-600">
                  + New Configuration
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Right Panel - Version Assignment (60%) */}
        <div className="lg:col-span-3">
          <div className="bg-white rounded-lg border border-gray-200 p-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Assign to Versions</h3>

            {!selectedConfig ? (
              <div className="text-center py-12">
                <svg
                  className="mx-auto h-12 w-12 text-gray-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122"
                  />
                </svg>
                <h3 className="mt-4 text-sm font-medium text-gray-900">Select a configuration</h3>
                <p className="mt-2 text-sm text-gray-500">
                  Choose a configuration from the left to assign it to versions.
                </p>
              </div>
            ) : (
              <div className="space-y-3">
                <div className="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-lg">
                  <p className="text-sm text-blue-800">
                    <span className="font-semibold">{selectedConfig.name}</span> (Min:{' '}
                    {selectedConfig.minimum_version})
                  </p>
                  <p className="text-xs text-blue-600 mt-1">
                    Click or drag to assign to compatible versions
                  </p>
                </div>

                {versions.length === 0 ? (
                  <div className="text-center py-8 text-gray-500 text-sm">
                    No versions available for this container.
                  </div>
                ) : (
                  versions.map((version) => {
                    const isAssigned = version.configuration_id === selectedConfig.id;
                    const isCompatible = true; // Will implement actual check

                    return (
                      <div
                        key={version.id}
                        className={`p-4 rounded-lg border-2 ${
                          isAssigned
                            ? 'border-green-500 bg-green-50'
                            : isCompatible
                            ? 'border-gray-200 bg-white hover:border-blue-300'
                            : 'border-red-200 bg-red-50'
                        }`}
                      >
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-3">
                            <span className="font-semibold text-gray-900">{version.version}</span>
                            {isAssigned && (
                              <svg
                                className="h-5 w-5 text-green-600"
                                fill="currentColor"
                                viewBox="0 0 20 20"
                              >
                                <path
                                  fillRule="evenodd"
                                  d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                                  clipRule="evenodd"
                                />
                              </svg>
                            )}
                            {!isAssigned && !isCompatible && (
                              <svg
                                className="h-5 w-5 text-red-600"
                                fill="currentColor"
                                viewBox="0 0 20 20"
                              >
                                <path
                                  fillRule="evenodd"
                                  d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                                  clipRule="evenodd"
                                />
                              </svg>
                            )}
                          </div>
                          <div className="flex gap-2">
                            {isAssigned ? (
                              <button className="px-3 py-1 text-sm text-red-600 hover:text-red-700 font-medium">
                                Unassign
                              </button>
                            ) : isCompatible ? (
                              <button className="px-3 py-1 text-sm text-blue-600 hover:text-blue-700 font-medium">
                                Assign
                              </button>
                            ) : (
                              <span className="px-3 py-1 text-sm text-gray-400 font-medium">
                                Incompatible
                              </span>
                            )}
                          </div>
                        </div>
                      </div>
                    );
                  })
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
