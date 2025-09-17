import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { ReactNode } from 'react';
import { ThemeProvider, useTheme } from './useTheme';

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
};
Object.defineProperty(window, 'localStorage', {
  value: localStorageMock,
});

// Mock matchMedia
const mockMatchMedia = vi.fn();
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: mockMatchMedia,
});

const wrapper = ({ children }: { children: ReactNode }) => (
  <ThemeProvider>{children}</ThemeProvider>
);

describe('useTheme', () => {
  beforeEach(() => {
    localStorageMock.getItem.mockClear();
    localStorageMock.setItem.mockClear();
    localStorageMock.removeItem.mockClear();
    mockMatchMedia.mockClear();

    // Default matchMedia mock
    mockMatchMedia.mockReturnValue({
      matches: false,
      media: '(prefers-color-scheme: dark)',
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    });
  });

  afterEach(() => {
    document.documentElement.classList.remove('dark');
  });

  describe('initialization', () => {
    it('should initialize with system theme mode when no stored preference', () => {
      localStorageMock.getItem.mockReturnValue(null);

      const { result } = renderHook(() => useTheme(), { wrapper });

      expect(result.current.themeMode).toBe('system');
      expect(result.current.isSystemMode).toBe(true);
      expect(localStorageMock.getItem).toHaveBeenCalledWith('burndler_theme_mode');
    });

    it('should initialize with stored theme mode', () => {
      localStorageMock.getItem.mockReturnValue('dark');

      const { result } = renderHook(() => useTheme(), { wrapper });

      expect(result.current.themeMode).toBe('dark');
      expect(result.current.isDarkMode).toBe(true);
    });

    it('should detect system dark mode preference', () => {
      localStorageMock.getItem.mockReturnValue('system');
      mockMatchMedia.mockReturnValue({
        matches: true,
        media: '(prefers-color-scheme: dark)',
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      });

      const { result } = renderHook(() => useTheme(), { wrapper });

      expect(result.current.themeMode).toBe('system');
      expect(result.current.theme.mode).toBe('dark');
    });

    it('should detect system light mode preference', () => {
      localStorageMock.getItem.mockReturnValue('system');
      mockMatchMedia.mockReturnValue({
        matches: false,
        media: '(prefers-color-scheme: dark)',
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      });

      const { result } = renderHook(() => useTheme(), { wrapper });

      expect(result.current.themeMode).toBe('system');
      expect(result.current.theme.mode).toBe('light');
    });
  });

  describe('theme switching', () => {
    it('should switch to dark mode', () => {
      const { result } = renderHook(() => useTheme(), { wrapper });

      act(() => {
        result.current.setThemeMode('dark');
      });

      expect(result.current.themeMode).toBe('dark');
      expect(result.current.isDarkMode).toBe(true);
      expect(result.current.theme.mode).toBe('dark');
      expect(localStorageMock.setItem).toHaveBeenCalledWith('burndler_theme_mode', 'dark');
      expect(document.documentElement.classList.contains('dark')).toBe(true);
    });

    it('should switch to light mode', () => {
      const { result } = renderHook(() => useTheme(), { wrapper });

      act(() => {
        result.current.setThemeMode('light');
      });

      expect(result.current.themeMode).toBe('light');
      expect(result.current.isLightMode).toBe(true);
      expect(result.current.theme.mode).toBe('light');
      expect(localStorageMock.setItem).toHaveBeenCalledWith('burndler_theme_mode', 'light');
      expect(document.documentElement.classList.contains('dark')).toBe(false);
    });

    it('should switch to system mode', () => {
      mockMatchMedia.mockReturnValue({
        matches: true,
        media: '(prefers-color-scheme: dark)',
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      });

      const { result } = renderHook(() => useTheme(), { wrapper });

      act(() => {
        result.current.setThemeMode('system');
      });

      expect(result.current.themeMode).toBe('system');
      expect(result.current.isSystemMode).toBe(true);
      expect(result.current.theme.mode).toBe('dark'); // Based on system preference
      expect(localStorageMock.setItem).toHaveBeenCalledWith('burndler_theme_mode', 'system');
    });
  });

  describe('DOM manipulation', () => {
    it('should add dark class to document element in dark mode', () => {
      const { result } = renderHook(() => useTheme(), { wrapper });

      act(() => {
        result.current.setThemeMode('dark');
      });

      expect(document.documentElement.classList.contains('dark')).toBe(true);
    });

    it('should remove dark class from document element in light mode', () => {
      document.documentElement.classList.add('dark');

      const { result } = renderHook(() => useTheme(), { wrapper });

      act(() => {
        result.current.setThemeMode('light');
      });

      expect(document.documentElement.classList.contains('dark')).toBe(false);
    });

    it('should update DOM based on system preference when in system mode', async () => {
      const mockMediaQuery = {
        matches: true,
        media: '(prefers-color-scheme: dark)',
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      };
      mockMatchMedia.mockReturnValue(mockMediaQuery);

      const { result } = renderHook(() => useTheme(), { wrapper });

      act(() => {
        result.current.setThemeMode('system');
      });

      expect(document.documentElement.classList.contains('dark')).toBe(true);

      // Simulate system theme change
      mockMediaQuery.matches = false;

      // Create a new event object for the callback
      const changeEvent = {
        ...mockMediaQuery,
        matches: false,
      } as unknown as MediaQueryListEvent;

      act(() => {
        const onchange = mockMediaQuery.onchange as ((event: MediaQueryListEvent) => void) | null;
        onchange?.(changeEvent);
      });

      // Wait for the DOM update
      await act(async () => {
        await new Promise((resolve) => setTimeout(resolve, 0));
      });

      expect(document.documentElement.classList.contains('dark')).toBe(false);
    });
  });

  describe('localStorage persistence', () => {
    it('should save theme mode to localStorage', () => {
      const { result } = renderHook(() => useTheme(), { wrapper });

      act(() => {
        result.current.setThemeMode('dark');
      });

      expect(localStorageMock.setItem).toHaveBeenCalledWith('burndler_theme_mode', 'dark');
    });

    it('should load theme mode from localStorage on initialization', () => {
      localStorageMock.getItem.mockReturnValue('light');

      const { result } = renderHook(() => useTheme(), { wrapper });

      expect(result.current.themeMode).toBe('light');
      expect(localStorageMock.getItem).toHaveBeenCalledWith('burndler_theme_mode');
    });

    it('should handle invalid localStorage values gracefully', () => {
      localStorageMock.getItem.mockReturnValue('invalid-theme');

      const { result } = renderHook(() => useTheme(), { wrapper });

      expect(result.current.themeMode).toBe('system'); // Should fallback to system
    });
  });

  describe('error handling', () => {
    it('should throw error when used outside ThemeProvider', () => {
      expect(() => {
        renderHook(() => useTheme());
      }).toThrow('useTheme must be used within a ThemeProvider');
    });

    it('should handle localStorage errors gracefully', () => {
      localStorageMock.setItem.mockImplementation(() => {
        throw new Error('localStorage error');
      });

      const { result } = renderHook(() => useTheme(), { wrapper });

      expect(() => {
        act(() => {
          result.current.setThemeMode('dark');
        });
      }).not.toThrow();
    });
  });

  describe('theme object properties', () => {
    it('should provide correct theme object for light mode', () => {
      const { result } = renderHook(() => useTheme(), { wrapper });

      act(() => {
        result.current.setThemeMode('light');
      });

      expect(result.current.theme).toMatchObject({
        name: 'Light',
        mode: 'light',
        colors: expect.objectContaining({
          background: expect.any(String),
          foreground: expect.any(String),
          primary: expect.objectContaining({
            500: expect.any(String),
          }),
        }),
      });
    });

    it('should provide correct theme object for dark mode', () => {
      const { result } = renderHook(() => useTheme(), { wrapper });

      act(() => {
        result.current.setThemeMode('dark');
      });

      expect(result.current.theme).toMatchObject({
        name: 'Dark',
        mode: 'dark',
        colors: expect.objectContaining({
          background: expect.any(String),
          foreground: expect.any(String),
          primary: expect.objectContaining({
            500: expect.any(String),
          }),
        }),
      });
    });
  });
});
