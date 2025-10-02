import { useState } from 'react';
import { ContainerConfiguration } from '../../../services/configurationService';
import { ContainerVersion } from '../../../types/container';
import { formatVersion } from '../../../utils/versionCompatibility';

interface ConfigurationCardProps {
  config: ContainerConfiguration;
  isSelected: boolean;
  versionsUsingConfig: ContainerVersion[];
  onSelect: () => void;
  onEdit: () => void;
  onDelete: () => void;
}

export function ConfigurationCard({
  config,
  isSelected,
  versionsUsingConfig,
  onSelect,
  onEdit,
  onDelete,
}: ConfigurationCardProps) {
  const [isDragging, setIsDragging] = useState(false);
  const inUse = versionsUsingConfig.length > 0;

  const handleDragStart = (e: React.DragEvent) => {
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('configId', config.id.toString());
    e.dataTransfer.setData('configName', config.name);
    e.dataTransfer.setData('configMinVersion', config.minimum_version);
    setIsDragging(true);
  };

  const handleDragEnd = () => {
    setIsDragging(false);
  };

  return (
    <div
      draggable
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
      onClick={onSelect}
      className={`p-4 rounded-lg border-2 cursor-pointer transition-all ${
        isDragging
          ? 'opacity-50 border-dashed border-blue-400'
          : isSelected
          ? 'border-blue-500 bg-blue-50 shadow-md'
          : inUse
          ? 'border-blue-200 bg-white hover:border-blue-300 hover:shadow'
          : 'border-gray-200 bg-white hover:border-gray-300 hover:shadow'
      }`}
      role="button"
      tabIndex={0}
      aria-selected={isSelected}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          onSelect();
        }
      }}
    >
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          {/* Configuration Name */}
          <h4 className="font-semibold text-gray-900 truncate">{config.name}</h4>

          {/* Badges */}
          <div className="mt-2 flex items-center gap-2 flex-wrap">
            {/* Minimum Version Badge */}
            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800">
              <svg
                className="mr-1 h-3 w-3"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"
                />
              </svg>
              Min: {formatVersion(config.minimum_version)}
            </span>

            {/* Usage Count Badge */}
            {inUse && (
              <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800">
                <svg
                  className="mr-1 h-3 w-3"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fillRule="evenodd"
                    d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z"
                    clipRule="evenodd"
                  />
                </svg>
                {versionsUsingConfig.length} version{versionsUsingConfig.length !== 1 ? 's' : ''}
              </span>
            )}

            {/* Drag Indicator */}
            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-500">
              <svg className="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4"
                />
              </svg>
            </span>
          </div>

          {/* Description */}
          {config.description && (
            <p className="mt-2 text-sm text-gray-600 line-clamp-2">{config.description}</p>
          )}

          {/* Metadata */}
          <div className="mt-3 flex items-center gap-4 text-xs text-gray-500">
            {config.created_at && (
              <span className="flex items-center gap-1">
                <svg className="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
                Created: {new Date(config.created_at).toLocaleDateString()}
              </span>
            )}
            {config.ui_schema && (
              <span className="flex items-center gap-1">
                <svg className="h-3 w-3" fill="currentColor" viewBox="0 0 20 20">
                  <path
                    fillRule="evenodd"
                    d="M6 2a2 2 0 00-2 2v12a2 2 0 002 2h8a2 2 0 002-2V7.414A2 2 0 0015.414 6L12 2.586A2 2 0 0010.586 2H6zm5 6a1 1 0 10-2 0v3.586l-1.293-1.293a1 1 0 10-1.414 1.414l3 3a1 1 0 001.414 0l3-3a1 1 0 00-1.414-1.414L11 11.586V8z"
                    clipRule="evenodd"
                  />
                </svg>
                Has UI Schema
              </span>
            )}
          </div>
        </div>
      </div>

      {/* Action Buttons */}
      <div className="mt-4 flex gap-2 pt-3 border-t border-gray-200">
        <button
          onClick={(e) => {
            e.stopPropagation();
            onEdit();
          }}
          className="flex-1 px-3 py-1.5 text-sm text-blue-600 hover:text-blue-700 hover:bg-blue-50 rounded font-medium transition-colors"
        >
          <span className="flex items-center justify-center gap-1">
            <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
              />
            </svg>
            Edit
          </span>
        </button>
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDelete();
          }}
          disabled={inUse}
          className={`flex-1 px-3 py-1.5 text-sm rounded font-medium transition-colors ${
            inUse
              ? 'text-gray-400 cursor-not-allowed bg-gray-50'
              : 'text-red-600 hover:text-red-700 hover:bg-red-50'
          }`}
          title={inUse ? 'Cannot delete: configuration is in use by versions' : 'Delete configuration'}
        >
          <span className="flex items-center justify-center gap-1">
            <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
              />
            </svg>
            Delete
          </span>
        </button>
      </div>
    </div>
  );
}
