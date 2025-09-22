import { ReactNode } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

interface LayoutProps {
  children: ReactNode;
}

export default function Layout({ children }: LayoutProps) {
  const { user, isAuthenticated, isDeveloper, logout } = useAuth();
  const location = useLocation();

  const isActive = (path: string) => location.pathname === path;

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="bg-card shadow-sm border-b border-border">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <h1 className="text-xl font-bold text-foreground">Burndler</h1>
              <nav className="ml-10 flex space-x-4">
                <Link
                  to="/merge"
                  className={`px-3 py-2 rounded-md text-sm font-medium ${
                    isActive('/merge') || isActive('/')
                      ? 'bg-primary-100 text-primary-700'
                      : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                  }`}
                >
                  Compose Merger
                </Link>
                {isDeveloper && (
                  <Link
                    to="/package"
                    className={`px-3 py-2 rounded-md text-sm font-medium ${
                      isActive('/package')
                        ? 'bg-primary-100 text-primary-700'
                        : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                    }`}
                  >
                    Package Builder
                  </Link>
                )}
              </nav>
            </div>

            <div className="flex items-center space-x-4">
              {isAuthenticated ? (
                <>
                  <div className="flex items-center space-x-2">
                    <span className="text-sm text-foreground">{user?.email}</span>
                    <span
                      className={`px-2 py-1 text-xs rounded-full ${
                        isDeveloper
                          ? 'bg-success/10 text-success border border-success/20'
                          : 'bg-info/10 text-info border border-info/20'
                      }`}
                    >
                      {user?.role}
                    </span>
                  </div>
                  <button
                    onClick={logout}
                    className="text-sm text-muted-foreground hover:text-foreground"
                  >
                    Logout
                  </button>
                </>
              ) : (
                <div className="text-sm text-muted-foreground">Demo Mode (Not Authenticated)</div>
              )}
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">{children}</main>

      {/* Footer */}
      <footer className="bg-card border-t border-border mt-auto">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="text-center text-sm text-muted-foreground">
            Burndler - Docker Compose Orchestration for Offline Deployment
          </div>
        </div>
      </footer>
    </div>
  );
}
