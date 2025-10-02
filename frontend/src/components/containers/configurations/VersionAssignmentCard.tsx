import { useState } from 'react';
import { ContainerVersion } from '../../../types/container';
import { ContainerConfiguration } from '../../../services/configurationService';
import { isVersionCompatible, formatVersion } from '../../../utils/versionCompatibility';

interface VersionAssignmentCardProps {
  version: ContainerVersion;
  selectedConfig: ContainerConfiguration;
  onAssign: () => void;
  onUnassign: () => void;
}

export function VersionAssignmentCard({
  version,
  selectedConfig,
  onAssign,
  onUnassign,
}: VersionAssignmentCardProps) {
  const [isDropTarget, setIsDropTarget] = useState(false);

  const isAssigned = version.configuration_id === selectedConfig.id;
  const isCompatible = isVersionCompatible(version.version, selectedConfig.minimum_version);

  const handleDragOver = (e: React.DragEvent) => {
    if (isCompatible && !isAssigned) {
      e.preventDefault();
      e.dataTransfer.dropEffect = 'move';
      setIsDropTarget(true);
    }
  };

  const handleDragLeave = () => {
    setIsDropTarget(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDropTarget(false);

    const configId = e.dataTransfer.getData('configId');
    if (configId && parseInt(configId) === selectedConfig.id && isCompatible && !isAssigned) {
      onAssign();
    }
  };

  return (
    <div
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
      className={`p-4 rounded-lg border-2 transition-all ${
        isDropTarget
          ? 'border-blue-500 bg-blue-50 shadow-lg scale-105'
          : isAssigned
          ? 'border-green-500 bg-green-50'
          : isCompatible
          ? 'border-gray-200 bg-white hover:border-blue-300 hover:shadow'
          : 'border-red-200 bg-red-50'
      }`}
    >
      <div className="flex items-center justify-between">
        {/* Left: Version Info */}
        <div className="flex items-center gap-3 flex-1">
          {/* Version Number */}
          <span className="font-semibold text-gray-900">{formatVersion(version.version)}</span>

          {/* Status Icons */}
          {isAssigned && (
            <div className="flex items-center gap-1 px-2 py-1 bg-green-100 rounded-md">
              <svg className="h-5 w-5 text-green-600" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fillRule="evenodd"
                  d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                  clipRule="evenodd"
                />
              </svg>
              <span className="text-xs font-medium text-green-700">Assigned</span>
            </div>
          )}

          {!isAssigned && isCompatible && (
            <div className="flex items-center gap-1 px-2 py-1 bg-blue-100 rounded-md">
              <svg className="h-5 w-5 text-blue-600" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fillRule="evenodd"
                  d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                  clipRule="evenodd"
                />
              </svg>
              <span className="text-xs font-medium text-blue-700">Compatible</span>
            </div>
          )}

          {!isAssigned && !isCompatible && (
            <div className="flex items-center gap-1 px-2 py-1 bg-red-100 rounded-md">
              <svg className="h-5 w-5 text-red-600" fill="currentColor" viewBox="0 0 20 20">
                <path
                  fillRule="evenodd"
                  d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                  clipRule="evenodd"
                />
              </svg>
              <span className="text-xs font-medium text-red-700">Incompatible</span>
            </div>
          )}

          {/* Published Badge */}
          {version.published && (
            <span className="px-2 py-1 bg-gray-100 text-gray-600 text-xs font-medium rounded-md">
              Published
            </span>
          )}
        </div>

        {/* Right: Actions */}
        <div className="flex gap-2">
          {isAssigned ? (
            <button
              onClick={onUnassign}
              className="px-4 py-1.5 text-sm text-red-600 hover:text-red-700 hover:bg-red-50 rounded font-medium transition-colors border border-red-200 hover:border-red-300"
            >
              Unassign
            </button>
          ) : isCompatible ? (
            <button
              onClick={onAssign}
              className="px-4 py-1.5 text-sm text-blue-600 hover:text-blue-700 hover:bg-blue-50 rounded font-medium transition-colors border border-blue-200 hover:border-blue-300"
            >
              Assign
            </button>
          ) : (
            <span className="px-4 py-1.5 text-sm text-gray-400 font-medium cursor-not-allowed">
              Incompatible
            </span>
          )}
        </div>
      </div>

      {/* Compatibility Message */}
      {!isAssigned && (
        <div className="mt-3 pt-3 border-t border-gray-200">
          <p className="text-xs text-gray-600">
            {isCompatible ? (
              <>
                <span className="font-medium">✓ Compatible:</span> Version {formatVersion(version.version)} meets
                the minimum requirement ({formatVersion(selectedConfig.minimum_version)})
              </>
            ) : (
              <>
                <span className="font-medium">✗ Incompatible:</span> Version{' '}
                {formatVersion(version.version)} is below the minimum requirement (
                {formatVersion(selectedConfig.minimum_version)})
              </>
            )}
          </p>
        </div>
      )}

      {/* Drop Zone Indicator */}
      {isDropTarget && (
        <div className="mt-3 pt-3 border-t border-blue-300">
          <div className="flex items-center justify-center gap-2 text-blue-600">
            <svg className="h-5 w-5 animate-bounce" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M19 14l-7 7m0 0l-7-7m7 7V3"
              />
            </svg>
            <span className="text-sm font-medium">Drop to assign configuration</span>
          </div>
        </div>
      )}
    </div>
  );
}
