import { Theme } from '../types/theme'

export const lightTheme: Theme = {
  name: 'Light',
  mode: 'light',
  colors: {
    primary: {
      50: '#eff6ff',
      100: '#dbeafe',
      200: '#bfdbfe',
      300: '#93c5fd',
      400: '#60a5fa',
      500: '#3b82f6',
      600: '#2563eb',
      700: '#1d4ed8',
      800: '#1e40af',
      900: '#1e3a8a',
    },

    background: '#ffffff',
    foreground: '#0f172a',

    card: '#ffffff',
    cardForeground: '#0f172a',

    popover: '#ffffff',
    popoverForeground: '#0f172a',

    secondary: '#f1f5f9',
    secondaryForeground: '#334155',

    muted: '#f8fafc',
    mutedForeground: '#64748b',

    accent: '#f8fafc',
    accentForeground: '#334155',

    destructive: '#ef4444',
    destructiveForeground: '#ffffff',

    border: '#e2e8f0',
    input: '#f8fafc',

    ring: '#3b82f6',

    success: '#22c55e',
    successForeground: '#ffffff',

    warning: '#f59e0b',
    warningForeground: '#ffffff',

    info: '#06b6d4',
    infoForeground: '#ffffff',
  },
}