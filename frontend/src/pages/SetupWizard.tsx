import { useState, useEffect } from 'react';
import { useSetup } from '../hooks/useSetup';
import { useAuth } from '../hooks/useAuth';
import { Navigate } from 'react-router-dom';
import { ErrorType } from '../services/setup';
import { BackendConnectionError } from '../components/BackendConnectionError';
import {
  CheckCircleIcon,
  ExclamationTriangleIcon,
  CogIcon,
  UserPlusIcon,
  BuildingOfficeIcon,
  LanguageIcon,
} from '@heroicons/react/24/outline';
import SetupStatus from '../components/setup/SetupStatus';
import SystemLanguage from '../components/setup/SystemLanguage';
import AdminSetup from '../components/setup/AdminSetup';
import SystemConfig from '../components/setup/SystemConfig';
import SetupComplete from '../components/setup/SetupComplete';
import { SetupWizardProvider } from '../contexts/SetupWizardContext';

type SetupStep = 'status' | 'language' | 'admin' | 'config' | 'complete';

export default function SetupWizard() {
  const {
    setupStatus,
    loading,
    error,
    errorType,
    isBackendConnected,
    isBackendDown,
    isSetupCompleted,
    isSetupRequired,
  } = useSetup();
  const { isAuthenticated } = useAuth();
  const [currentStep, setCurrentStep] = useState<SetupStep>('status');

  useEffect(() => {
    if (!loading && setupStatus) {
      if (setupStatus.is_completed) {
        setCurrentStep('complete');
      }
    }
  }, [setupStatus, loading]);

  // If setup is completed and user is authenticated, redirect to dashboard
  if (isSetupCompleted && isAuthenticated) {
    return <Navigate to="/" replace />;
  }

  // Show backend connection error if backend is down
  if (errorType === ErrorType.BACKEND_DOWN || isBackendDown || !isBackendConnected) {
    return (
      <BackendConnectionError
        error={error || undefined}
        showRetry={true}
        showDebugInfo={process.env.NODE_ENV === 'development'}
      >
        <div className="text-center">
          <p className="text-sm text-gray-600 mb-2">
            The setup wizard requires a connection to the backend server.
          </p>
          <p className="text-xs text-gray-500">
            Once the server is running, the setup process will continue automatically.
          </p>
        </div>
      </BackendConnectionError>
    );
  }

  // If setup is not required, redirect to login
  if (!isSetupRequired && !loading) {
    return <Navigate to="/login" replace />;
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading setup status...</p>
        </div>
      </div>
    );
  }

  // Show other types of errors with improved messaging
  if (error) {
    let errorTitle = 'Setup Error';
    let errorDetails = error;

    if (errorType === ErrorType.NETWORK_ERROR) {
      errorTitle = 'Network Error';
      errorDetails =
        'Unable to connect to the backend server. Please check your connection and try again.';
    } else if (errorType === ErrorType.PARSE_ERROR) {
      errorTitle = 'Communication Error';
      errorDetails =
        'Received invalid response from backend server. This may indicate a server configuration issue.';
    } else if (errorType === ErrorType.API_ERROR) {
      errorTitle = 'API Error';
    }

    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-6">
          <div className="flex items-center text-red-600 mb-4">
            <ExclamationTriangleIcon className="h-8 w-8 mr-3" />
            <h2 className="text-xl font-semibold">{errorTitle}</h2>
          </div>
          <p className="text-gray-700 mb-4">{errorDetails}</p>

          {process.env.NODE_ENV === 'development' && (
            <div className="bg-gray-50 border border-gray-200 rounded p-3 mb-4">
              <p className="text-xs text-gray-600">
                <strong>Debug Info:</strong> Error Type: {errorType || 'Unknown'}
              </p>
            </div>
          )}

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

  const steps = [
    {
      id: 'status',
      name: 'System Check',
      icon: CheckCircleIcon,
      completed: currentStep !== 'status',
    },
    {
      id: 'language',
      name: 'System Language',
      icon: LanguageIcon,
      completed: ['config', 'admin', 'complete'].includes(currentStep),
    },
    {
      id: 'config',
      name: 'System Config',
      icon: CogIcon,
      completed: ['admin', 'complete'].includes(currentStep),
    },
    {
      id: 'admin',
      name: 'Admin Account',
      icon: UserPlusIcon,
      completed: currentStep === 'complete',
    },
    { id: 'complete', name: 'Complete', icon: BuildingOfficeIcon, completed: false },
  ];

  const currentStepIndex = steps.findIndex((step) => step.id === currentStep);

  const renderStepContent = () => {
    switch (currentStep) {
      case 'status':
        return (
          <SetupStatus
            setupStatus={setupStatus!}
            onContinue={() => {
              setCurrentStep('language');
            }}
          />
        );
      case 'language':
        return (
          <SystemLanguage
            onContinue={() => {
              setCurrentStep('config');
            }}
          />
        );
      case 'admin':
        return (
          <AdminSetup
            hasAdmin={setupStatus?.admin_exists ?? false}
            onAdminCreated={() => {
              setCurrentStep('complete');
            }}
          />
        );
      case 'config':
        return (
          <SystemConfig
            onConfigComplete={() => {
              if (setupStatus?.admin_exists) {
                setCurrentStep('complete');
              } else {
                setCurrentStep('admin');
              }
            }}
          />
        );
      case 'complete':
        return <SetupComplete />;
      default:
        return null;
    }
  };

  return (
    <SetupWizardProvider>
      <div className="min-h-screen bg-gray-50">
        {/* Header */}
        <div className="bg-white shadow">
          <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
            <div className="flex items-center">
              <BuildingOfficeIcon className="h-8 w-8 text-blue-600 mr-3" />
              <h1 className="text-2xl font-bold text-gray-900">Burndler Setup</h1>
            </div>
            <p className="mt-2 text-sm text-gray-600">
              Welcome to Burndler! Let's get your system configured.
            </p>
          </div>
        </div>

        {/* Progress Steps */}
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="mb-8">
            <nav aria-label="Progress">
              <ol className="flex items-center">
                {steps.map((step, stepIdx) => (
                  <li
                    key={step.name}
                    className={`relative ${stepIdx !== steps.length - 1 ? 'pr-8 sm:pr-20' : ''}`}
                  >
                    {stepIdx !== steps.length - 1 && (
                      <div className="absolute inset-0 flex items-center" aria-hidden="true">
                        <div
                          className={`h-0.5 w-full ${stepIdx < currentStepIndex ? 'bg-blue-600' : 'bg-gray-200'}`}
                        />
                      </div>
                    )}
                    <div className="relative flex items-center justify-center">
                      <div
                        className={`
                        h-9 w-9 rounded-full flex items-center justify-center
                        ${
                          step.id === currentStep
                            ? 'bg-blue-600 text-white'
                            : step.completed || stepIdx < currentStepIndex
                              ? 'bg-blue-600 text-white'
                              : 'bg-white border-2 border-gray-300 text-gray-500'
                        }
                      `}
                      >
                        <step.icon className="h-5 w-5" />
                      </div>
                      <span
                        className={`
                        ml-2 text-sm font-medium
                        ${
                          step.id === currentStep
                            ? 'text-blue-600'
                            : step.completed || stepIdx < currentStepIndex
                              ? 'text-gray-900'
                              : 'text-gray-500'
                        }
                      `}
                      >
                        {step.name}
                      </span>
                    </div>
                  </li>
                ))}
              </ol>
            </nav>
          </div>

          {/* Step Content */}
          <div className="bg-white rounded-lg shadow-lg overflow-hidden">{renderStepContent()}</div>
        </div>
      </div>
    </SetupWizardProvider>
  );
}
