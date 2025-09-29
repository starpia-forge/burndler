export type ThemeMode = 'light' | 'dark' | 'system';

export interface ThemeColors {
  // Brand colors
  primary: {
    50: string;
    100: string;
    200: string;
    300: string;
    400: string;
    500: string;
    600: string;
    700: string;
    800: string;
    900: string;
  };

  // Background colors
  background: string;
  foreground: string;

  // Card/Surface colors
  card: string;
  cardForeground: string;

  // Popover colors
  popover: string;
  popoverForeground: string;

  // Secondary colors
  secondary: string;
  secondaryForeground: string;

  // Muted colors
  muted: string;
  mutedForeground: string;

  // Accent colors
  accent: string;
  accentForeground: string;

  // Destructive/Error colors
  destructive: string;
  destructiveForeground: string;

  // Border and input colors
  border: string;
  input: string;

  // Ring color for focus states
  ring: string;

  // Success colors
  success: string;
  successForeground: string;

  // Warning colors
  warning: string;
  warningForeground: string;

  // Info colors
  info: string;
  infoForeground: string;
}

export interface Theme {
  name: string;
  mode: 'light' | 'dark';
  colors: ThemeColors;
}

export interface ThemeContextType {
  theme: Theme;
  themeMode: ThemeMode;
  setThemeMode: (mode: ThemeMode) => void;
  isDarkMode: boolean;
  isLightMode: boolean;
  isSystemMode: boolean;
}
