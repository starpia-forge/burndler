import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useAuth } from '../hooks/useAuth';
import ThemeToggle from './ThemeToggle';
import LanguageSelector from './LanguageSelector';

interface FormData {
  email: string;
  password: string;
}

interface FormErrors {
  email?: string;
  password?: string;
  submit?: string;
}

const LoginPage = () => {
  const navigate = useNavigate();
  const { login } = useAuth();
  const { t } = useTranslation(['auth', 'common', 'errors']);
  const [formData, setFormData] = useState<FormData>({ email: '', password: '' });
  const [errors, setErrors] = useState<FormErrors>({});
  const [isLoading, setIsLoading] = useState(false);

  const validateEmail = (email: string): string | undefined => {
    if (!email.trim()) {
      return t('auth:emailRequired');
    }
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      return t('auth:invalidEmail');
    }
    return undefined;
  };

  const validatePassword = (password: string): string | undefined => {
    if (!password) {
      return t('auth:passwordRequired');
    }
    if (password.length < 6) {
      return t('auth:passwordMinLength');
    }
    return undefined;
  };

  const handleInputChange = (field: keyof FormData, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));

    // Clear field-specific error when user starts typing
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: undefined, submit: undefined }));
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validate form
    const emailError = validateEmail(formData.email);
    const passwordError = validatePassword(formData.password);

    if (emailError || passwordError) {
      setErrors({
        email: emailError,
        password: passwordError,
      });
      return;
    }

    setIsLoading(true);
    setErrors({});

    try {
      await login(formData.email, formData.password);

      // Navigate to dashboard on successful login
      navigate('/dashboard');
    } catch (error) {
      setErrors({
        submit: error instanceof Error ? error.message : t('auth:loginFailed'),
      });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div
      data-testid="login-container"
      className="min-h-screen flex items-center justify-center bg-background px-4 py-12 sm:px-6 lg:px-8"
    >
      {/* Theme toggle and language selector in top right corner */}
      <div className="absolute top-4 right-4 flex items-center space-x-3">
        <LanguageSelector />
        <ThemeToggle />
      </div>

      <div
        data-testid="login-card"
        className="w-full max-w-md space-y-8 bg-card border border-border rounded-lg p-8 shadow-sm"
      >
        <div className="text-center">
          <h1 className="text-2xl font-bold text-foreground">{t('auth:signIn')}</h1>
          <p className="mt-2 text-sm text-muted-foreground">{t('auth:signInToAccount')}</p>
        </div>

        <form
          className="mt-8 space-y-6"
          onSubmit={handleSubmit}
          role="form"
          aria-label="Login form"
          noValidate
        >
          <div className="space-y-4">
            {/* Email field */}
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-foreground">
                {t('auth:email')}
              </label>
              <input
                id="email"
                name="email"
                type="email"
                autoComplete="email"
                required
                aria-required="true"
                aria-describedby={errors.email ? 'email-error' : undefined}
                className="mt-1 block w-full px-3 py-2 bg-background border border-input rounded-md text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                placeholder={t('auth:enterEmail')}
                value={formData.email}
                onChange={(e) => handleInputChange('email', e.target.value)}
              />
              {errors.email && (
                <p id="email-error" className="mt-1 text-sm text-destructive" role="alert">
                  {errors.email}
                </p>
              )}
            </div>

            {/* Password field */}
            <div>
              <label htmlFor="password" className="block text-sm font-medium text-foreground">
                {t('auth:password')}
              </label>
              <input
                id="password"
                name="password"
                type="password"
                autoComplete="current-password"
                required
                aria-required="true"
                aria-describedby={errors.password ? 'password-error' : undefined}
                className="mt-1 block w-full px-3 py-2 bg-background border border-input rounded-md text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
                placeholder={t('auth:enterPassword')}
                value={formData.password}
                onChange={(e) => handleInputChange('password', e.target.value)}
              />
              {errors.password && (
                <p id="password-error" className="mt-1 text-sm text-destructive" role="alert">
                  {errors.password}
                </p>
              )}
            </div>
          </div>

          {/* Submit error */}
          {errors.submit && (
            <div className="rounded-md bg-destructive/10 border border-destructive/20 p-4">
              <p className="text-sm text-destructive" role="alert">
                {errors.submit}
              </p>
            </div>
          )}

          {/* Submit button */}
          <div>
            <button
              type="submit"
              disabled={isLoading}
              className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-primary-foreground bg-primary-500 hover:bg-primary-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? t('common:loading') : t('auth:signInButton')}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default LoginPage;
