import { useEffect } from 'react';
import { CheckCircleIcon, ArrowRightIcon } from '@heroicons/react/24/outline';
import { useAuth } from '../../hooks/useAuth';
import { useNavigate } from 'react-router-dom';

export default function SetupComplete() {
  const { isAuthenticated } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    // Auto-redirect to dashboard after 5 seconds if user is authenticated
    if (isAuthenticated) {
      const timer = setTimeout(() => {
        navigate('/');
      }, 5000);

      return () => clearTimeout(timer);
    }
  }, [isAuthenticated, navigate]);

  const handleGoToDashboard = () => {
    if (isAuthenticated) {
      navigate('/');
    } else {
      navigate('/login');
    }
  };

  const completedItems = [
    'System database initialized',
    'Admin account created',
    'Company profile configured',
    'System settings applied',
    'Default namespace configured',
    'Security policies activated',
  ];

  return (
    <div className="p-8 text-center">
      <div className="mb-8">
        <CheckCircleIcon className="h-20 w-20 text-green-500 mx-auto mb-4" />
        <h2 className="text-3xl font-bold text-gray-900 mb-2">Setup Complete!</h2>
        <p className="text-lg text-gray-600">
          Your Burndler system has been successfully configured and is ready to use.
        </p>
      </div>

      <div className="max-w-md mx-auto mb-8">
        <div className="bg-green-50 border border-green-200 rounded-lg p-6">
          <h3 className="text-lg font-medium text-green-900 mb-4">What we've configured:</h3>
          <ul className="text-sm text-green-700 space-y-2">
            {completedItems.map((item, index) => (
              <li key={index} className="flex items-center">
                <CheckCircleIcon className="h-4 w-4 text-green-500 mr-2 flex-shrink-0" />
                {item}
              </li>
            ))}
          </ul>
        </div>
      </div>

      <div className="space-y-4">
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <p className="text-sm text-blue-700">
            <strong>Next Steps:</strong>{' '}
            {isAuthenticated
              ? 'You can now access your dashboard and start managing your Docker Compose deployments.'
              : 'Please log in with your admin credentials to access the dashboard.'}
          </p>
        </div>

        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <button
            onClick={handleGoToDashboard}
            className="inline-flex items-center px-6 py-3 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors"
          >
            {isAuthenticated ? 'Go to Dashboard' : 'Go to Login'}
            <ArrowRightIcon className="ml-2 h-5 w-5" />
          </button>
        </div>

        {isAuthenticated && (
          <p className="text-sm text-gray-500">Redirecting automatically in 5 seconds...</p>
        )}
      </div>

      <div className="mt-8 pt-8 border-t border-gray-200">
        <div className="text-sm text-gray-500">
          <p className="mb-2">Need help getting started?</p>
          <div className="flex justify-center space-x-4">
            <a href="#" className="text-blue-600 hover:text-blue-500">
              Documentation
            </a>
            <span>·</span>
            <a href="#" className="text-blue-600 hover:text-blue-500">
              API Guide
            </a>
            <span>·</span>
            <a href="#" className="text-blue-600 hover:text-blue-500">
              Support
            </a>
          </div>
        </div>
      </div>
    </div>
  );
}
