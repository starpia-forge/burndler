import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ThemeProvider } from '../hooks/useTheme'
import ThemeToggle from './ThemeToggle'

// Mock the theme hook for controlled testing
vi.mock('../hooks/useTheme', async () => {
  const actual = await vi.importActual('../hooks/useTheme')
  return {
    ...actual,
    useTheme: vi.fn(),
  }
})

const mockUseTheme = vi.mocked(await import('../hooks/useTheme')).useTheme

const renderWithTheme = (component: React.ReactElement) => {
  return render(
    <ThemeProvider>
      {component}
    </ThemeProvider>
  )
}

describe('ThemeToggle', () => {
  const mockSetThemeMode = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('rendering', () => {
    it('should render toggle button', () => {
      mockUseTheme.mockReturnValue({
        theme: {
          name: 'Light',
          mode: 'light',
          colors: {} as any,
        },
        themeMode: 'light',
        setThemeMode: mockSetThemeMode,
        isDarkMode: false,
        isLightMode: true,
        isSystemMode: false,
      })

      renderWithTheme(<ThemeToggle />)

      expect(screen.getByRole('button')).toBeInTheDocument()
      expect(screen.getByLabelText(/toggle theme/i)).toBeInTheDocument()
    })

    it('should show light mode icon when in light mode', () => {
      mockUseTheme.mockReturnValue({
        theme: {
          name: 'Light',
          mode: 'light',
          colors: {} as any,
        },
        themeMode: 'light',
        setThemeMode: mockSetThemeMode,
        isDarkMode: false,
        isLightMode: true,
        isSystemMode: false,
      })

      renderWithTheme(<ThemeToggle />)

      // Should show sun icon for light mode
      expect(screen.getByTestId('sun-icon')).toBeInTheDocument()
    })

    it('should show dark mode icon when in dark mode', () => {
      mockUseTheme.mockReturnValue({
        theme: {
          name: 'Dark',
          mode: 'dark',
          colors: {} as any,
        },
        themeMode: 'dark',
        setThemeMode: mockSetThemeMode,
        isDarkMode: true,
        isLightMode: false,
        isSystemMode: false,
      })

      renderWithTheme(<ThemeToggle />)

      // Should show moon icon for dark mode
      expect(screen.getByTestId('moon-icon')).toBeInTheDocument()
    })

    it('should show system mode icon when in system mode', () => {
      mockUseTheme.mockReturnValue({
        theme: {
          name: 'Light',
          mode: 'light',
          colors: {} as any,
        },
        themeMode: 'system',
        setThemeMode: mockSetThemeMode,
        isDarkMode: false,
        isLightMode: false,
        isSystemMode: true,
      })

      renderWithTheme(<ThemeToggle />)

      // Should show computer/monitor icon for system mode
      expect(screen.getByTestId('computer-icon')).toBeInTheDocument()
    })
  })

  describe('dropdown menu', () => {
    it('should show dropdown menu when clicked', async () => {
      const user = userEvent.setup()

      mockUseTheme.mockReturnValue({
        theme: {
          name: 'Light',
          mode: 'light',
          colors: {} as any,
        },
        themeMode: 'light',
        setThemeMode: mockSetThemeMode,
        isDarkMode: false,
        isLightMode: true,
        isSystemMode: false,
      })

      renderWithTheme(<ThemeToggle />)

      const toggleButton = screen.getByRole('button')
      await user.click(toggleButton)

      expect(screen.getByText('Light')).toBeInTheDocument()
      expect(screen.getByText('Dark')).toBeInTheDocument()
      expect(screen.getByText('System')).toBeInTheDocument()
    })

    it('should show current theme as selected in dropdown', async () => {
      const user = userEvent.setup()

      mockUseTheme.mockReturnValue({
        theme: {
          name: 'Dark',
          mode: 'dark',
          colors: {} as any,
        },
        themeMode: 'dark',
        setThemeMode: mockSetThemeMode,
        isDarkMode: true,
        isLightMode: false,
        isSystemMode: false,
      })

      renderWithTheme(<ThemeToggle />)

      const toggleButton = screen.getByRole('button')
      await user.click(toggleButton)

      // Current theme should be marked as selected
      const darkOption = screen.getByText('Dark').closest('button')
      expect(darkOption).toHaveClass('bg-accent') // or whatever class indicates selection
    })
  })

  describe('theme switching', () => {
    it('should switch to light mode when light option is clicked', async () => {
      const user = userEvent.setup()

      mockUseTheme.mockReturnValue({
        theme: {
          name: 'Dark',
          mode: 'dark',
          colors: {} as any,
        },
        themeMode: 'dark',
        setThemeMode: mockSetThemeMode,
        isDarkMode: true,
        isLightMode: false,
        isSystemMode: false,
      })

      renderWithTheme(<ThemeToggle />)

      const toggleButton = screen.getByRole('button')
      await user.click(toggleButton)

      const lightOption = screen.getByText('Light')
      await user.click(lightOption)

      expect(mockSetThemeMode).toHaveBeenCalledWith('light')
    })

    it('should switch to dark mode when dark option is clicked', async () => {
      const user = userEvent.setup()

      mockUseTheme.mockReturnValue({
        theme: {
          name: 'Light',
          mode: 'light',
          colors: {} as any,
        },
        themeMode: 'light',
        setThemeMode: mockSetThemeMode,
        isDarkMode: false,
        isLightMode: true,
        isSystemMode: false,
      })

      renderWithTheme(<ThemeToggle />)

      const toggleButton = screen.getByRole('button')
      await user.click(toggleButton)

      const darkOption = screen.getByText('Dark')
      await user.click(darkOption)

      expect(mockSetThemeMode).toHaveBeenCalledWith('dark')
    })

    it('should switch to system mode when system option is clicked', async () => {
      const user = userEvent.setup()

      mockUseTheme.mockReturnValue({
        theme: {
          name: 'Light',
          mode: 'light',
          colors: {} as any,
        },
        themeMode: 'light',
        setThemeMode: mockSetThemeMode,
        isDarkMode: false,
        isLightMode: true,
        isSystemMode: false,
      })

      renderWithTheme(<ThemeToggle />)

      const toggleButton = screen.getByRole('button')
      await user.click(toggleButton)

      const systemOption = screen.getByText('System')
      await user.click(systemOption)

      expect(mockSetThemeMode).toHaveBeenCalledWith('system')
    })
  })

  describe('accessibility', () => {
    it('should have proper ARIA attributes', () => {
      mockUseTheme.mockReturnValue({
        theme: {
          name: 'Light',
          mode: 'light',
          colors: {} as any,
        },
        themeMode: 'light',
        setThemeMode: mockSetThemeMode,
        isDarkMode: false,
        isLightMode: true,
        isSystemMode: false,
      })

      renderWithTheme(<ThemeToggle />)

      const toggleButton = screen.getByRole('button')
      expect(toggleButton).toHaveAttribute('aria-label', expect.stringContaining('Toggle theme'))
      expect(toggleButton).toHaveAttribute('aria-haspopup', 'true')
    })

    it('should support keyboard navigation', async () => {
      const user = userEvent.setup()

      mockUseTheme.mockReturnValue({
        theme: {
          name: 'Light',
          mode: 'light',
          colors: {} as any,
        },
        themeMode: 'light',
        setThemeMode: mockSetThemeMode,
        isDarkMode: false,
        isLightMode: true,
        isSystemMode: false,
      })

      renderWithTheme(<ThemeToggle />)

      const toggleButton = screen.getByRole('button')

      // Focus and activate with Enter
      toggleButton.focus()
      await user.keyboard('{Enter}')

      expect(screen.getByText('Light')).toBeInTheDocument()
      expect(screen.getByText('Dark')).toBeInTheDocument()
      expect(screen.getByText('System')).toBeInTheDocument()
    })

    it('should close dropdown when Escape is pressed', async () => {
      const user = userEvent.setup()

      mockUseTheme.mockReturnValue({
        theme: {
          name: 'Light',
          mode: 'light',
          colors: {} as any,
        },
        themeMode: 'light',
        setThemeMode: mockSetThemeMode,
        isDarkMode: false,
        isLightMode: true,
        isSystemMode: false,
      })

      renderWithTheme(<ThemeToggle />)

      const toggleButton = screen.getByRole('button')
      await user.click(toggleButton)

      expect(screen.getByText('Light')).toBeInTheDocument()

      await user.keyboard('{Escape}')

      expect(screen.queryByText('Light')).not.toBeInTheDocument()
    })
  })

  describe('responsive design', () => {
    it('should be compact on mobile screens', () => {
      mockUseTheme.mockReturnValue({
        theme: {
          name: 'Light',
          mode: 'light',
          colors: {} as any,
        },
        themeMode: 'light',
        setThemeMode: mockSetThemeMode,
        isDarkMode: false,
        isLightMode: true,
        isSystemMode: false,
      })

      renderWithTheme(<ThemeToggle />)

      const toggleButton = screen.getByRole('button')
      expect(toggleButton).toHaveClass('p-2') // Should be compact padding
    })
  })

  describe('icons', () => {
    it('should display correct icon for each theme mode', () => {
      const modes = [
        { mode: 'light' as const, icon: 'sun-icon' },
        { mode: 'dark' as const, icon: 'moon-icon' },
        { mode: 'system' as const, icon: 'computer-icon' },
      ]

      modes.forEach(({ mode, icon }) => {
        mockUseTheme.mockReturnValue({
          theme: {
            name: mode === 'light' ? 'Light' : 'Dark',
            mode: mode === 'system' ? 'light' : mode,
            colors: {} as any,
          },
          themeMode: mode,
          setThemeMode: mockSetThemeMode,
          isDarkMode: mode === 'dark',
          isLightMode: mode === 'light',
          isSystemMode: mode === 'system',
        })

        const { unmount } = renderWithTheme(<ThemeToggle />)
        expect(screen.getByTestId(icon)).toBeInTheDocument()
        unmount()
      })
    })
  })
})