import '@testing-library/jest-dom';
import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';

// Initialize i18n for testing
i18n.use(initReactI18next).init({
  lng: 'en',
  fallbackLng: 'en',
  ns: ['common', 'auth', 'modules', 'projects', 'setup', 'errors'],
  defaultNS: 'common',
  resources: {
    en: {
      common: {
        appName: 'Burndler',
        loading: 'Loading...',
        theme: {
          light: 'Light',
          dark: 'Dark',
          system: 'System',
        },
      },
      auth: {
        signIn: 'Sign In',
        signOut: 'Sign Out',
        profile: 'Profile',
        settings: 'Settings',
        email: 'Email',
        password: 'Password',
        signInToAccount: 'Sign in to your account',
        signInButton: 'Sign in',
        enterEmail: 'Enter your email',
        enterPassword: 'Enter your password',
        emailRequired: 'Email is required',
        passwordRequired: 'Password is required',
        invalidEmail: 'Please enter a valid email address',
        invalidCredentials: 'Invalid email or password',
        passwordMinLength: 'Password must be at least 6 characters',
        loginFailed: 'Login failed. Please try again.',
        role: {
          developer: 'Developer',
          engineer: 'Engineer',
        },
      },
    },
  },
  interpolation: {
    escapeValue: false,
  },
});

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  constructor(cb: any) {
    this.cb = cb;
  }
  cb: any;
  observe() {}
  unobserve() {}
  disconnect() {}
};

// Mock scrollIntoView
Element.prototype.scrollIntoView = () => {};

// Mock matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => {},
  }),
});
