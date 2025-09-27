import { NavLink } from 'react-router-dom';
import {
  HomeIcon,
  DocumentDuplicateIcon,
  CubeIcon,
  Squares2X2Icon,
  DocumentCheckIcon,
  ClockIcon,
  Cog6ToothIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  CommandLineIcon,
  ShieldCheckIcon,
} from '@heroicons/react/24/outline';
import { useState } from 'react';
import { useAuth } from '../hooks/useAuth';

interface NavItem {
  name: string;
  href: string;
  icon: React.ComponentType<any>;
  requiredRole?: 'Developer' | 'Engineer';
}

const navigation: NavItem[] = [
  { name: 'Dashboard', href: '/', icon: HomeIcon },
  { name: 'Modules', href: '/modules', icon: Squares2X2Icon },
  { name: 'Compose Merger', href: '/merge', icon: DocumentDuplicateIcon },
  { name: 'Package Builder', href: '/package', icon: CubeIcon, requiredRole: 'Developer' },
  { name: 'Lint Reports', href: '/lint', icon: DocumentCheckIcon },
  { name: 'Build History', href: '/history', icon: ClockIcon },
  { name: 'CLI Tools', href: '/cli', icon: CommandLineIcon },
  { name: 'RBAC Manager', href: '/rbac', icon: ShieldCheckIcon, requiredRole: 'Developer' },
  { name: 'Settings', href: '/settings', icon: Cog6ToothIcon },
];

export default function Sidebar() {
  const [collapsed, setCollapsed] = useState(false);
  const { user, isDeveloper } = useAuth();

  const filteredNavigation = navigation.filter((item) => {
    if (!item.requiredRole) return true;
    if (item.requiredRole === 'Developer' && !isDeveloper) return false;
    return true;
  });

  return (
    <aside
      className={`${
        collapsed ? 'w-20' : 'w-64'
      } fixed left-0 top-16 bottom-0 z-30 flex flex-col bg-gray-900 transition-all duration-300 ease-in-out`}
    >
      {/* Collapse Toggle */}
      <button
        onClick={() => setCollapsed(!collapsed)}
        className="absolute -right-3 top-6 bg-gray-900 text-white p-1 rounded-full border-2 border-gray-700 hover:bg-gray-800 transition-colors"
      >
        {collapsed ? (
          <ChevronRightIcon className="w-4 h-4" />
        ) : (
          <ChevronLeftIcon className="w-4 h-4" />
        )}
      </button>

      {/* Navigation */}
      <nav className="flex-1 px-2 py-4 space-y-1 overflow-y-auto">
        {filteredNavigation.map((item) => (
          <NavLink
            key={item.name}
            to={item.href}
            className={({ isActive }) =>
              `group flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-colors ${
                isActive
                  ? 'bg-gradient-to-r from-blue-600 to-purple-600 text-white'
                  : 'text-gray-300 hover:bg-gray-800 hover:text-white'
              }`
            }
            title={collapsed ? item.name : undefined}
          >
            <item.icon className={`${collapsed ? 'mx-auto' : 'mr-3'} h-6 w-6 flex-shrink-0`} />
            {!collapsed && <span className="truncate">{item.name}</span>}
            {!collapsed && item.requiredRole && (
              <span className="ml-auto inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-700 text-gray-300">
                {item.requiredRole}
              </span>
            )}
          </NavLink>
        ))}
      </nav>

      {/* User Info Footer */}
      {!collapsed && user && (
        <div className="flex-shrink-0 p-4 border-t border-gray-800">
          <div className="flex items-center">
            <div className="w-10 h-10 bg-gradient-to-br from-blue-400 to-purple-500 rounded-full flex items-center justify-center flex-shrink-0">
              <span className="text-white font-semibold text-sm">
                {user.email?.charAt(0).toUpperCase()}
              </span>
            </div>
            <div className="ml-3 overflow-hidden">
              <p className="text-sm font-medium text-white truncate">{user.name || user.email}</p>
              <p className="text-xs text-gray-400">{user.role}</p>
            </div>
          </div>
        </div>
      )}

      {/* Version */}
      <div className={`p-4 text-center ${collapsed ? 'px-2' : ''}`}>
        <p className="text-xs text-gray-500">{collapsed ? 'v0.1' : 'Burndler v0.1.0'}</p>
      </div>
    </aside>
  );
}
