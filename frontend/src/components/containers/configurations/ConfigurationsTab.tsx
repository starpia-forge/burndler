import { useState } from 'react';
import { useContainerConfigurations } from '../../../hooks/useContainerConfigurations';
import {
  ContainerConfiguration,
  CreateContainerConfigurationRequest,
} from '../../../services/configurationService';
import { ContainerVersion } from '../../../types/container';
import { ConfigurationListPanel } from './ConfigurationListPanel';
import { VersionAssignmentPanel } from './VersionAssignmentPanel';
import { ConfigurationModal } from './ConfigurationModal';
import { ConfigurationForm } from './ConfigurationForm';
import { useConfirmationModal } from '../../../hooks/useConfirmationModal';
import ConfirmationModal from '../../common/ConfirmationModal';
import containerService from '../../../services/containerService';

interface ConfigurationsTabProps {
  containerId: string;
  versions: ContainerVersion[];
  onVersionUpdate: () => void;
}

type ModalMode = 'create' | 'edit' | null;

export function ConfigurationsTab({
  containerId,
  versions,
  onVersionUpdate,
}: ConfigurationsTabProps) {
  const [selectedConfig, setSelectedConfig] = useState<ContainerConfiguration | null>(null);
  const [modalMode, setModalMode] = useState<ModalMode>(null);
  const [editingConfig, setEditingConfig] = useState<ContainerConfiguration | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [toastMessage, setToastMessage] = useState<{
    type: 'success' | 'error';
    message: string;
  } | null>(null);

  const confirmationModal = useConfirmationModal();

  const {
    configurations,
    loading,
    error,
    refetch,
    createConfig,
    updateConfig,
    deleteConfig,
    getVersionsUsingConfig,
  } = useContainerConfigurations({
    containerId,
    autoFetch: true,
  });

  const showToast = (type: 'success' | 'error', message: string) => {
    setToastMessage({ type, message });
    setTimeout(() => setToastMessage(null), 5000);
  };

  const handleCreateConfig = () => {
    setEditingConfig(null);
    setModalMode('create');
  };

  const handleEditConfig = (config: ContainerConfiguration) => {
    setEditingConfig(config);
    setModalMode('edit');
  };

  const handleDeleteConfig = (config: ContainerConfiguration) => {
    const versionsUsing = getVersionsUsingConfig(config.id, versions);

    if (versionsUsing.length > 0) {
      confirmationModal.openModal({
        title: 'Cannot Delete Configuration',
        message: (
          <div>
            <p className="mb-2">
              This configuration is currently in use by the following versions:
            </p>
            <ul className="list-disc list-inside space-y-1 mb-3">
              {versionsUsing.map((v) => (
                <li key={v.id} className="text-sm">
                  {v.version}
                </li>
              ))}
            </ul>
            <p className="text-sm text-gray-600">
              Please unassign this configuration from all versions before deleting it.
            </p>
          </div>
        ),
        variant: 'warning',
        confirmLabel: 'Understood',
        onConfirm: () => {}, // Just close modal
      });
      return;
    }

    confirmationModal.openModal({
      title: 'Delete Configuration',
      message: `Are you sure you want to delete the configuration "${config.name}"? This action cannot be undone.`,
      variant: 'danger',
      confirmLabel: 'Delete',
      onConfirm: async () => {
        try {
          await deleteConfig(config.name);
          showToast('success', `Configuration "${config.name}" deleted successfully`);
          if (selectedConfig?.id === config.id) {
            setSelectedConfig(null);
          }
        } catch (err: any) {
          showToast('error', err.message || 'Failed to delete configuration');
        }
      },
    });
  };

  const handleModalSubmit = async (data: CreateContainerConfigurationRequest) => {
    setIsSubmitting(true);
    try {
      if (modalMode === 'create') {
        const newConfig = await createConfig(data);
        showToast('success', `Configuration "${newConfig.name}" created successfully`);
        setModalMode(null);
      } else if (modalMode === 'edit' && editingConfig) {
        const updatedConfig = await updateConfig(editingConfig.name, {
          minimum_version: data.minimum_version,
          description: data.description,
        });
        showToast('success', `Configuration "${updatedConfig.name}" updated successfully`);
        setModalMode(null);
        if (selectedConfig?.id === updatedConfig.id) {
          setSelectedConfig(updatedConfig);
        }
      }
    } catch (err: any) {
      showToast('error', err.message || 'Failed to save configuration');
      throw err; // Re-throw to keep form state
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleAssignConfig = async (versionId: number, configId: number) => {
    const version = versions.find((v) => v.id === versionId);
    if (!version) return;

    try {
      await containerService.updateVersion(parseInt(containerId), version.version, {
        configuration_id: configId,
      });
      showToast('success', `Configuration assigned to ${version.version}`);
      onVersionUpdate();
    } catch (err: any) {
      showToast('error', err.message || 'Failed to assign configuration');
    }
  };

  const handleUnassignConfig = async (versionId: number) => {
    const version = versions.find((v) => v.id === versionId);
    if (!version) return;

    confirmationModal.openModal({
      title: 'Unassign Configuration',
      message: `Are you sure you want to unassign the configuration from version ${version.version}?`,
      variant: 'warning',
      confirmLabel: 'Unassign',
      onConfirm: async () => {
        try {
          await containerService.updateVersion(parseInt(containerId), version.version, {
            configuration_id: null,
          });
          showToast('success', `Configuration unassigned from ${version.version}`);
          onVersionUpdate();
        } catch (err: any) {
          showToast('error', err.message || 'Failed to unassign configuration');
        }
      },
    });
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
    <>
      <div className="space-y-6">
        {/* Toast Notification */}
        {toastMessage && (
          <div
            className={`fixed top-4 right-4 z-50 px-6 py-4 rounded-lg shadow-lg border-l-4 animate-slide-in ${
              toastMessage.type === 'success'
                ? 'bg-green-50 border-green-500 text-green-800'
                : 'bg-red-50 border-red-500 text-red-800'
            }`}
          >
            <div className="flex items-center gap-3">
              {toastMessage.type === 'success' ? (
                <svg className="h-5 w-5 text-green-500" fill="currentColor" viewBox="0 0 20 20">
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                    clipRule="evenodd"
                  />
                </svg>
              ) : (
                <svg className="h-5 w-5 text-red-500" fill="currentColor" viewBox="0 0 20 20">
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                    clipRule="evenodd"
                  />
                </svg>
              )}
              <p className="font-medium">{toastMessage.message}</p>
            </div>
          </div>
        )}

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

      {/* Configuration Modal */}
      <ConfigurationModal
        isOpen={modalMode !== null}
        onClose={() => {
          setModalMode(null);
          setEditingConfig(null);
        }}
        title={modalMode === 'create' ? 'Create Configuration' : 'Edit Configuration'}
      >
        <ConfigurationForm
          mode={modalMode || 'create'}
          initialData={editingConfig || undefined}
          existingNames={configurations.map((c) => c.name)}
          onSubmit={handleModalSubmit}
          onCancel={() => {
            setModalMode(null);
            setEditingConfig(null);
          }}
          isSubmitting={isSubmitting}
        />
      </ConfigurationModal>

      {/* Confirmation Modal */}
      <ConfirmationModal
        isOpen={confirmationModal.isOpen}
        onClose={confirmationModal.closeModal}
        onConfirm={confirmationModal.handleConfirm}
        title={confirmationModal.title}
        message={confirmationModal.message}
        confirmLabel={confirmationModal.confirmLabel}
        cancelLabel={confirmationModal.cancelLabel}
        variant={confirmationModal.variant}
        isLoading={confirmationModal.isLoading}
      />
    </>
  );
}
