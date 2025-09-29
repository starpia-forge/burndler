import { createContext, useContext, useState, useEffect, ReactNode, useCallback } from 'react';
import {
  SetupStatus,
  AdminCreateRequest,
  AdminCreateResponse,
  SetupCompleteRequest,
} from '../types/setup';
import { setupService, ApiError, ErrorType } from '../services/setup';
import { useBackendConnection } from './useBackendConnection';

export interface SetupContextType {
  setupStatus: SetupStatus | null;
  loading: boolean;
  error: string | null;
  errorType: ErrorType | null;
  isBackendConnected: boolean;
  isBackendDown: boolean;
  isSetupCompleted: boolean;
  isSetupRequired: boolean;
  hasAdmin: boolean;
  setupToken: string | null;
  refreshStatus: () => Promise<void>;
  initialize: () => Promise<void>;
  createAdmin: (request: AdminCreateRequest) => Promise<AdminCreateResponse>;
  completeSetup: (request: SetupCompleteRequest) => Promise<void>;
}

const SetupContext = createContext<SetupContextType | undefined>(undefined);

export function SetupProvider({ children }: { children: ReactNode }) {
  const [setupStatus, setSetupStatus] = useState<SetupStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [errorType, setErrorType] = useState<ErrorType | null>(null);

  // Get backend connection status
  const { isConnected: isBackendConnected, isBackendDown } = useBackendConnection();

  const refreshStatus = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      setErrorType(null);
      const status = await setupService.getStatus();
      setSetupStatus(status);
    } catch (err) {
      const apiError = err as ApiError;

      // Set error type for better error handling
      if (apiError.type) {
        setErrorType(apiError.type);
      }

      // Provide context-specific error messages
      let errorMessage = 'Failed to get setup status';

      if (apiError.type === ErrorType.BACKEND_DOWN) {
        errorMessage = 'Backend server is not running';
      } else if (apiError.type === ErrorType.NETWORK_ERROR) {
        errorMessage = 'Network connection failed';
      } else if (apiError.type === ErrorType.PARSE_ERROR) {
        errorMessage = 'Invalid response from backend';
      } else if (apiError.message) {
        errorMessage = apiError.message;
      }

      setError(errorMessage);
      console.error('Setup status error:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    // Check setup status on mount
    refreshStatus();
  }, [refreshStatus]);

  const initialize = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      setErrorType(null);
      await setupService.initialize();
      await refreshStatus();
    } catch (err) {
      const apiError = err as ApiError;

      if (apiError.type) {
        setErrorType(apiError.type);
      }

      let errorMessage = 'Failed to initialize setup';
      if (apiError.type === ErrorType.BACKEND_DOWN) {
        errorMessage = 'Backend server is not running';
      } else if (apiError.message) {
        errorMessage = apiError.message;
      }

      setError(errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  }, [refreshStatus]);

  const createAdmin = useCallback(
    async (request: AdminCreateRequest): Promise<AdminCreateResponse> => {
      try {
        setLoading(true);
        setError(null);
        setErrorType(null);
        const response = await setupService.createAdmin(request);
        await refreshStatus(); // Refresh status after creating admin
        return response;
      } catch (err) {
        const apiError = err as ApiError;

        if (apiError.type) {
          setErrorType(apiError.type);
        }

        let errorMessage = 'Failed to create admin';
        if (apiError.type === ErrorType.BACKEND_DOWN) {
          errorMessage = 'Backend server is not running';
        } else if (apiError.message) {
          errorMessage = apiError.message;
        }

        setError(errorMessage);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [refreshStatus]
  );

  const completeSetup = useCallback(
    async (request: SetupCompleteRequest) => {
      try {
        setLoading(true);
        setError(null);
        setErrorType(null);
        await setupService.complete(request);
        await refreshStatus(); // Refresh status after completing setup
      } catch (err) {
        const apiError = err as ApiError;

        if (apiError.type) {
          setErrorType(apiError.type);
        }

        let errorMessage = 'Failed to complete setup';
        if (apiError.type === ErrorType.BACKEND_DOWN) {
          errorMessage = 'Backend server is not running';
        } else if (apiError.message) {
          errorMessage = apiError.message;
        }

        setError(errorMessage);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [refreshStatus]
  );

  const value: SetupContextType = {
    setupStatus,
    loading,
    error,
    errorType,
    isBackendConnected,
    isBackendDown,
    isSetupCompleted: setupStatus?.is_completed ?? false,
    isSetupRequired: setupStatus?.requires_setup ?? true,
    hasAdmin: setupStatus?.admin_exists ?? false,
    setupToken: setupStatus?.setup_token ?? null,
    refreshStatus,
    initialize,
    createAdmin,
    completeSetup,
  };

  return <SetupContext.Provider value={value}>{children}</SetupContext.Provider>;
}

export function useSetup(): SetupContextType {
  const context = useContext(SetupContext);
  if (context === undefined) {
    throw new Error('useSetup must be used within a SetupProvider');
  }
  return context;
}
