import {
  CheckCircleIcon,
  ExclamationTriangleIcon,
  InformationCircleIcon,
} from '@heroicons/react/24/outline';
import { SetupStatus as SetupStatusType } from '../../types/setup';

interface SetupStatusProps {
  setupStatus: SetupStatusType;
  onContinue: () => void;
}

export default function SetupStatus({ setupStatus, onContinue }: SetupStatusProps) {
  const statusItems = [
    {
      label: 'System Database',
      status: 'ready',
      message: 'Database connection established',
    },
    {
      label: 'Admin Account',
      status: setupStatus.admin_exists ? 'ready' : 'pending',
      message: setupStatus.admin_exists
        ? 'Admin account exists'
        : 'Admin account needs to be created',
    },
    {
      label: 'Setup Status',
      status: setupStatus.is_completed ? 'complete' : 'pending',
      message: setupStatus.is_completed ? 'System setup is complete' : 'System setup required',
    },
  ];

  return (
    <div className="p-8">
      <div className="text-center mb-8">
        <InformationCircleIcon className="h-16 w-16 text-blue-600 mx-auto mb-4" />
        <h2 className="text-2xl font-bold text-gray-900 mb-2">System Status Check</h2>
        <p className="text-gray-600">
          Let's check your system status before proceeding with the setup.
        </p>
      </div>

      <div className="space-y-4 mb-8">
        {statusItems.map((item, index) => (
          <div key={index} className="flex items-center p-4 bg-gray-50 rounded-lg">
            <div className="flex-shrink-0">
              {item.status === 'ready' || item.status === 'complete' ? (
                <CheckCircleIcon className="h-6 w-6 text-green-500" />
              ) : (
                <ExclamationTriangleIcon className="h-6 w-6 text-yellow-500" />
              )}
            </div>
            <div className="ml-4 flex-1">
              <div className="flex justify-between items-center">
                <h3 className="text-sm font-medium text-gray-900">{item.label}</h3>
                <span
                  className={`
                  px-2 py-1 text-xs font-medium rounded-full
                  ${
                    item.status === 'ready' || item.status === 'complete'
                      ? 'bg-green-100 text-green-800'
                      : 'bg-yellow-100 text-yellow-800'
                  }
                `}
                >
                  {item.status === 'complete'
                    ? 'Complete'
                    : item.status === 'ready'
                      ? 'Ready'
                      : 'Pending'}
                </span>
              </div>
              <p className="text-sm text-gray-500 mt-1">{item.message}</p>
            </div>
          </div>
        ))}
      </div>

      {setupStatus.requires_setup && (
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
          <div className="flex">
            <InformationCircleIcon className="h-5 w-5 text-blue-400 flex-shrink-0" />
            <div className="ml-3">
              <h3 className="text-sm font-medium text-blue-800">Setup Required</h3>
              <div className="text-sm text-blue-700 mt-1">
                <p>
                  Your system needs to be configured before you can use Burndler. This process will:
                </p>
                <ul className="list-disc list-inside mt-2 space-y-1">
                  <li>Create an initial admin account</li>
                  <li>Configure system settings</li>
                  <li>Set up your organization profile</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      )}

      <div className="flex justify-end">
        <button
          onClick={onContinue}
          className="bg-blue-600 text-white px-6 py-2 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors"
        >
          Continue Setup
        </button>
      </div>
    </div>
  );
}
