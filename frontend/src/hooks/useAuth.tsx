import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { User, AuthContextType } from '../types/auth';
import { authService } from '../services/auth';

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check for existing token on mount
    const initAuth = async () => {
      try {
        const token = authService.getAccessToken();
        if (token) {
          // Validate token by making an authenticated request
          try {
            const response = await fetch('/api/v1/auth/me', {
              headers: {
                Authorization: `Bearer ${token}`,
              },
            });

            if (response.ok) {
              const userData = await response.json();
              setUser(userData);
            } else {
              // Token is invalid, clear it
              authService.logout();
              setUser(null);
            }
          } catch (error) {
            // Network error or invalid response, clear token
            console.error('Token validation failed:', error);
            authService.logout();
            setUser(null);
          }
        } else {
          // No token found, user is not authenticated
          setUser(null);
        }
      } catch (error) {
        console.error('Auth initialization error:', error);
        authService.logout();
        setUser(null);
      } finally {
        setLoading(false);
      }
    };

    initAuth();
  }, []);

  const login = async (email: string, password: string) => {
    setLoading(true);
    try {
      const response = await authService.login(email, password);
      setUser(response.user);
      return response;
    } catch (error) {
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const logout = () => {
    authService.logout();
    setUser(null);
  };

  const value: AuthContextType = {
    user,
    login,
    logout,
    isAuthenticated: !!user,
    loading,
    isDeveloper: user?.role === 'Developer' || user?.role === 'Admin',
    isAdmin: user?.role === 'Admin',
    canCreateContainer: user?.role === 'Developer' || user?.role === 'Admin',
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
