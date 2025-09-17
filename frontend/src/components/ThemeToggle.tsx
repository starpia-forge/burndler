import { useState, useRef, useEffect } from 'react';
import { useTheme } from '../hooks/useTheme';

const ThemeToggle = () => {
  const { setThemeMode, isLightMode, isDarkMode, isSystemMode } = useTheme();
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Handle keyboard events
  const handleKeyDown = (event: React.KeyboardEvent) => {
    if (event.key === 'Escape') {
      setIsOpen(false);
    } else if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      setIsOpen(!isOpen);
    }
  };

  const handleOptionClick = (mode: 'light' | 'dark' | 'system') => {
    setThemeMode(mode);
    setIsOpen(false);
  };

  // Get current icon based on theme mode
  const getCurrentIcon = () => {
    if (isLightMode) {
      return <SunIcon />;
    } else if (isDarkMode) {
      return <MoonIcon />;
    } else if (isSystemMode) {
      return <ComputerIcon />;
    }
    return <SunIcon />;
  };

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        type="button"
        className="p-2 rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground transition-colors"
        aria-label="Toggle theme"
        aria-haspopup="true"
        aria-expanded={isOpen}
        onClick={() => setIsOpen(!isOpen)}
        onKeyDown={handleKeyDown}
      >
        {getCurrentIcon()}
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-32 bg-popover border border-border rounded-md shadow-lg z-50">
          <div className="py-1">
            <button
              type="button"
              className={`w-full text-left px-3 py-2 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2 ${
                isLightMode ? 'bg-accent' : ''
              }`}
              onClick={() => handleOptionClick('light')}
            >
              <SunIcon />
              Light
            </button>
            <button
              type="button"
              className={`w-full text-left px-3 py-2 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2 ${
                isDarkMode ? 'bg-accent' : ''
              }`}
              onClick={() => handleOptionClick('dark')}
            >
              <MoonIcon />
              Dark
            </button>
            <button
              type="button"
              className={`w-full text-left px-3 py-2 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2 ${
                isSystemMode ? 'bg-accent' : ''
              }`}
              onClick={() => handleOptionClick('system')}
            >
              <ComputerIcon />
              System
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

// Icon components with test IDs for testing
const SunIcon = () => (
  <svg
    data-testid="sun-icon"
    xmlns="http://www.w3.org/2000/svg"
    width="16"
    height="16"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <circle cx="12" cy="12" r="4" />
    <path d="M12 2v2" />
    <path d="M12 20v2" />
    <path d="m4.93 4.93 1.41 1.41" />
    <path d="m17.66 17.66 1.41 1.41" />
    <path d="M2 12h2" />
    <path d="M20 12h2" />
    <path d="m6.34 17.66-1.41-1.41" />
    <path d="m19.07 4.93-1.41-1.41" />
  </svg>
);

const MoonIcon = () => (
  <svg
    data-testid="moon-icon"
    xmlns="http://www.w3.org/2000/svg"
    width="16"
    height="16"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z" />
  </svg>
);

const ComputerIcon = () => (
  <svg
    data-testid="computer-icon"
    xmlns="http://www.w3.org/2000/svg"
    width="16"
    height="16"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <rect width="14" height="8" x="5" y="2" rx="2" />
    <rect width="20" height="8" x="2" y="14" rx="2" />
    <path d="M6 18h2" />
    <path d="M12 18h6" />
  </svg>
);

export default ThemeToggle;
