import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { LanguageIcon, CheckIcon } from '@heroicons/react/24/outline';
import { LANGUAGES, type Language } from '../../i18n/types';

interface SystemLanguageProps {
  onContinue: () => void;
}

export default function SystemLanguage({ onContinue }: SystemLanguageProps) {
  const { t, i18n } = useTranslation('setup');
  const [selectedLanguage, setSelectedLanguage] = useState<Language>(
    (i18n.language as Language) || 'en'
  );

  const handleLanguageSelect = (languageCode: Language) => {
    setSelectedLanguage(languageCode);
    i18n.changeLanguage(languageCode);
  };

  const handleContinue = () => {
    onContinue();
  };

  return (
    <div className="p-8">
      <div className="text-center mb-8">
        <LanguageIcon className="h-16 w-16 text-primary-600 mx-auto mb-4" />
        <h2 className="text-2xl font-bold text-foreground mb-2">{t('languageStep.title')}</h2>
        <p className="text-muted-foreground">{t('languageStep.description')}</p>
      </div>

      <div className="max-w-md mx-auto space-y-4 mb-8">
        <h3 className="text-lg font-medium text-foreground text-center mb-6">
          {t('languageStep.selectLanguage')}
        </h3>

        <div className="space-y-3">
          {LANGUAGES.map((language) => (
            <button
              key={language.code}
              onClick={() => handleLanguageSelect(language.code)}
              className={`
                w-full p-4 rounded-lg border-2 transition-all duration-200 hover:shadow-md
                ${
                  selectedLanguage === language.code
                    ? 'border-primary-600 bg-primary/10 shadow-sm'
                    : 'border-border bg-card hover:border-primary/30'
                }
              `}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <div className="text-left">
                    <div className="font-medium text-foreground">{language.nativeName}</div>
                    <div className="text-sm text-muted-foreground">{language.name}</div>
                  </div>
                </div>
                {selectedLanguage === language.code && (
                  <div className="flex-shrink-0">
                    <div className="h-6 w-6 bg-primary-600 rounded-full flex items-center justify-center">
                      <CheckIcon className="h-4 w-4 text-white" />
                    </div>
                  </div>
                )}
              </div>
            </button>
          ))}
        </div>
      </div>

      <div className="flex justify-center">
        <button
          onClick={handleContinue}
          className="px-6 py-3 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2"
        >
          {t('languageStep.continue')}
        </button>
      </div>
    </div>
  );
}
