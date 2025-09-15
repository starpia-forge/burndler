import { Theme } from '../types/theme'

export const darkTheme: Theme = {
  name: 'Dark',
  mode: 'dark',
  colors: {
    primary: {
      50: '#1e3a8a',
      100: '#1e40af',
      200: '#1d4ed8',
      300: '#2563eb',
      400: '#3b82f6',
      500: '#60a5fa',
      600: '#93c5fd',
      700: '#bfdbfe',
      800: '#dbeafe',
      900: '#eff6ff',
    },

    background: '#0f172a',
    foreground: '#f8fafc',

    card: '#1e293b',
    cardForeground: '#f8fafc',

    popover: '#1e293b',
    popoverForeground: '#f8fafc',

    secondary: '#334155',
    secondaryForeground: '#f1f5f9',

    muted: '#1e293b',
    mutedForeground: '#94a3b8',

    accent: '#334155',
    accentForeground: '#f1f5f9',

    destructive: '#dc2626',
    destructiveForeground: '#ffffff',

    border: '#334155',
    input: '#1e293b',

    ring: '#60a5fa',

    success: '#16a34a',
    successForeground: '#ffffff',

    warning: '#ca8a04',
    warningForeground: '#ffffff',

    info: '#0891b2',
    infoForeground: '#ffffff',
  },
}