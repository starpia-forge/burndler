import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ThemeProvider } from '../hooks/useTheme';
import { AuthProvider } from '../hooks/useAuth';
import LoginPage from './LoginPage';

// Mock react-router-dom
const mockNavigateFn = vi.fn();
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigateFn,
}));

// Mock the auth service
vi.mock('../services/auth', () => ({
  authService: {
    login: vi.fn(),
    logout: vi.fn(),
    getAccessToken: vi.fn(),
    getRefreshToken: vi.fn(),
    refreshToken: vi.fn(),
    isAuthenticated: vi.fn(),
  },
}));

const renderWithProviders = (component: React.ReactElement) => {
  return render(
    <ThemeProvider>
      <AuthProvider>{component}</AuthProvider>
    </ThemeProvider>
  );
};

describe('LoginPage', () => {
  let mockLogin: ReturnType<typeof vi.fn>;

  beforeEach(async () => {
    vi.clearAllMocks();

    // Import mocked modules
    const { authService } = await import('../services/auth');

    mockLogin = authService.login as any;

    // Set up default mock returns
    (authService.getAccessToken as any).mockReturnValue(null);
    (authService.isAuthenticated as any).mockReturnValue(false);
  });

  describe('rendering', () => {
    it('should render login form elements', () => {
      renderWithProviders(<LoginPage />);

      expect(screen.getByRole('heading', { name: /sign in/i })).toBeInTheDocument();
      expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
    });

    it('should render email and password input fields', () => {
      renderWithProviders(<LoginPage />);

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/password/i);

      expect(emailInput).toHaveAttribute('type', 'email');
      expect(passwordInput).toHaveAttribute('type', 'password');
      expect(emailInput).toBeRequired();
      expect(passwordInput).toBeRequired();
    });

    it('should have proper form structure', () => {
      renderWithProviders(<LoginPage />);

      const form = screen.getByRole('form');
      expect(form).toBeInTheDocument();

      // Check form has proper submission handling
      expect(form).toHaveAttribute('noValidate');
    });
  });

  describe('form validation', () => {
    it('should show validation errors for empty fields', async () => {
      const user = userEvent.setup();
      renderWithProviders(<LoginPage />);

      const submitButton = screen.getByRole('button', { name: /sign in/i });
      await user.click(submitButton);

      expect(screen.getByText(/email is required/i)).toBeInTheDocument();
      expect(screen.getByText(/password is required/i)).toBeInTheDocument();
    });

    it('should show validation error for invalid email format', async () => {
      const user = userEvent.setup();
      renderWithProviders(<LoginPage />);

      const emailInput = screen.getByLabelText(/email/i);
      const submitButton = screen.getByRole('button', { name: /sign in/i });

      await user.type(emailInput, 'invalid-email');
      await user.click(submitButton);

      expect(screen.getByText(/please enter a valid email/i)).toBeInTheDocument();
    });

    it('should show validation error for short password', async () => {
      const user = userEvent.setup();
      renderWithProviders(<LoginPage />);

      const passwordInput = screen.getByLabelText(/password/i);
      const submitButton = screen.getByRole('button', { name: /sign in/i });

      await user.type(passwordInput, '123');
      await user.click(submitButton);

      expect(screen.getByText(/password must be at least 6 characters/i)).toBeInTheDocument();
    });

    it('should clear validation errors when user starts typing', async () => {
      const user = userEvent.setup();
      renderWithProviders(<LoginPage />);

      const emailInput = screen.getByLabelText(/email/i);
      const submitButton = screen.getByRole('button', { name: /sign in/i });

      // Trigger validation error
      await user.click(submitButton);
      expect(screen.getByText(/email is required/i)).toBeInTheDocument();

      // Start typing should clear error
      await user.type(emailInput, 'user@example.com');
      expect(screen.queryByText(/email is required/i)).not.toBeInTheDocument();
    });
  });

  describe('form submission', () => {
    it('should submit form with valid credentials', async () => {
      const user = userEvent.setup();
      mockLogin.mockResolvedValue({
        accessToken: 'fake-token',
        refreshToken: 'fake-refresh-token',
        user: { id: 1, email: 'user@example.com', role: 'Developer' },
      });

      renderWithProviders(<LoginPage />);

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/password/i);
      const submitButton = screen.getByRole('button', { name: /sign in/i });

      await user.type(emailInput, 'user@example.com');
      await user.type(passwordInput, 'password123');
      await user.click(submitButton);

      await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledWith('user@example.com', 'password123');
      });
    });

    it('should show loading state during submission', async () => {
      const user = userEvent.setup();
      mockLogin.mockImplementation(() => new Promise((resolve) => setTimeout(resolve, 1000)));

      renderWithProviders(<LoginPage />);

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/password/i);
      const submitButton = screen.getByRole('button', { name: /sign in/i });

      await user.type(emailInput, 'user@example.com');
      await user.type(passwordInput, 'password123');
      await user.click(submitButton);

      expect(screen.getByText(/loading/i)).toBeInTheDocument();
      expect(submitButton).toBeDisabled();
    });

    it('should navigate to dashboard on successful login', async () => {
      const user = userEvent.setup();
      mockLogin.mockResolvedValue({
        accessToken: 'fake-token',
        refreshToken: 'fake-refresh-token',
        user: { id: 1, email: 'user@example.com', role: 'Developer' },
      });

      renderWithProviders(<LoginPage />);

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/password/i);
      const submitButton = screen.getByRole('button', { name: /sign in/i });

      await user.type(emailInput, 'user@example.com');
      await user.type(passwordInput, 'password123');
      await user.click(submitButton);

      await waitFor(() => {
        expect(mockNavigateFn).toHaveBeenCalledWith('/dashboard');
      });
    });

    it('should show error message on login failure', async () => {
      const user = userEvent.setup();
      mockLogin.mockRejectedValue(new Error('Invalid credentials'));

      renderWithProviders(<LoginPage />);

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/password/i);
      const submitButton = screen.getByRole('button', { name: /sign in/i });

      await user.type(emailInput, 'user@example.com');
      await user.type(passwordInput, 'wrongpassword');
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText(/invalid credentials/i)).toBeInTheDocument();
      });
    });

    it('should handle network errors gracefully', async () => {
      const user = userEvent.setup();
      mockLogin.mockRejectedValue(new Error('Network error'));

      renderWithProviders(<LoginPage />);

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/password/i);
      const submitButton = screen.getByRole('button', { name: /sign in/i });

      await user.type(emailInput, 'user@example.com');
      await user.type(passwordInput, 'password123');
      await user.click(submitButton);

      await waitFor(() => {
        expect(screen.getByText(/network error/i)).toBeInTheDocument();
      });
    });
  });

  describe('accessibility', () => {
    it('should have proper ARIA labels and roles', () => {
      renderWithProviders(<LoginPage />);

      const form = screen.getByRole('form');
      expect(form).toHaveAttribute('aria-label', 'Login form');

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/password/i);

      expect(emailInput).toHaveAttribute('aria-required', 'true');
      expect(passwordInput).toHaveAttribute('aria-required', 'true');
    });

    it('should associate error messages with form fields', async () => {
      const user = userEvent.setup();
      renderWithProviders(<LoginPage />);

      const emailInput = screen.getByLabelText(/email/i);
      const submitButton = screen.getByRole('button', { name: /sign in/i });

      await user.click(submitButton);

      const emailError = screen.getByText(/email is required/i);
      expect(emailError).toHaveAttribute('id');
      expect(emailInput).toHaveAttribute('aria-describedby', emailError.id);
    });

    it('should support keyboard navigation', async () => {
      const user = userEvent.setup();
      renderWithProviders(<LoginPage />);

      const emailInput = screen.getByLabelText(/email/i);
      const passwordInput = screen.getByLabelText(/password/i);
      const submitButton = screen.getByRole('button', { name: /sign in/i });

      // Focus email input directly to test keyboard navigation from there
      emailInput.focus();
      expect(emailInput).toHaveFocus();

      await user.tab();
      expect(passwordInput).toHaveFocus();

      await user.tab();
      expect(submitButton).toHaveFocus();
    });
  });

  describe('responsive design', () => {
    it('should have responsive layout classes', () => {
      renderWithProviders(<LoginPage />);

      const container = screen.getByTestId('login-container');
      expect(container).toHaveClass('min-h-screen', 'flex', 'items-center', 'justify-center');

      const formCard = screen.getByTestId('login-card');
      expect(formCard).toHaveClass('w-full', 'max-w-md');
    });
  });

  describe('theming', () => {
    it('should use theme colors and styles', () => {
      renderWithProviders(<LoginPage />);

      const formCard = screen.getByTestId('login-card');
      expect(formCard).toHaveClass('bg-card', 'border', 'border-border');

      const emailInput = screen.getByLabelText(/email/i);
      expect(emailInput).toHaveClass('bg-background', 'border-input');
    });

    it('should include theme toggle component', () => {
      renderWithProviders(<LoginPage />);

      const themeToggle = screen.getByLabelText(/toggle theme/i);
      expect(themeToggle).toBeInTheDocument();
    });
  });
});
