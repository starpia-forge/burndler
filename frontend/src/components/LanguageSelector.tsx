import { Fragment } from 'react';
import { Menu, Transition } from '@headlessui/react';
import { LanguageIcon, ChevronDownIcon } from '@heroicons/react/24/outline';
import { useTranslation } from 'react-i18next';
import { LANGUAGES, type Language } from '../i18n/types';

export default function LanguageSelector() {
  const { i18n } = useTranslation();

  const currentLanguage = LANGUAGES.find((lang) => lang.code === i18n.language) || LANGUAGES[0];

  const handleLanguageChange = (languageCode: Language) => {
    i18n.changeLanguage(languageCode);
  };

  return (
    <Menu as="div" className="relative">
      <Menu.Button className="flex items-center space-x-2 px-3 py-2 rounded-lg hover:bg-accent transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500">
        <LanguageIcon className="w-5 h-5 text-muted-foreground" />
        <span className="text-sm font-medium text-foreground hidden sm:block">
          {currentLanguage.nativeName}
        </span>
        <ChevronDownIcon className="w-4 h-4 text-muted-foreground" />
      </Menu.Button>

      <Transition
        as={Fragment}
        enter="transition ease-out duration-100"
        enterFrom="transform opacity-0 scale-95"
        enterTo="transform opacity-100 scale-100"
        leave="transition ease-in duration-75"
        leaveFrom="transform opacity-100 scale-100"
        leaveTo="transform opacity-0 scale-95"
      >
        <Menu.Items className="absolute right-0 mt-2 w-48 rounded-lg bg-popover border border-border shadow-lg focus:outline-none z-50">
          <div className="p-1">
            {LANGUAGES.map((language) => (
              <Menu.Item key={language.code}>
                {({ active }) => (
                  <button
                    onClick={() => handleLanguageChange(language.code)}
                    className={`${active ? 'bg-accent' : ''} ${
                      i18n.language === language.code
                        ? 'bg-primary/10 text-primary'
                        : 'text-foreground'
                    } group flex w-full items-center rounded-md px-3 py-2 text-sm transition-colors`}
                  >
                    <div className="flex items-center justify-between w-full">
                      <div className="flex flex-col items-start">
                        <span className="font-medium">{language.nativeName}</span>
                        <span className="text-xs text-muted-foreground">{language.name}</span>
                      </div>
                      {i18n.language === language.code && (
                        <div className="w-2 h-2 bg-primary rounded-full"></div>
                      )}
                    </div>
                  </button>
                )}
              </Menu.Item>
            ))}
          </div>
        </Menu.Items>
      </Transition>
    </Menu>
  );
}
