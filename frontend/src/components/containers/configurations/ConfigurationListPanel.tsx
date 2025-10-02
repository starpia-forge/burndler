import { ContainerConfiguration } from '../../../services/configurationService';
import { ContainerVersion } from '../../../types/container';
import { ConfigurationCard } from './ConfigurationCard';

interface ConfigurationListPanelProps {
  configurations: ContainerConfiguration[];
  selectedConfig: ContainerConfiguration | null;
  versions: ContainerVersion[];
  onSelectConfig: (config: ContainerConfiguration) => void;
  onCreateConfig: () => void;
  onEditConfig: (config: ContainerConfiguration) => void;
  onDeleteConfig: (config: ContainerConfiguration) => void;
  getVersionsUsingConfig: (configId: number, versions: ContainerVersion[]) => ContainerVersion[];
}

export function ConfigurationListPanel({
  configurations,
  selectedConfig,
  versions,
  onSelectConfig,
  onCreateConfig,
  onEditConfig,
  onDeleteConfig,
  getVersionsUsingConfig,
}: ConfigurationListPanelProps) {
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4 h-full">
      {/* Header */}
      <div className="mb-4">
        <h3 className="text-lg font-semibold text-gray-900">Configuration Templates</h3>
        <p className="text-sm text-gray-600 mt-1">
          Manage reusable configuration templates for this container
        </p>
      </div>

      {/* Configuration List */}
      {configurations.length === 0 ? (
        <div className="text-center py-12">
          <svg
            className="mx-auto h-16 w-16 text-gray-400"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
            />
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
            />
          </svg>
          <h3 className="mt-4 text-base font-medium text-gray-900">No configurations</h3>
          <p className="mt-2 text-sm text-gray-500 max-w-sm mx-auto">
            Create your first configuration template to define reusable settings for container
            versions.
          </p>
          <button
            onClick={onCreateConfig}
            className="mt-6 px-5 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium shadow-sm hover:shadow transition-all inline-flex items-center gap-2"
          >
            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 4v16m8-8H4"
              />
            </svg>
            Create Configuration
          </button>
        </div>
      ) : (
        <div className="space-y-3">
          {/* Configuration Cards */}
          {configurations.map((config) => {
            const versionsUsing = getVersionsUsingConfig(config.id, versions);

            return (
              <ConfigurationCard
                key={config.id}
                config={config}
                isSelected={selectedConfig?.id === config.id}
                versionsUsingConfig={versionsUsing}
                onSelect={() => onSelectConfig(config)}
                onEdit={() => onEditConfig(config)}
                onDelete={() => onDeleteConfig(config)}
              />
            );
          })}

          {/* Create Button */}
          <button
            onClick={onCreateConfig}
            className="w-full mt-4 px-4 py-3 border-2 border-dashed border-gray-300 rounded-lg text-gray-600 hover:border-blue-400 hover:text-blue-600 hover:bg-blue-50 transition-all font-medium flex items-center justify-center gap-2"
          >
            <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 4v16m8-8H4"
              />
            </svg>
            Add New Configuration
          </button>
        </div>
      )}

      {/* Quick Stats */}
      {configurations.length > 0 && (
        <div className="mt-6 pt-4 border-t border-gray-200">
          <div className="grid grid-cols-2 gap-4 text-center">
            <div className="p-3 bg-gray-50 rounded-lg">
              <div className="text-2xl font-bold text-gray-900">{configurations.length}</div>
              <div className="text-xs text-gray-600 mt-1">Total Configs</div>
            </div>
            <div className="p-3 bg-blue-50 rounded-lg">
              <div className="text-2xl font-bold text-blue-900">
                {
                  configurations.filter((c) => getVersionsUsingConfig(c.id, versions).length > 0)
                    .length
                }
              </div>
              <div className="text-xs text-blue-600 mt-1">In Use</div>
            </div>
          </div>
        </div>
      )}

      {/* Helper Text */}
      {configurations.length > 0 && (
        <div className="mt-4 p-3 bg-blue-50 border border-blue-200 rounded-lg">
          <div className="flex items-start gap-2">
            <svg
              className="h-5 w-5 text-blue-600 flex-shrink-0 mt-0.5"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path
                fillRule="evenodd"
                d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                clipRule="evenodd"
              />
            </svg>
            <div>
              <p className="text-sm font-medium text-blue-900">Quick Tip</p>
              <p className="text-xs text-blue-700 mt-1">
                Select a configuration to see compatible versions, or drag it to a version to
                assign.
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
