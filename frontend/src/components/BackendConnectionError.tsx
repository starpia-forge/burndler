import { ReactNode } from 'react';
import {
  ExclamationTriangleIcon,
  ServerIcon,
  ArrowPathIcon,
  CommandLineIcon,
  InformationCircleIcon,
} from '@heroicons/react/24/outline';
import { useBackendConnection } from '../hooks/useBackendConnection';

interface BackendConnectionErrorProps {
  error?: string;
  showRetry?: boolean;
  showDebugInfo?: boolean;
  children?: ReactNode;
}

export function BackendConnectionError({
  error,
  showRetry = true,
  showDebugInfo = false,
  children,
}: BackendConnectionErrorProps) {
  const { connectionStatus, statusMessage, retry, debugInfo } = useBackendConnection();

  const isDevelopment = process.env.NODE_ENV === 'development';

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-6">
        {/* Header */}
        <div className="flex items-center text-orange-600 mb-4">
          <ServerIcon className="h-8 w-8 mr-3" />
          <h2 className="text-xl font-semibold">Backend Connection Issue</h2>
        </div>

        {/* Main Error Message */}
        <div className="mb-4">
          <p className="text-gray-700 mb-2">
            {error || statusMessage || 'Unable to connect to the backend server.'}
          </p>

          {connectionStatus.error?.includes('ECONNREFUSED') && (
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-3 mb-4">
              <div className="flex items-start">
                <InformationCircleIcon className="h-5 w-5 text-blue-600 mt-0.5 mr-2 flex-shrink-0" />
                <div className="text-sm text-blue-800">
                  <p className="font-medium mb-1">Server Not Running</p>
                  <p>The backend server appears to be offline or not yet started.</p>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Development Instructions */}
        {isDevelopment && (
          <div className="bg-gray-50 border border-gray-200 rounded-lg p-4 mb-4">
            <div className="flex items-start mb-3">
              <CommandLineIcon className="h-5 w-5 text-gray-600 mt-0.5 mr-2 flex-shrink-0" />
              <h3 className="font-medium text-gray-900">Development Setup</h3>
            </div>

            <div className="text-sm text-gray-700 space-y-2">
              <p>To start the backend server, run one of these commands:</p>
              <div className="bg-gray-800 text-green-400 rounded p-2 font-mono text-xs">
                <div>make dev-backend</div>
                <div className="text-gray-500"># or</div>
                <div>make dev</div>
              </div>
              <p className="text-xs text-gray-600">
                The backend should be available at <code>http://localhost:8080</code>
              </p>
            </div>
          </div>
        )}

        {/* Production Instructions */}
        {!isDevelopment && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-4">
            <div className="flex items-start">
              <ExclamationTriangleIcon className="h-5 w-5 text-red-600 mt-0.5 mr-2 flex-shrink-0" />
              <div className="text-sm text-red-800">
                <p className="font-medium mb-1">Service Unavailable</p>
                <p>
                  The system is temporarily unavailable. Please try again in a few moments or
                  contact your system administrator.
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Connection Status */}
        <div className="mb-4 text-sm text-gray-600">
          <div className="flex items-center justify-between">
            <span>Status:</span>
            <span
              className={`font-medium ${connectionStatus.isConnected ? 'text-green-600' : 'text-red-600'}`}
            >
              {connectionStatus.isConnected ? 'Connected' : 'Disconnected'}
            </span>
          </div>
          <div className="flex items-center justify-between">
            <span>Last checked:</span>
            <span>{connectionStatus.lastChecked.toLocaleTimeString()}</span>
          </div>
          {connectionStatus.retryCount > 0 && (
            <div className="flex items-center justify-between">
              <span>Retry attempts:</span>
              <span>{connectionStatus.retryCount}</span>
            </div>
          )}
        </div>

        {/* Actions */}
        <div className="flex flex-col space-y-2">
          {showRetry && (
            <button
              onClick={retry}
              className="flex items-center justify-center w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 transition-colors"
            >
              <ArrowPathIcon className="h-4 w-4 mr-2" />
              Try Again
            </button>
          )}

          <button
            onClick={() => window.location.reload()}
            className="w-full bg-gray-100 text-gray-700 py-2 px-4 rounded-md hover:bg-gray-200 transition-colors"
          >
            Refresh Page
          </button>

          {isDevelopment && (
            <button
              onClick={() => {
                console.log('Backend Connection Debug Info:', debugInfo);
                alert('Debug information logged to console (F12)');
              }}
              className="w-full text-sm bg-gray-50 text-gray-600 py-1 px-4 rounded-md hover:bg-gray-100 transition-colors"
            >
              Show Debug Info
            </button>
          )}
        </div>

        {/* Debug Information (Development Only) */}
        {showDebugInfo && isDevelopment && (
          <details className="mt-4 text-xs text-gray-600">
            <summary className="cursor-pointer font-medium mb-2">Debug Information</summary>
            <pre className="bg-gray-100 p-2 rounded overflow-auto text-xs">
              {JSON.stringify(debugInfo, null, 2)}
            </pre>
          </details>
        )}

        {/* Custom Content */}
        {children && <div className="mt-4 pt-4 border-t border-gray-200">{children}</div>}
      </div>
    </div>
  );
}

/**
 * Simple inline error component for use within existing layouts
 */
export function InlineBackendError({ onRetry }: { onRetry?: () => void }) {
  const { statusMessage, retry } = useBackendConnection();

  return (
    <div className="bg-orange-50 border border-orange-200 rounded-lg p-4">
      <div className="flex items-center">
        <ExclamationTriangleIcon className="h-5 w-5 text-orange-600 mr-3" />
        <div className="flex-1">
          <h3 className="font-medium text-orange-800">Backend Connection Issue</h3>
          <p className="text-sm text-orange-700 mt-1">{statusMessage}</p>
        </div>
        <button
          onClick={onRetry || retry}
          className="ml-3 bg-orange-100 text-orange-800 px-3 py-1 rounded text-sm hover:bg-orange-200 transition-colors"
        >
          Retry
        </button>
      </div>
    </div>
  );
}

export default BackendConnectionError;
