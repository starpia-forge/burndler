import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { User, AuthContextType } from '../types/auth'
import { authService } from '../services/auth'

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Check for existing token on mount
    const initAuth = async () => {
      try {
        const token = authService.getAccessToken()
        if (token) {
          // TODO: Validate token and get user info
          // For now, we'll set a default user if token exists
          setUser({
            id: 1,
            email: 'user@example.com',
            role: 'Developer'
          })
        }
      } catch (error) {
        console.error('Auth initialization error:', error)
        authService.logout()
      } finally {
        setLoading(false)
      }
    }

    initAuth()
  }, [])

  const login = async (email: string, password: string) => {
    setLoading(true)
    try {
      const response = await authService.login(email, password)
      setUser(response.user)
      return response
    } catch (error) {
      throw error
    } finally {
      setLoading(false)
    }
  }

  const logout = () => {
    authService.logout()
    setUser(null)
  }

  const value: AuthContextType = {
    user,
    login,
    logout,
    isAuthenticated: !!user,
    loading,
    isDeveloper: user?.role === 'Developer'
  }

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}