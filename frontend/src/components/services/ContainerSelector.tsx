import React, { useState, useEffect, useCallback } from 'react';
import {
  PlusIcon,
  TrashIcon,
  MagnifyingGlassIcon,
  CubeIcon,
  CheckIcon,
} from '@heroicons/react/24/outline';
import { Container, ContainerVersion } from '../../types/container';
import { ServiceContainerFormState, ContainerSelectorProps } from '../../types/service';
import { useContainers } from '../../hooks/useContainers';
import containerService from '../../services/containerService';

export const ContainerSelector: React.FC<ContainerSelectorProps> = ({
  onSelectionChange,
  initialSelection = [],
  disabled = false,
}) => {
  const [selectedContainers, setSelectedContainers] = useState<ServiceContainerFormState[]>(
    initialSelection
  );
  const [searchTerm, setSearchTerm] = useState('');
  const [showAddModal, setShowAddModal] = useState(false);
  const [loadingVersions, setLoadingVersions] = useState<Record<number, boolean>>({});
  const [containerVersions, setContainerVersions] = useState<Record<number, ContainerVersion[]>>({});

  const {
    containers,
    loading: containersLoading,
    error: containersError,
    updateFilter,
  } = useContainers({
    initialFilters: {
      page: 1,
      page_size: 20,
      active: true,
      published_only: true,
    },
  });

  // Update search filter when searchTerm changes
  useEffect(() => {
    updateFilter('search', searchTerm);
  }, [searchTerm, updateFilter]);

  // Notify parent of selection changes
  useEffect(() => {
    onSelectionChange(selectedContainers);
  }, [selectedContainers, onSelectionChange]);

  // Load versions for a container
  const loadContainerVersions = useCallback(async (containerId: number) => {
    if (containerVersions[containerId] || loadingVersions[containerId]) {
      return; // Already loaded or loading
    }

    setLoadingVersions(prev => ({ ...prev, [containerId]: true }));
    try {
      const response = await containerService.listVersions(containerId, { published_only: true });
      setContainerVersions(prev => ({
        ...prev,
        [containerId]: response.data,
      }));
    } catch (error) {
      console.error(`Failed to load versions for container ${containerId}:`, error);
    } finally {
      setLoadingVersions(prev => ({ ...prev, [containerId]: false }));
    }
  }, [containerVersions, loadingVersions]);

  const handleAddContainer = (container: Container) => {
    // Load versions for this container
    loadContainerVersions(container.id);

    // Add to selection with default values
    const newSelection: ServiceContainerFormState = {
      container_id: container.id,
      container_version: '', // Will be set when versions load
      variables: {},
      order: selectedContainers.length + 1,
    };

    setSelectedContainers(prev => [...prev, newSelection]);
    setShowAddModal(false);
  };

  const handleRemoveContainer = (index: number) => {
    setSelectedContainers(prev => {
      const updated = prev.filter((_, i) => i !== index);
      // Reorder remaining containers
      return updated.map((container, i) => ({
        ...container,
        order: i + 1,
      }));
    });
  };

  const handleVersionChange = (index: number, version: string) => {
    setSelectedContainers(prev =>
      prev.map((container, i) =>
        i === index ? { ...container, container_version: version } : container
      )
    );
  };

  const handleVariablesChange = (index: number, variables: Record<string, any>) => {
    setSelectedContainers(prev =>
      prev.map((container, i) =>
        i === index ? { ...container, variables } : container
      )
    );
  };

  const moveContainer = (fromIndex: number, toIndex: number) => {
    if (fromIndex === toIndex) return;

    setSelectedContainers(prev => {
      const updated = [...prev];
      const [moved] = updated.splice(fromIndex, 1);
      updated.splice(toIndex, 0, moved);

      // Update order values
      return updated.map((container, i) => ({
        ...container,
        order: i + 1,
      }));
    });
  };

  const getContainerName = (containerId: number) => {
    const container = containers.find(c => c.id === containerId);
    return container?.name || `Container ${containerId}`;
  };

  const getAvailableVersions = (containerId: number) => {
    return containerVersions[containerId] || [];
  };

  const isContainerAlreadySelected = (containerId: number) => {
    return selectedContainers.some(sc => sc.container_id === containerId);
  };

  // Set default version when versions load
  useEffect(() => {
    selectedContainers.forEach((selected, index) => {
      const versions = getAvailableVersions(selected.container_id);
      if (versions.length > 0 && !selected.container_version) {
        // Set to latest version (first in list, assuming sorted by version desc)
        handleVersionChange(index, versions[0].version);
      }
    });
  }, [containerVersions]); // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-medium text-foreground">
          Selected Containers ({selectedContainers.length})
        </h3>
        {!disabled && (
          <button
            onClick={() => setShowAddModal(true)}
            className="inline-flex items-center px-3 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            <PlusIcon className="h-4 w-4 mr-2" />
            Add Container
          </button>
        )}
      </div>

      {/* Selected Containers */}
      {selectedContainers.length === 0 ? (
        <div className="text-center py-8 border-2 border-dashed border-border rounded-lg">
          <CubeIcon className="mx-auto h-12 w-12 text-muted-foreground" />
          <h3 className="mt-2 text-sm font-medium text-foreground">No containers selected</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Add containers to compose your service
          </p>
        </div>
      ) : (
        <div className="space-y-3">
          {selectedContainers.map((selected, index) => {
            const versions = getAvailableVersions(selected.container_id);
            const isLoadingVersions = loadingVersions[selected.container_id];

            return (
              <div
                key={`${selected.container_id}-${index}`}
                className="bg-card border border-border rounded-lg p-4"
              >
                <div className="flex items-start justify-between">
                  <div className="flex items-start space-x-3 min-w-0 flex-1">
                    <div className="flex flex-col items-center space-y-1">
                      <span className="inline-flex items-center justify-center w-6 h-6 bg-blue-100 text-blue-800 text-xs font-medium rounded-full">
                        {selected.order}
                      </span>
                      {!disabled && (
                        <div className="flex flex-col space-y-1">
                          <button
                            onClick={() => moveContainer(index, Math.max(0, index - 1))}
                            disabled={index === 0}
                            className="p-1 text-muted-foreground hover:text-foreground disabled:opacity-30"
                            title="Move up"
                          >
                            ↑
                          </button>
                          <button
                            onClick={() => moveContainer(index, Math.min(selectedContainers.length - 1, index + 1))}
                            disabled={index === selectedContainers.length - 1}
                            className="p-1 text-muted-foreground hover:text-foreground disabled:opacity-30"
                            title="Move down"
                          >
                            ↓
                          </button>
                        </div>
                      )}
                    </div>

                    <div className="min-w-0 flex-1">
                      <h4 className="text-sm font-medium text-foreground">
                        {getContainerName(selected.container_id)}
                      </h4>

                      {/* Version Selection */}
                      <div className="mt-2">
                        <label className="block text-xs font-medium text-muted-foreground mb-1">
                          Version
                        </label>
                        {isLoadingVersions ? (
                          <div className="h-8 bg-muted rounded animate-pulse" />
                        ) : (
                          <select
                            value={selected.container_version}
                            onChange={(e) => handleVersionChange(index, e.target.value)}
                            disabled={disabled}
                            className="block w-full px-3 py-2 border border-border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-background text-foreground text-sm disabled:bg-muted disabled:text-muted-foreground"
                          >
                            <option value="">Select version</option>
                            {versions.map((version) => (
                              <option key={version.version} value={version.version}>
                                {version.version}
                              </option>
                            ))}
                          </select>
                        )}
                      </div>

                      {/* Variables (basic JSON editor) */}
                      <div className="mt-3">
                        <label className="block text-xs font-medium text-muted-foreground mb-1">
                          Variables (JSON)
                        </label>
                        <textarea
                          value={JSON.stringify(selected.variables, null, 2)}
                          onChange={(e) => {
                            try {
                              const parsed = JSON.parse(e.target.value);
                              handleVariablesChange(index, parsed);
                            } catch (error) {
                              // Invalid JSON, don't update state
                            }
                          }}
                          disabled={disabled}
                          rows={3}
                          className="block w-full px-3 py-2 border border-border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-background text-foreground text-sm font-mono disabled:bg-muted disabled:text-muted-foreground"
                          placeholder="{}"
                        />
                      </div>
                    </div>
                  </div>

                  {!disabled && (
                    <button
                      onClick={() => handleRemoveContainer(index)}
                      className="p-1 text-muted-foreground hover:text-red-500 transition-colors ml-2"
                      title="Remove container"
                    >
                      <TrashIcon className="h-4 w-4" />
                    </button>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Add Container Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-card rounded-lg border border-border p-6 max-w-2xl w-full mx-4 max-h-[80vh] flex flex-col">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-medium text-foreground">Add Container</h3>
              <button
                onClick={() => setShowAddModal(false)}
                className="text-muted-foreground hover:text-foreground"
              >
                ✕
              </button>
            </div>

            {/* Search */}
            <div className="relative mb-4">
              <MagnifyingGlassIcon className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <input
                type="text"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                placeholder="Search containers..."
                className="w-full pl-10 pr-3 py-2 border border-border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 bg-background text-foreground"
              />
            </div>

            {/* Container List */}
            <div className="flex-1 overflow-y-auto">
              {containersLoading ? (
                <div className="space-y-2">
                  {[...Array(3)].map((_, i) => (
                    <div key={i} className="h-16 bg-muted rounded animate-pulse" />
                  ))}
                </div>
              ) : containersError ? (
                <div className="text-center py-8">
                  <p className="text-red-600">{containersError}</p>
                </div>
              ) : containers.length === 0 ? (
                <div className="text-center py-8">
                  <p className="text-muted-foreground">No containers found</p>
                </div>
              ) : (
                <div className="space-y-2">
                  {containers.map((container) => {
                    const isSelected = isContainerAlreadySelected(container.id);
                    return (
                      <div
                        key={container.id}
                        className={`p-3 border rounded-lg cursor-pointer transition-colors ${
                          isSelected
                            ? 'border-green-500 bg-green-50 text-green-800'
                            : 'border-border hover:border-blue-500 hover:bg-blue-50'
                        }`}
                        onClick={() => !isSelected && handleAddContainer(container)}
                      >
                        <div className="flex items-center justify-between">
                          <div className="min-w-0 flex-1">
                            <div className="flex items-center space-x-2">
                              <h4 className="text-sm font-medium truncate">
                                {container.name}
                              </h4>
                              {isSelected && (
                                <CheckIcon className="h-4 w-4 text-green-600" />
                              )}
                            </div>
                            {container.description && (
                              <p className="text-xs text-muted-foreground truncate">
                                {container.description}
                              </p>
                            )}
                            <p className="text-xs text-muted-foreground">
                              by {container.author}
                            </p>
                          </div>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default ContainerSelector;