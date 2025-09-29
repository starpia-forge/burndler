import { useState, useCallback } from 'react';

interface ConfirmationModalState {
  isOpen: boolean;
  title: string;
  message: React.ReactNode;
  confirmLabel?: string;
  cancelLabel?: string;
  variant?: 'danger' | 'warning' | 'info';
  onConfirm?: () => void | Promise<void>;
  isLoading: boolean;
}

interface UseConfirmationModalReturn {
  isOpen: boolean;
  title: string;
  message: React.ReactNode;
  confirmLabel?: string;
  cancelLabel?: string;
  variant?: 'danger' | 'warning' | 'info';
  isLoading: boolean;
  openModal: (config: ConfirmationModalConfig) => void;
  closeModal: () => void;
  handleConfirm: () => Promise<void>;
}

interface ConfirmationModalConfig {
  title: string;
  message: React.ReactNode;
  confirmLabel?: string;
  cancelLabel?: string;
  variant?: 'danger' | 'warning' | 'info';
  onConfirm: () => void | Promise<void>;
}

export const useConfirmationModal = (): UseConfirmationModalReturn => {
  const [state, setState] = useState<ConfirmationModalState>({
    isOpen: false,
    title: '',
    message: '',
    confirmLabel: undefined,
    cancelLabel: undefined,
    variant: 'danger',
    onConfirm: undefined,
    isLoading: false,
  });

  const openModal = useCallback((config: ConfirmationModalConfig) => {
    setState({
      isOpen: true,
      title: config.title,
      message: config.message,
      confirmLabel: config.confirmLabel,
      cancelLabel: config.cancelLabel,
      variant: config.variant || 'danger',
      onConfirm: config.onConfirm,
      isLoading: false,
    });
  }, []);

  const closeModal = useCallback(() => {
    setState((prev) => ({
      ...prev,
      isOpen: false,
      isLoading: false,
    }));
  }, []);

  const handleConfirm = useCallback(async () => {
    if (state.onConfirm && !state.isLoading) {
      setState((prev) => ({ ...prev, isLoading: true }));

      try {
        await state.onConfirm();
        closeModal();
      } catch (error) {
        // If there's an error, keep the modal open and stop loading
        setState((prev) => ({ ...prev, isLoading: false }));
        // Re-throw the error so the calling component can handle it
        throw error;
      }
    }
  }, [state.onConfirm, state.isLoading, closeModal]);

  return {
    isOpen: state.isOpen,
    title: state.title,
    message: state.message,
    confirmLabel: state.confirmLabel,
    cancelLabel: state.cancelLabel,
    variant: state.variant,
    isLoading: state.isLoading,
    openModal,
    closeModal,
    handleConfirm,
  };
};
