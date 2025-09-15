import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { ThemeMode, Theme, ThemeContextType } from '../types/theme'
import { lightTheme, darkTheme } from '../themes'

const ThemeContext = createContext<ThemeContextType | undefined>(undefined)

const THEME_STORAGE_KEY = 'burndler_theme_mode'

function getSystemTheme(): 'light' | 'dark' {
  if (typeof window === 'undefined') return 'light'

  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
  return mediaQuery.matches ? 'dark' : 'light'
}

function getStoredThemeMode(): ThemeMode {
  try {
    const stored = localStorage.getItem(THEME_STORAGE_KEY)
    if (stored && ['light', 'dark', 'system'].includes(stored)) {
      return stored as ThemeMode
    }
  } catch (error) {
    // localStorage might not be available
  }
  return 'system'
}

function setStoredThemeMode(mode: ThemeMode): void {
  try {
    localStorage.setItem(THEME_STORAGE_KEY, mode)
  } catch (error) {
    // localStorage might not be available
  }
}

function getEffectiveTheme(themeMode: ThemeMode): Theme {
  if (themeMode === 'system') {
    const systemTheme = getSystemTheme()
    return systemTheme === 'dark' ? darkTheme : lightTheme
  }
  return themeMode === 'dark' ? darkTheme : lightTheme
}

function updateDocumentClass(isDark: boolean): void {
  if (typeof document === 'undefined') return

  if (isDark) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [themeMode, setThemeModeState] = useState<ThemeMode>(() => getStoredThemeMode())
  const [systemTheme, setSystemTheme] = useState<'light' | 'dark'>(() => getSystemTheme())

  const theme = getEffectiveTheme(themeMode)
  const isDarkMode = themeMode === 'dark' || (themeMode === 'system' && systemTheme === 'dark')

  // Update document class whenever theme changes
  useEffect(() => {
    updateDocumentClass(isDarkMode)
  }, [isDarkMode])

  // Listen for system theme changes
  useEffect(() => {
    if (typeof window === 'undefined') return

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')

    const handleChange = (e: MediaQueryListEvent) => {
      setSystemTheme(e.matches ? 'dark' : 'light')
    }

    // Set up event listener
    if (mediaQuery.addEventListener) {
      mediaQuery.addEventListener('change', handleChange)
    } else if (mediaQuery.addListener) {
      // Fallback for older browsers
      mediaQuery.addListener(handleChange)
    }

    // Also support the onchange property for mocking in tests
    if (!mediaQuery.onchange) {
      mediaQuery.onchange = handleChange
    }

    return () => {
      if (mediaQuery.removeEventListener) {
        mediaQuery.removeEventListener('change', handleChange)
      } else if (mediaQuery.removeListener) {
        mediaQuery.removeListener(handleChange)
      }
    }
  }, [])

  const setThemeMode = (mode: ThemeMode) => {
    setThemeModeState(mode)
    setStoredThemeMode(mode)
  }

  const value: ThemeContextType = {
    theme,
    themeMode,
    setThemeMode,
    isDarkMode: themeMode === 'dark' || (themeMode === 'system' && systemTheme === 'dark'),
    isLightMode: themeMode === 'light' || (themeMode === 'system' && systemTheme === 'light'),
    isSystemMode: themeMode === 'system',
  }

  return (
    <ThemeContext.Provider value={value}>
      {children}
    </ThemeContext.Provider>
  )
}

export function useTheme(): ThemeContextType {
  const context = useContext(ThemeContext)
  if (context === undefined) {
    throw new Error('useTheme must be used within a ThemeProvider')
  }
  return context
}