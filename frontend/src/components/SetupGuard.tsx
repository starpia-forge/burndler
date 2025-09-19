import { ReactNode } from 'react';
import { Navigate } from 'react-router-dom';
import { useSetup } from '../hooks/useSetup';
import { ErrorType } from '../services/setup';
import { BackendConnectionError } from './BackendConnectionError';

interface SetupGuardProps {
  children: ReactNode;
}

export function SetupGuard({ children }: SetupGuardProps) {
  const {
    isSetupCompleted,
    isSetupRequired,
    loading,
    error,
    errorType,
    isBackendConnected,
    isBackendDown,
  } = useSetup();

  // Show backend connection error if backend is down
  if (errorType === ErrorType.BACKEND_DOWN || isBackendDown || !isBackendConnected) {
    return (
      <BackendConnectionError error={error || undefined} showRetry={true} showDebugInfo={false} />
    );
  }

  // Show loading spinner while checking setup status
  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Checking system status...</p>
        </div>
      </div>
    );
  }

  // Show network or parse errors
  if (error && (errorType === ErrorType.NETWORK_ERROR || errorType === ErrorType.PARSE_ERROR)) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-6">
          <div className="flex items-center text-red-600 mb-4">
            <svg className="h-8 w-8 mr-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.728-.833-2.498 0L4.316 18.5c-.77.833.192 2.5 1.732 2.5z"
              />
            </svg>
            <h2 className="text-xl font-semibold">System Error</h2>
          </div>
          <p className="text-gray-700 mb-4">{error}</p>
          <button
            onClick={() => window.location.reload()}
            className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  // If setup is required and not completed, redirect to setup
  if (isSetupRequired && !isSetupCompleted) {
    return <Navigate to="/setup" replace />;
  }

  // If setup is completed, render the children
  return <>{children}</>;
}

export default SetupGuard;
