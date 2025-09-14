import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { User } from '../types'

interface AuthContextType {
  user: User | null
  token: string | null
  login: (token: string) => void
  logout: () => void
  isAuthenticated: boolean
  isDeveloper: boolean
  isEngineer: boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [token, setToken] = useState<string | null>(null)

  useEffect(() => {
    // Load token from localStorage on mount
    const savedToken = localStorage.getItem('burndler_token')
    if (savedToken) {
      setToken(savedToken)
      // Decode JWT to get user info (simplified - in production use a proper JWT library)
      try {
        const payload = JSON.parse(atob(savedToken.split('.')[1]))
        setUser({
          id: payload.user_id,
          email: payload.email,
          name: payload.name || payload.email,
          role: payload.role,
        })
      } catch (error) {
        console.error('Failed to decode token:', error)
        localStorage.removeItem('burndler_token')
      }
    }
  }, [])

  const login = (newToken: string) => {
    setToken(newToken)
    localStorage.setItem('burndler_token', newToken)

    // Decode JWT to get user info
    try {
      const payload = JSON.parse(atob(newToken.split('.')[1]))
      setUser({
        id: payload.user_id,
        email: payload.email,
        name: payload.name || payload.email,
        role: payload.role,
      })
    } catch (error) {
      console.error('Failed to decode token:', error)
    }
  }

  const logout = () => {
    setUser(null)
    setToken(null)
    localStorage.removeItem('burndler_token')
  }

  const value: AuthContextType = {
    user,
    token,
    login,
    logout,
    isAuthenticated: !!token,
    isDeveloper: user?.role === 'Developer',
    isEngineer: user?.role === 'Engineer',
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}