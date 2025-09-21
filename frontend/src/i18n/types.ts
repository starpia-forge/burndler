import 'react-i18next';

// Import the English translations to use as type source
import commonEn from '../locales/en/common.json';
import authEn from '../locales/en/auth.json';
import modulesEn from '../locales/en/modules.json';
import projectsEn from '../locales/en/projects.json';
import setupEn from '../locales/en/setup.json';
import errorsEn from '../locales/en/errors.json';

// Define the resources type
declare module 'react-i18next' {
  interface CustomTypeOptions {
    defaultNS: 'common';
    resources: {
      common: typeof commonEn;
      auth: typeof authEn;
      modules: typeof modulesEn;
      projects: typeof projectsEn;
      setup: typeof setupEn;
      errors: typeof errorsEn;
    };
  }
}

// Language types
export type Language = 'en' | 'ko';

export interface LanguageOption {
  code: Language;
  name: string;
  nativeName: string;
}

export const LANGUAGES: LanguageOption[] = [
  { code: 'en', name: 'English', nativeName: 'English' },
  { code: 'ko', name: 'Korean', nativeName: '한국어' },
];
