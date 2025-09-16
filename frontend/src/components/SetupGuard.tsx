import { ReactNode } from 'react';
import { Navigate } from 'react-router-dom';
import { useSetup } from '../hooks/useSetup';

interface SetupGuardProps {
  children: ReactNode;
}

export function SetupGuard({ children }: SetupGuardProps) {
  const { isSetupCompleted, isSetupRequired, loading } = useSetup();

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

  // If setup is required and not completed, redirect to setup
  if (isSetupRequired && !isSetupCompleted) {
    return <Navigate to="/setup" replace />;
  }

  // If setup is completed, render the children
  return <>{children}</>;
}

export default SetupGuard;
