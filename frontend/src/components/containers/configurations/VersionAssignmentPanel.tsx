import { ContainerVersion } from '../../../types/container';
import { ContainerConfiguration } from '../../../services/configurationService';
import { VersionAssignmentCard } from './VersionAssignmentCard';
import { isVersionCompatible } from '../../../utils/versionCompatibility';

interface VersionAssignmentPanelProps {
  selectedConfig: ContainerConfiguration | null;
  versions: ContainerVersion[];
  onAssignConfig: (versionId: number, configId: number) => void;
  onUnassignConfig: (versionId: number) => void;
}

export function VersionAssignmentPanel({
  selectedConfig,
  versions,
  onAssignConfig,
  onUnassignConfig,
}: VersionAssignmentPanelProps) {
  if (!selectedConfig) {
    return (
      <div className="bg-white rounded-lg border border-gray-200 p-6 h-full flex items-center justify-center">
        <div className="text-center max-w-md">
          <svg
            className="mx-auto h-20 w-20 text-gray-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122"
            />
          </svg>
          <h3 className="mt-6 text-base font-medium text-gray-900">Select a configuration</h3>
          <p className="mt-2 text-sm text-gray-500">
            Choose a configuration from the left panel to view compatible versions and assign it.
          </p>
          <div className="mt-6 p-4 bg-blue-50 border border-blue-200 rounded-lg text-left">
            <p className="text-sm font-medium text-blue-900">How it works:</p>
            <ul className="mt-2 text-xs text-blue-700 space-y-1 list-disc list-inside">
              <li>Click on a configuration to select it</li>
              <li>View which versions are compatible</li>
              <li>Click "Assign" or drag & drop to assign</li>
              <li>Each version can have one configuration</li>
            </ul>
          </div>
        </div>
      </div>
    );
  }

  // Calculate compatibility stats
  const compatibleVersions = versions.filter((v) =>
    isVersionCompatible(v.version, selectedConfig.minimum_version)
  );
  const assignedVersions = versions.filter((v) => v.configuration_id === selectedConfig.id);
  const incompatibleVersions = versions.filter(
    (v) => !isVersionCompatible(v.version, selectedConfig.minimum_version)
  );

  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4 h-full">
      {/* Header */}
      <div className="mb-4">
        <h3 className="text-lg font-semibold text-gray-900">Assign to Versions</h3>
        <p className="text-sm text-gray-600 mt-1">
          Manage version assignments for{' '}
          <span className="font-semibold text-gray-900">{selectedConfig.name}</span>
        </p>
      </div>

      {/* Selected Configuration Summary */}
      <div className="mb-6 p-4 bg-gradient-to-r from-blue-50 to-indigo-50 border border-blue-200 rounded-lg">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <h4 className="font-semibold text-blue-900">{selectedConfig.name}</h4>
            <div className="mt-2 flex items-center gap-3 flex-wrap">
              <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-blue-100 text-blue-800">
                Min Version: {selectedConfig.minimum_version}
              </span>
              <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-green-100 text-green-800">
                {assignedVersions.length} Assigned
              </span>
              <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-700">
                {compatibleVersions.length} Compatible
              </span>
            </div>
          </div>
          <svg className="h-12 w-12 text-blue-300" fill="currentColor" viewBox="0 0 20 20">
            <path
              fillRule="evenodd"
              d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z"
              clipRule="evenodd"
            />
          </svg>
        </div>
        <p className="text-xs text-blue-700 mt-3">
          💡 Drag this configuration to a compatible version below, or click the "Assign" button
        </p>
      </div>

      {/* Version List */}
      {versions.length === 0 ? (
        <div className="text-center py-12">
          <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"
            />
          </svg>
          <p className="mt-4 text-sm text-gray-500">No versions available for this container.</p>
          <p className="mt-1 text-xs text-gray-400">Create a version to assign configurations.</p>
        </div>
      ) : (
        <div className="space-y-3 max-h-[600px] overflow-y-auto pr-2">
          {/* Assigned Versions First */}
          {assignedVersions.length > 0 && (
            <div className="mb-4">
              <h4 className="text-sm font-semibold text-gray-700 mb-2 flex items-center gap-2">
                <svg className="h-4 w-4 text-green-600" fill="currentColor" viewBox="0 0 20 20">
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                    clipRule="evenodd"
                  />
                </svg>
                Currently Assigned ({assignedVersions.length})
              </h4>
              {assignedVersions.map((version) => (
                <div key={version.id} className="mb-2">
                  <VersionAssignmentCard
                    version={version}
                    selectedConfig={selectedConfig}
                    onAssign={() => onAssignConfig(version.id, selectedConfig.id)}
                    onUnassign={() => onUnassignConfig(version.id)}
                  />
                </div>
              ))}
            </div>
          )}

          {/* Compatible But Unassigned Versions */}
          {compatibleVersions.filter((v) => v.configuration_id !== selectedConfig.id).length > 0 && (
            <div className="mb-4">
              <h4 className="text-sm font-semibold text-gray-700 mb-2 flex items-center gap-2">
                <svg className="h-4 w-4 text-blue-600" fill="currentColor" viewBox="0 0 20 20">
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                    clipRule="evenodd"
                  />
                </svg>
                Compatible Versions ({compatibleVersions.filter((v) => v.configuration_id !== selectedConfig.id).length})
              </h4>
              {compatibleVersions
                .filter((v) => v.configuration_id !== selectedConfig.id)
                .map((version) => (
                  <div key={version.id} className="mb-2">
                    <VersionAssignmentCard
                      version={version}
                      selectedConfig={selectedConfig}
                      onAssign={() => onAssignConfig(version.id, selectedConfig.id)}
                      onUnassign={() => onUnassignConfig(version.id)}
                    />
                  </div>
                ))}
            </div>
          )}

          {/* Incompatible Versions */}
          {incompatibleVersions.length > 0 && (
            <div>
              <h4 className="text-sm font-semibold text-gray-700 mb-2 flex items-center gap-2">
                <svg className="h-4 w-4 text-red-600" fill="currentColor" viewBox="0 0 20 20">
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                    clipRule="evenodd"
                  />
                </svg>
                Incompatible Versions ({incompatibleVersions.length})
              </h4>
              {incompatibleVersions.map((version) => (
                <div key={version.id} className="mb-2">
                  <VersionAssignmentCard
                    version={version}
                    selectedConfig={selectedConfig}
                    onAssign={() => onAssignConfig(version.id, selectedConfig.id)}
                    onUnassign={() => onUnassignConfig(version.id)}
                  />
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
