import { createContext, useContext, useState, useEffect, ReactNode, useCallback } from 'react';
import { connectionMonitor, ConnectionStatus } from '../services/connectionMonitor';

export interface BackendConnectionContextType {
  connectionStatus: ConnectionStatus;
  isConnected: boolean;
  isBackendDown: boolean;
  statusMessage: string;
  debugInfo: Record<string, any>;
  retry: () => void;
  startMonitoring: () => void;
  stopMonitoring: () => void;
}

const BackendConnectionContext = createContext<BackendConnectionContextType | undefined>(undefined);

export function BackendConnectionProvider({ children }: { children: ReactNode }) {
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>(
    connectionMonitor.getStatus()
  );

  useEffect(() => {
    // Subscribe to connection status changes
    const unsubscribe = connectionMonitor.subscribe((status) => {
      setConnectionStatus(status);
    });

    // Start monitoring when component mounts
    connectionMonitor.start();

    // Cleanup on unmount
    return () => {
      unsubscribe();
      connectionMonitor.stop();
    };
  }, []);

  const retry = useCallback(() => {
    connectionMonitor.retry();
  }, []);

  const startMonitoring = useCallback(() => {
    connectionMonitor.start();
  }, []);

  const stopMonitoring = useCallback(() => {
    connectionMonitor.stop();
  }, []);

  const getDebugInfo = useCallback(() => {
    return connectionMonitor.getDebugInfo();
  }, []);

  const value: BackendConnectionContextType = {
    connectionStatus,
    isConnected: connectionStatus.isConnected,
    isBackendDown: connectionMonitor.isBackendDown(),
    statusMessage: connectionMonitor.getStatusMessage(),
    debugInfo: getDebugInfo(),
    retry,
    startMonitoring,
    stopMonitoring,
  };

  return (
    <BackendConnectionContext.Provider value={value}>{children}</BackendConnectionContext.Provider>
  );
}

export function useBackendConnection(): BackendConnectionContextType {
  const context = useContext(BackendConnectionContext);
  if (context === undefined) {
    throw new Error('useBackendConnection must be used within a BackendConnectionProvider');
  }
  return context;
}

/**
 * Hook specifically for checking if backend is available before making API calls
 */
export function useBackendAvailability() {
  const { isConnected, isBackendDown, retry } = useBackendConnection();

  const checkAvailability = useCallback(async (): Promise<boolean> => {
    if (!isConnected) {
      // Try one immediate check before giving up
      await connectionMonitor.checkConnection();
      return connectionMonitor.getStatus().isConnected;
    }
    return true;
  }, [isConnected]);

  return {
    isAvailable: isConnected,
    isBackendDown,
    checkAvailability,
    retry,
  };
}

export default useBackendConnection;
