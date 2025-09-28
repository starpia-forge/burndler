import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import HttpBackend from 'i18next-http-backend';

// Import translation files for better bundling
import commonEn from '../locales/en/common.json';
import authEn from '../locales/en/auth.json';
import containersEn from '../locales/en/containers.json';
import projectsEn from '../locales/en/projects.json';
import setupEn from '../locales/en/setup.json';
import errorsEn from '../locales/en/errors.json';

import commonKo from '../locales/ko/common.json';
import authKo from '../locales/ko/auth.json';
import containersKo from '../locales/ko/containers.json';
import projectsKo from '../locales/ko/projects.json';
import setupKo from '../locales/ko/setup.json';
import errorsKo from '../locales/ko/errors.json';

const resources = {
  en: {
    common: commonEn,
    auth: authEn,
    containers: containersEn,
    projects: projectsEn,
    setup: setupEn,
    errors: errorsEn,
  },
  ko: {
    common: commonKo,
    auth: authKo,
    containers: containersKo,
    projects: projectsKo,
    setup: setupKo,
    errors: errorsKo,
  },
};

i18n
  .use(HttpBackend)
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'en',
    supportedLngs: ['en', 'ko'],

    // Namespace configuration
    ns: ['common', 'auth', 'containers', 'projects', 'setup', 'errors'],
    defaultNS: 'common',

    // Language detection options
    detection: {
      order: ['localStorage', 'navigator', 'htmlTag'],
      caches: ['localStorage'],
      lookupLocalStorage: 'burndler-language',
    },

    interpolation: {
      escapeValue: false, // React already escapes values
    },

    // Development settings
    debug: process.env.NODE_ENV === 'development',

    // Backend options (for dynamic loading if needed)
    backend: {
      loadPath: '/locales/{{lng}}/{{ns}}.json',
    },
  });

export default i18n;
