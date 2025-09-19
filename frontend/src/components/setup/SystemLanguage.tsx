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
        <LanguageIcon className="h-16 w-16 text-blue-600 mx-auto mb-4" />
        <h2 className="text-2xl font-bold text-gray-900 mb-2">{t('languageStep.title')}</h2>
        <p className="text-gray-600">{t('languageStep.description')}</p>
      </div>

      <div className="max-w-md mx-auto space-y-4 mb-8">
        <h3 className="text-lg font-medium text-gray-900 text-center mb-6">
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
                    ? 'border-blue-600 bg-blue-50 shadow-sm'
                    : 'border-gray-200 bg-white hover:border-gray-300'
                }
              `}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <div className="text-left">
                    <div className="font-medium text-gray-900">{language.nativeName}</div>
                    <div className="text-sm text-gray-500">{language.name}</div>
                  </div>
                </div>
                {selectedLanguage === language.code && (
                  <div className="flex-shrink-0">
                    <div className="h-6 w-6 bg-blue-600 rounded-full flex items-center justify-center">
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
          className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
        >
          {t('languageStep.continue')}
        </button>
      </div>
    </div>
  );
}
