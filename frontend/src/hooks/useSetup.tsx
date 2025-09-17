import { createContext, useContext, useState, useEffect, ReactNode, useCallback } from 'react';
import {
  SetupStatus,
  AdminCreateRequest,
  AdminCreateResponse,
  SetupCompleteRequest,
} from '../types/setup';
import { setupService } from '../services/setup';

export interface SetupContextType {
  setupStatus: SetupStatus | null;
  loading: boolean;
  error: string | null;
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

  const refreshStatus = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const status = await setupService.getStatus();
      setSetupStatus(status);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to get setup status');
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
      await setupService.initialize();
      await refreshStatus();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to initialize setup');
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
        const response = await setupService.createAdmin(request);
        await refreshStatus(); // Refresh status after creating admin
        return response;
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to create admin');
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
        await setupService.complete(request);
        await refreshStatus(); // Refresh status after completing setup
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to complete setup');
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
