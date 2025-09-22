import {
  CubeIcon,
  DocumentDuplicateIcon,
  DocumentCheckIcon,
  ChartBarIcon,
  ServerIcon,
  CheckCircleIcon,
  XCircleIcon,
} from '@heroicons/react/24/outline';
import { useAuth } from '../hooks/useAuth';
import { Link } from 'react-router-dom';

export default function Dashboard() {
  const { user, isDeveloper } = useAuth();

  const stats = [
    { name: 'Total Merges', value: '24', icon: DocumentDuplicateIcon, color: 'bg-primary-500' },
    { name: 'Packages Built', value: '12', icon: CubeIcon, color: 'bg-success' },
    { name: 'Lint Checks', value: '156', icon: DocumentCheckIcon, color: 'bg-warning' },
    { name: 'Deploy Ready', value: '8', icon: ServerIcon, color: 'bg-info' },
  ];

  const recentActivity = [
    {
      id: 1,
      action: 'Merged compose files',
      project: 'web-app',
      status: 'success',
      time: '2 hours ago',
    },
    {
      id: 2,
      action: 'Built offline package',
      project: 'api-service',
      status: 'success',
      time: '4 hours ago',
    },
    {
      id: 3,
      action: 'Lint validation failed',
      project: 'database',
      status: 'error',
      time: '5 hours ago',
    },
    {
      id: 4,
      action: 'Merged compose files',
      project: 'monitoring',
      status: 'success',
      time: '1 day ago',
    },
    {
      id: 5,
      action: 'Built offline package',
      project: 'auth-service',
      status: 'success',
      time: '2 days ago',
    },
  ];

  const quickActions = [
    {
      name: 'Merge Compose Files',
      href: '/merge',
      icon: DocumentDuplicateIcon,
      description: 'Combine multiple Docker Compose files',
    },
    {
      name: 'Build Package',
      href: '/package',
      icon: CubeIcon,
      description: 'Create offline deployment package',
      requireDeveloper: true,
    },
    {
      name: 'View Lint Reports',
      href: '/lint',
      icon: DocumentCheckIcon,
      description: 'Check validation and lint results',
    },
    {
      name: 'CLI Tools',
      href: '/cli',
      icon: ServerIcon,
      description: 'Access command-line utilities',
    },
  ];

  return (
    <div className="space-y-6">
      {/* Welcome Header */}
      <div className="bg-gradient-to-r from-primary-600 to-primary-700 rounded-2xl p-8 text-white">
        <h1 className="text-3xl font-bold">Welcome back, {user?.name || user?.email || 'User'}!</h1>
        <p className="mt-2 text-primary-100">
          Manage your Docker Compose orchestration and offline deployments from one place.
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat) => (
          <div
            key={stat.name}
            className="bg-card overflow-hidden rounded-lg shadow hover:shadow-lg transition-shadow border border-border"
          >
            <div className="p-5">
              <div className="flex items-center">
                <div className={`${stat.color} rounded-lg p-3`}>
                  <stat.icon className="h-6 w-6 text-white" />
                </div>
                <div className="ml-5 w-0 flex-1">
                  <dl>
                    <dt className="text-sm font-medium text-muted-foreground truncate">
                      {stat.name}
                    </dt>
                    <dd className="text-2xl font-bold text-foreground">{stat.value}</dd>
                  </dl>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Quick Actions */}
        <div className="bg-card rounded-lg shadow border border-border">
          <div className="px-6 py-4 border-b border-border">
            <h2 className="text-lg font-semibold text-foreground">Quick Actions</h2>
          </div>
          <div className="p-6 space-y-4">
            {quickActions.map((action) => {
              if (action.requireDeveloper && !isDeveloper) return null;
              return (
                <Link
                  key={action.name}
                  to={action.href}
                  className="flex items-center p-4 bg-muted/50 rounded-lg hover:bg-muted transition-colors group"
                >
                  <div className="flex-shrink-0">
                    <action.icon className="h-8 w-8 text-muted-foreground group-hover:text-primary-600" />
                  </div>
                  <div className="ml-4 flex-1">
                    <p className="text-sm font-medium text-foreground">{action.name}</p>
                    <p className="text-sm text-muted-foreground">{action.description}</p>
                  </div>
                  <ChartBarIcon className="h-5 w-5 text-muted-foreground" />
                </Link>
              );
            })}
          </div>
        </div>

        {/* Recent Activity */}
        <div className="bg-card rounded-lg shadow border border-border">
          <div className="px-6 py-4 border-b border-border">
            <h2 className="text-lg font-semibold text-foreground">Recent Activity</h2>
          </div>
          <div className="p-6">
            <div className="flow-root">
              <ul className="-mb-8">
                {recentActivity.map((activity, idx) => (
                  <li key={activity.id}>
                    <div className="relative pb-8">
                      {idx !== recentActivity.length - 1 && (
                        <span
                          className="absolute top-4 left-4 -ml-px h-full w-0.5 bg-border"
                          aria-hidden="true"
                        />
                      )}
                      <div className="relative flex space-x-3">
                        <div>
                          {activity.status === 'success' ? (
                            <CheckCircleIcon className="h-8 w-8 text-success" />
                          ) : (
                            <XCircleIcon className="h-8 w-8 text-destructive" />
                          )}
                        </div>
                        <div className="flex-1 min-w-0">
                          <div>
                            <p className="text-sm text-foreground">
                              {activity.action}{' '}
                              <span className="font-medium text-foreground">
                                {activity.project}
                              </span>
                            </p>
                            <p className="text-sm text-muted-foreground">{activity.time}</p>
                          </div>
                        </div>
                      </div>
                    </div>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>
      </div>

      {/* System Status */}
      <div className="bg-card rounded-lg shadow border border-border">
        <div className="px-6 py-4 border-b border-border">
          <h2 className="text-lg font-semibold text-foreground">System Status</h2>
        </div>
        <div className="p-6">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="flex items-center p-4 bg-success/10 rounded-lg border border-success/20">
              <CheckCircleIcon className="h-8 w-8 text-success" />
              <div className="ml-3">
                <p className="text-sm font-medium text-foreground">API Service</p>
                <p className="text-sm text-muted-foreground">Operational</p>
              </div>
            </div>
            <div className="flex items-center p-4 bg-success/10 rounded-lg border border-success/20">
              <CheckCircleIcon className="h-8 w-8 text-success" />
              <div className="ml-3">
                <p className="text-sm font-medium text-foreground">Database</p>
                <p className="text-sm text-muted-foreground">Connected</p>
              </div>
            </div>
            <div className="flex items-center p-4 bg-success/10 rounded-lg border border-success/20">
              <CheckCircleIcon className="h-8 w-8 text-success" />
              <div className="ml-3">
                <p className="text-sm font-medium text-foreground">Storage</p>
                <p className="text-sm text-muted-foreground">Available</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
