import { NavLink } from 'react-router-dom';
import {
  HomeIcon,
  CubeIcon,
  Squares2X2Icon,
  DocumentCheckIcon,
  ClockIcon,
  Cog6ToothIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  CommandLineIcon,
  ShieldCheckIcon,
  RectangleStackIcon,
} from '@heroicons/react/24/outline';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../hooks/useAuth';

interface NavItem {
  translationKey: string;
  href: string;
  icon: React.ComponentType<any>;
  requiredRole?: 'Developer' | 'Engineer';
}

const navigation: NavItem[] = [
  { translationKey: 'navigation.dashboard', href: '/', icon: HomeIcon },
  { translationKey: 'navigation.containers', href: '/containers', icon: Squares2X2Icon },
  { translationKey: 'navigation.services', href: '/services', icon: RectangleStackIcon },
  {
    translationKey: 'navigation.packageBuilder',
    href: '/package',
    icon: CubeIcon,
    requiredRole: 'Developer',
  },
  { translationKey: 'navigation.lintReports', href: '/lint', icon: DocumentCheckIcon },
  { translationKey: 'navigation.buildHistory', href: '/history', icon: ClockIcon },
  { translationKey: 'navigation.cliTools', href: '/cli', icon: CommandLineIcon },
  {
    translationKey: 'navigation.rbacManager',
    href: '/rbac',
    icon: ShieldCheckIcon,
    requiredRole: 'Developer',
  },
  { translationKey: 'navigation.settings', href: '/settings', icon: Cog6ToothIcon },
];

export default function Sidebar() {
  const [collapsed, setCollapsed] = useState(false);
  const { user, isDeveloper } = useAuth();
  const { t } = useTranslation(['common']);

  const filteredNavigation = navigation.filter((item) => {
    if (!item.requiredRole) return true;
    if (item.requiredRole === 'Developer' && !isDeveloper) return false;
    return true;
  });

  return (
    <aside
      className={`${
        collapsed ? 'w-20' : 'w-64'
      } fixed left-0 top-16 bottom-0 z-30 flex flex-col bg-card border-r border-border transition-all duration-300 ease-in-out`}
    >
      {/* Collapse Toggle */}
      <button
        onClick={() => setCollapsed(!collapsed)}
        className="absolute -right-3 top-6 bg-card text-foreground p-1 rounded-full border-2 border-border hover:bg-accent transition-colors"
      >
        {collapsed ? (
          <ChevronRightIcon className="w-4 h-4" />
        ) : (
          <ChevronLeftIcon className="w-4 h-4" />
        )}
      </button>

      {/* Navigation */}
      <nav className="flex-1 px-2 py-4 space-y-1 overflow-y-auto">
        {filteredNavigation.map((item) => {
          const translatedName = t(item.translationKey);
          return (
            <NavLink
              key={item.translationKey}
              to={item.href}
              className={({ isActive }) =>
                `group flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-colors ${
                  isActive
                    ? 'bg-gradient-to-r from-blue-600 to-purple-600 text-white'
                    : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                }`
              }
              title={collapsed ? translatedName : undefined}
            >
              <item.icon className={`${collapsed ? 'mx-auto' : 'mr-3'} h-6 w-6 flex-shrink-0`} />
              {!collapsed && <span className="truncate">{translatedName}</span>}
              {!collapsed && item.requiredRole && (
                <span className="ml-auto inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-muted text-muted-foreground">
                  {t(`roles.${item.requiredRole.toLowerCase()}`)}
                </span>
              )}
            </NavLink>
          );
        })}
      </nav>

      {/* User Info Footer */}
      {!collapsed && user && (
        <div className="flex-shrink-0 p-4 border-t border-border">
          <div className="flex items-center">
            <div className="w-10 h-10 bg-gradient-to-br from-blue-400 to-purple-500 rounded-full flex items-center justify-center flex-shrink-0">
              <span className="text-white font-semibold text-sm">
                {user.email?.charAt(0).toUpperCase()}
              </span>
            </div>
            <div className="ml-3 overflow-hidden">
              <p className="text-sm font-medium text-foreground truncate">
                {user.name || user.email}
              </p>
              <p className="text-xs text-muted-foreground">
                {t(`roles.${user.role?.toLowerCase()}`)}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Version */}
      <div className={`p-4 text-center ${collapsed ? 'px-2' : ''}`}>
        <p className="text-xs text-muted-foreground">
          {collapsed ? t('version.short') : `${t('appName')} ${t('version.full')}`}
        </p>
      </div>
    </aside>
  );
}
