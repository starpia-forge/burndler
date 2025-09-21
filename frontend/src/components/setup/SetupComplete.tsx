import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { CheckCircleIcon, ArrowRightIcon } from '@heroicons/react/24/outline';
import { useAuth } from '../../hooks/useAuth';
import { useNavigate } from 'react-router-dom';
import { useSetup } from '../../hooks/useSetup';
import { useSetupWizardContext } from '../../contexts/SetupWizardContext';

export default function SetupComplete() {
  const { t } = useTranslation('setup');
  const { isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const { completeSetup, error: setupError } = useSetup();
  const { wizardData, clearWizardData } = useSetupWizardContext();
  const [isCompleting, setIsCompleting] = useState(true);
  const [isSetupFinished, setIsSetupFinished] = useState(false);
  const [localError, setLocalError] = useState<string | null>(null);

  useEffect(() => {
    const finalizeSetup = async () => {
      if (!wizardData.systemConfig) {
        setLocalError(t('completeStep.errors.missingConfig'));
        setIsCompleting(false);
        return;
      }

      try {
        await completeSetup({
          company_name: wizardData.systemConfig.companyName,
          system_settings: wizardData.systemConfig.systemSettings,
        });

        // Clear wizard data after successful completion
        clearWizardData();
        setIsSetupFinished(true);
        setIsCompleting(false);
      } catch (err) {
        setLocalError(t('completeStep.errors.setupFailed'));
        setIsCompleting(false);
      }
    };

    finalizeSetup();
  }, [wizardData.systemConfig, completeSetup, clearWizardData, t]);

  useEffect(() => {
    // Auto-redirect to dashboard after 5 seconds if user is authenticated and setup is finished
    if (isAuthenticated && isSetupFinished) {
      const timer = setTimeout(() => {
        navigate('/');
      }, 5000);

      return () => clearTimeout(timer);
    }
  }, [isAuthenticated, isSetupFinished, navigate]);

  const handleGoToDashboard = () => {
    if (isAuthenticated && isSetupFinished) {
      navigate('/');
    } else {
      navigate('/login');
    }
  };

  const handleRetry = () => {
    setLocalError(null);
    setIsCompleting(true);
    // The useEffect will automatically retry
  };

  const completedItems = [
    t('completeStep.success.completedItems.database'),
    t('completeStep.success.completedItems.admin'),
    t('completeStep.success.completedItems.company'),
    t('completeStep.success.completedItems.settings'),
    t('completeStep.success.completedItems.namespace'),
    t('completeStep.success.completedItems.security'),
  ];

  // Show loading state while completing setup
  if (isCompleting) {
    return (
      <div className="p-8 text-center">
        <div className="mb-8">
          <div className="animate-spin rounded-full h-20 w-20 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <h2 className="text-2xl font-bold text-foreground mb-2">
            {t('completeStep.finalizing.title')}
          </h2>
          <p className="text-muted-foreground">{t('completeStep.finalizing.description')}</p>
        </div>
      </div>
    );
  }

  // Show error state if setup failed
  if (localError || setupError) {
    return (
      <div className="p-8 text-center">
        <div className="mb-8">
          <div className="h-20 w-20 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
            <svg
              className="h-10 w-10 text-red-600"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </div>
          <h2 className="text-2xl font-bold text-foreground mb-2">
            {t('completeStep.failed.title')}
          </h2>
          <p className="text-muted-foreground mb-6">{localError || setupError}</p>
          <button
            onClick={handleRetry}
            className="bg-primary-600 text-white px-6 py-2 rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 transition-colors"
          >
            {t('completeStep.failed.retry')}
          </button>
        </div>
      </div>
    );
  }

  // Show success state when setup is complete
  return (
    <div className="p-8 text-center">
      <div className="mb-8">
        <CheckCircleIcon className="h-20 w-20 text-success mx-auto mb-4" />
        <h2 className="text-3xl font-bold text-foreground mb-2">
          {t('completeStep.success.title')}
        </h2>
        <p className="text-lg text-muted-foreground">{t('completeStep.success.description')}</p>
      </div>

      <div className="max-w-md mx-auto mb-8">
        <div className="bg-success/10 border border-success/20 rounded-lg p-6">
          <h3 className="text-lg font-medium text-success-foreground mb-4">
            {t('completeStep.success.completedTitle')}
          </h3>
          <ul className="text-sm text-success-foreground space-y-2">
            {completedItems.map((item, index) => (
              <li key={index} className="flex items-center">
                <CheckCircleIcon className="h-4 w-4 text-success mr-2 flex-shrink-0" />
                {item}
              </li>
            ))}
          </ul>
        </div>
      </div>

      <div className="space-y-4">
        <div className="bg-primary/10 border border-primary/20 rounded-lg p-4">
          <p className="text-sm text-primary-700">
            <strong>{t('completeStep.success.nextSteps')}</strong>{' '}
            {isAuthenticated
              ? t('completeStep.success.nextStepsAuth')
              : t('completeStep.success.nextStepsLogin')}
          </p>
        </div>

        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <button
            onClick={handleGoToDashboard}
            className="inline-flex items-center px-6 py-3 bg-primary-600 text-white rounded-md hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 transition-colors"
          >
            {isAuthenticated && isSetupFinished
              ? t('completeStep.success.goToDashboard')
              : t('completeStep.success.goToLogin')}
            <ArrowRightIcon className="ml-2 h-5 w-5" />
          </button>
        </div>

        {isAuthenticated && isSetupFinished && (
          <p className="text-sm text-muted-foreground">{t('completeStep.success.autoRedirect')}</p>
        )}
      </div>

      <div className="mt-8 pt-8 border-t border-border">
        <div className="text-sm text-muted-foreground">
          <p className="mb-2 text-foreground">{t('completeStep.success.helpTitle')}</p>
          <div className="flex justify-center space-x-4">
            <a href="#" className="text-primary-600 hover:text-primary-500">
              {t('completeStep.success.helpLinks.documentation')}
            </a>
            <span>·</span>
            <a href="#" className="text-primary-600 hover:text-primary-500">
              {t('completeStep.success.helpLinks.apiGuide')}
            </a>
            <span>·</span>
            <a href="#" className="text-primary-600 hover:text-primary-500">
              {t('completeStep.success.helpLinks.support')}
            </a>
          </div>
        </div>
      </div>
    </div>
  );
}
