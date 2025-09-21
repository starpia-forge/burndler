import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useSetup } from '../../hooks/useSetup';
import { UserPlusIcon, EyeIcon, EyeSlashIcon } from '@heroicons/react/24/outline';
import { useSetupWizardContext } from '../../contexts/SetupWizardContext';

interface AdminSetupProps {
  hasAdmin: boolean;
  onAdminCreated: () => void;
}

export default function AdminSetup({ hasAdmin, onAdminCreated }: AdminSetupProps) {
  const { t } = useTranslation('setup');
  const { createAdmin, loading, error } = useSetup();
  const { setAdminData } = useSetupWizardContext();
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    password: '',
    confirmPassword: '',
  });
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [formError, setFormError] = useState('');

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.value,
    });
    setFormError('');
  };

  const validateForm = () => {
    if (!formData.name.trim()) {
      setFormError(t('adminStep.errors.nameRequired'));
      return false;
    }
    if (!formData.email.trim()) {
      setFormError(t('adminStep.errors.emailRequired'));
      return false;
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      setFormError(t('adminStep.errors.invalidEmail'));
      return false;
    }
    if (formData.password.length < 8) {
      setFormError(t('adminStep.errors.passwordMinLength'));
      return false;
    }
    if (formData.password !== formData.confirmPassword) {
      setFormError(t('adminStep.errors.passwordMismatch'));
      return false;
    }
    return true;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    try {
      await createAdmin({
        name: formData.name,
        email: formData.email,
        password: formData.password,
      });

      // Save admin data to context for potential future use
      setAdminData({
        name: formData.name,
        email: formData.email,
        password: formData.password,
      });

      onAdminCreated();
    } catch (err) {
      // Error is already handled by useSetup hook
    }
  };

  if (hasAdmin) {
    return (
      <div className="p-8 text-center">
        <UserPlusIcon className="h-16 w-16 text-success mx-auto mb-4" />
        <h2 className="text-2xl font-bold text-foreground mb-2">
          {t('adminStep.adminReady.title')}
        </h2>
        <p className="text-muted-foreground mb-6">{t('adminStep.adminReady.description')}</p>
        <button
          onClick={onAdminCreated}
          className="bg-primary-600 text-white px-6 py-2 rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 transition-colors"
        >
          {t('adminStep.adminReady.continue')}
        </button>
      </div>
    );
  }

  return (
    <div className="p-8">
      <div className="text-center mb-8">
        <UserPlusIcon className="h-16 w-16 text-primary-600 mx-auto mb-4" />
        <h2 className="text-2xl font-bold text-foreground mb-2">{t('adminStep.title')}</h2>
        <p className="text-muted-foreground">{t('adminStep.description')}</p>
      </div>

      {(error || formError) && (
        <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4 mb-6">
          <div className="text-sm text-destructive">{error || formError}</div>
        </div>
      )}

      <form onSubmit={handleSubmit} className="max-w-md mx-auto space-y-6">
        <div>
          <label htmlFor="name" className="block text-sm font-medium text-foreground mb-2">
            {t('adminStep.fullNameLabel')}
          </label>
          <input
            type="text"
            id="name"
            name="name"
            value={formData.name}
            onChange={handleInputChange}
            required
            className="w-full px-3 py-2 bg-background border border-input rounded-md text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            placeholder={t('adminStep.fullNamePlaceholder')}
          />
        </div>

        <div>
          <label htmlFor="email" className="block text-sm font-medium text-foreground mb-2">
            {t('adminStep.emailLabel')}
          </label>
          <input
            type="email"
            id="email"
            name="email"
            value={formData.email}
            onChange={handleInputChange}
            required
            className="w-full px-3 py-2 bg-background border border-input rounded-md text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            placeholder={t('adminStep.emailPlaceholder')}
          />
        </div>

        <div>
          <label htmlFor="password" className="block text-sm font-medium text-foreground mb-2">
            {t('adminStep.passwordLabel')}
          </label>
          <div className="relative">
            <input
              type={showPassword ? 'text' : 'password'}
              id="password"
              name="password"
              value={formData.password}
              onChange={handleInputChange}
              required
              className="w-full px-3 py-2 pr-10 bg-background border border-input rounded-md text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              placeholder={t('adminStep.passwordPlaceholder')}
            />
            <button
              type="button"
              className="absolute inset-y-0 right-0 pr-3 flex items-center"
              onClick={() => setShowPassword(!showPassword)}
            >
              {showPassword ? (
                <EyeSlashIcon className="h-5 w-5 text-muted-foreground" />
              ) : (
                <EyeIcon className="h-5 w-5 text-muted-foreground" />
              )}
            </button>
          </div>
        </div>

        <div>
          <label
            htmlFor="confirmPassword"
            className="block text-sm font-medium text-foreground mb-2"
          >
            {t('adminStep.confirmPasswordLabel')}
          </label>
          <div className="relative">
            <input
              type={showConfirmPassword ? 'text' : 'password'}
              id="confirmPassword"
              name="confirmPassword"
              value={formData.confirmPassword}
              onChange={handleInputChange}
              required
              className="w-full px-3 py-2 pr-10 bg-background border border-input rounded-md text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent"
              placeholder={t('adminStep.confirmPasswordPlaceholder')}
            />
            <button
              type="button"
              className="absolute inset-y-0 right-0 pr-3 flex items-center"
              onClick={() => setShowConfirmPassword(!showConfirmPassword)}
            >
              {showConfirmPassword ? (
                <EyeSlashIcon className="h-5 w-5 text-muted-foreground" />
              ) : (
                <EyeIcon className="h-5 w-5 text-muted-foreground" />
              )}
            </button>
          </div>
        </div>

        <div className="bg-primary/10 border border-primary/20 rounded-lg p-4">
          <div className="text-sm text-primary-700">
            <p className="font-medium mb-1">{t('adminStep.privilegesTitle')}</p>
            <ul className="list-disc list-inside space-y-1">
              <li>{t('adminStep.privileges.access')}</li>
              <li>{t('adminStep.privileges.userManagement')}</li>
              <li>{t('adminStep.privileges.monitoring')}</li>
              <li>{t('adminStep.privileges.security')}</li>
            </ul>
          </div>
        </div>

        <button
          type="submit"
          disabled={loading}
          className="w-full bg-primary-600 text-white py-2 px-4 rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? t('adminStep.creatingAccount') : t('adminStep.createAccount')}
        </button>
      </form>
    </div>
  );
}
