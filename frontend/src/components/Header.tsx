import { Fragment } from 'react';
import { Menu, Transition } from '@headlessui/react';
import {
  UserCircleIcon,
  ArrowRightOnRectangleIcon,
  Cog6ToothIcon,
  UserIcon,
} from '@heroicons/react/24/outline';
import { ChevronDownIcon } from '@heroicons/react/20/solid';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../hooks/useAuth';
import ThemeToggle from './ThemeToggle';
import LanguageSelector from './LanguageSelector';

export default function Header() {
  const { user, isAuthenticated, logout } = useAuth();
  const { t } = useTranslation(['common', 'auth']);

  return (
    <header className="fixed top-0 left-0 right-0 z-40 bg-background border-b border-border shadow-sm">
      <div className="px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          {/* Logo and Project Name */}
          <div className="flex items-center">
            <div className="flex-shrink-0 flex items-center">
              <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
                <span className="text-white font-bold text-lg">B</span>
              </div>
              <h1 className="ml-3 text-xl font-semibold text-foreground">{t('common:appName')}</h1>
            </div>
          </div>

          {/* User Menu */}
          <div className="flex items-center space-x-4">
            <LanguageSelector />
            <ThemeToggle />
            {isAuthenticated ? (
              <Menu as="div" className="relative">
                <Menu.Button className="flex items-center text-sm rounded-full focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500">
                  <div className="flex items-center space-x-3 p-2 rounded-lg hover:bg-accent transition-colors">
                    <div className="text-right hidden sm:block">
                      <p className="text-sm font-medium text-foreground">
                        {user?.name || user?.email}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {user?.role && t(`auth:role.${user.role.toLowerCase()}`)}
                      </p>
                    </div>
                    <div className="w-10 h-10 bg-gradient-to-br from-blue-400 to-purple-500 rounded-full flex items-center justify-center">
                      <UserIcon className="w-6 h-6 text-white" />
                    </div>
                    <ChevronDownIcon className="w-5 h-5 text-muted-foreground" />
                  </div>
                </Menu.Button>

                <Transition
                  as={Fragment}
                  enter="transition ease-out duration-100"
                  enterFrom="transform opacity-0 scale-95"
                  enterTo="transform opacity-100 scale-100"
                  leave="transition ease-in duration-75"
                  leaveFrom="transform opacity-100 scale-100"
                  leaveTo="transform opacity-0 scale-95"
                >
                  <Menu.Items className="absolute right-0 mt-2 w-56 rounded-lg bg-popover border border-border shadow-lg focus:outline-none">
                    <div className="p-1">
                      <div className="px-3 py-2 border-b border-border">
                        <p className="text-sm font-medium text-foreground">{user?.email}</p>
                        <p className="text-xs text-muted-foreground mt-1">
                          <span
                            className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${
                              user?.role === 'Developer'
                                ? 'bg-green-100 text-green-800'
                                : 'bg-blue-100 text-blue-800'
                            }`}
                          >
                            {user?.role && t(`auth:role.${user.role.toLowerCase()}`)}
                          </span>
                        </p>
                      </div>

                      <Menu.Item>
                        {({ active }) => (
                          <button
                            className={`${
                              active ? 'bg-accent' : ''
                            } group flex w-full items-center rounded-md px-3 py-2 text-sm text-foreground`}
                          >
                            <UserCircleIcon className="mr-3 h-5 w-5 text-muted-foreground group-hover:text-foreground" />
                            {t('auth:profile')}
                          </button>
                        )}
                      </Menu.Item>

                      <Menu.Item>
                        {({ active }) => (
                          <button
                            className={`${
                              active ? 'bg-accent' : ''
                            } group flex w-full items-center rounded-md px-3 py-2 text-sm text-foreground`}
                          >
                            <Cog6ToothIcon className="mr-3 h-5 w-5 text-muted-foreground group-hover:text-foreground" />
                            {t('auth:settings')}
                          </button>
                        )}
                      </Menu.Item>

                      <div className="border-t border-border mt-1 pt-1">
                        <Menu.Item>
                          {({ active }) => (
                            <button
                              onClick={logout}
                              className={`${
                                active ? 'bg-accent' : ''
                              } group flex w-full items-center rounded-md px-3 py-2 text-sm text-foreground`}
                            >
                              <ArrowRightOnRectangleIcon className="mr-3 h-5 w-5 text-muted-foreground group-hover:text-foreground" />
                              {t('auth:signOut')}
                            </button>
                          )}
                        </Menu.Item>
                      </div>
                    </div>
                  </Menu.Items>
                </Transition>
              </Menu>
            ) : (
              <button className="flex items-center space-x-2 px-4 py-2 rounded-lg bg-primary-500 text-primary-foreground font-medium hover:bg-primary-600 transition-all">
                <ArrowRightOnRectangleIcon className="w-5 h-5" />
                <span>{t('auth:signIn')}</span>
              </button>
            )}
          </div>
        </div>
      </div>
    </header>
  );
}
