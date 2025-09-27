import React, { useState, useCallback, useEffect } from 'react';
import {
  MagnifyingGlassIcon,
  XMarkIcon,
  AdjustmentsHorizontalIcon,
} from '@heroicons/react/24/outline';
import { ModuleFilters as ModuleFiltersType } from '../../types/module';
import { useAuth } from '../../hooks/useAuth';

interface ModuleFiltersProps {
  filters: ModuleFiltersType;
  onFiltersChange: (filters: ModuleFiltersType) => void;
  onClearFilters: () => void;
  loading?: boolean;
  className?: string;
}

const AUTHOR_SUGGESTIONS = [
  'System Admin',
  'DevOps Team',
  'Backend Team',
  'Frontend Team',
  'Infrastructure Team',
];

export const ModuleFilters: React.FC<ModuleFiltersProps> = ({
  filters,
  onFiltersChange,
  onClearFilters,
  loading = false,
  className = '',
}) => {
  const { isDeveloper } = useAuth();
  const [searchTerm, setSearchTerm] = useState(filters.search || '');
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [showAuthorSuggestions, setShowAuthorSuggestions] = useState(false);

  // Debounced search
  useEffect(() => {
    const timer = setTimeout(() => {
      onFiltersChange({ ...filters, search: searchTerm });
    }, 300);

    return () => clearTimeout(timer);
  }, [searchTerm]);

  const handleFilterChange = useCallback(
    (key: keyof ModuleFiltersType, value: any) => {
      onFiltersChange({ ...filters, [key]: value });
    },
    [filters, onFiltersChange]
  );

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value);
  };

  const handleAuthorSelect = (author: string) => {
    handleFilterChange('author', author);
    setShowAuthorSuggestions(false);
  };

  const getActiveFilterCount = () => {
    let count = 0;
    if (filters.search) count++;
    if (filters.author) count++;
    if (filters.active !== undefined) count++;
    if (filters.published_only) count++;
    if (filters.show_deleted) count++;
    return count;
  };

  const activeFilterCount = getActiveFilterCount();

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Search and primary controls */}
      <div className="flex flex-col sm:flex-row gap-4">
        {/* Search input */}
        <div className="flex-1 relative">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <MagnifyingGlassIcon className="h-5 w-5 text-gray-400" />
          </div>
          <input
            type="text"
            placeholder="Search modules by name, description, or author..."
            value={searchTerm}
            onChange={handleSearchChange}
            disabled={loading}
            className="block w-full pl-10 pr-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md leading-5 bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:opacity-50"
          />
        </div>

        {/* Quick status filter */}
        <div className="flex items-center space-x-2">
          <select
            value={filters.active === undefined ? 'all' : filters.active ? 'active' : 'inactive'}
            onChange={(e) => {
              const value = e.target.value;
              handleFilterChange('active', value === 'all' ? undefined : value === 'active');
            }}
            disabled={loading}
            className="block px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:opacity-50"
          >
            <option value="all">All Modules</option>
            <option value="active">Active Only</option>
            <option value="inactive">Inactive Only</option>
          </select>

          {/* Advanced filters toggle */}
          <button
            onClick={() => setShowAdvanced(!showAdvanced)}
            className={`relative inline-flex items-center px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm font-medium transition-colors ${
              showAdvanced || activeFilterCount > 0
                ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300 border-blue-300 dark:border-blue-600'
                : 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700'
            } disabled:opacity-50`}
            disabled={loading}
          >
            <AdjustmentsHorizontalIcon className="h-4 w-4 mr-1" />
            Filters
            {activeFilterCount > 0 && (
              <span className="ml-1 bg-blue-600 dark:bg-blue-500 text-white text-xs px-1.5 py-0.5 rounded-full min-w-[1rem] text-center">
                {activeFilterCount}
              </span>
            )}
          </button>

          {/* Clear filters */}
          {activeFilterCount > 0 && (
            <button
              onClick={onClearFilters}
              disabled={loading}
              className="inline-flex items-center px-2 py-2 text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 disabled:opacity-50"
              title="Clear all filters"
            >
              <XMarkIcon className="h-4 w-4" />
            </button>
          )}
        </div>
      </div>

      {/* Advanced filters */}
      {showAdvanced && (
        <div className="bg-gray-50 dark:bg-gray-800/50 rounded-lg p-4 border border-gray-200 dark:border-gray-700">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {/* Author filter */}
            <div className="relative">
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Author
              </label>
              <input
                type="text"
                placeholder="Filter by author..."
                value={filters.author || ''}
                onChange={(e) => handleFilterChange('author', e.target.value)}
                onFocus={() => setShowAuthorSuggestions(true)}
                onBlur={() => setTimeout(() => setShowAuthorSuggestions(false), 200)}
                disabled={loading}
                className="block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:opacity-50"
              />

              {/* Author suggestions */}
              {showAuthorSuggestions && (
                <div className="absolute z-10 mt-1 w-full bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md shadow-lg max-h-40 overflow-auto">
                  {AUTHOR_SUGGESTIONS.filter(
                    (author) =>
                      !filters.author || author.toLowerCase().includes(filters.author.toLowerCase())
                  ).map((author) => (
                    <button
                      key={author}
                      onClick={() => handleAuthorSelect(author)}
                      className="block w-full text-left px-3 py-2 text-sm text-gray-900 dark:text-white hover:bg-gray-100 dark:hover:bg-gray-700 focus:outline-none focus:bg-gray-100 dark:focus:bg-gray-700"
                    >
                      {author}
                    </button>
                  ))}
                </div>
              )}
            </div>

            {/* Published filter */}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Version Status
              </label>
              <select
                value={filters.published_only ? 'published' : 'all'}
                onChange={(e) =>
                  handleFilterChange('published_only', e.target.value === 'published')
                }
                disabled={loading}
                className="block w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 disabled:opacity-50"
              >
                <option value="all">All Versions</option>
                <option value="published">Published Only</option>
              </select>
            </div>

            {/* Show deleted (Developer only) */}
            {isDeveloper && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Visibility
                </label>
                <label className="flex items-center space-x-2 pt-2">
                  <input
                    type="checkbox"
                    checked={filters.show_deleted || false}
                    onChange={(e) => handleFilterChange('show_deleted', e.target.checked)}
                    disabled={loading}
                    className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 dark:border-gray-600 rounded disabled:opacity-50"
                  />
                  <span className="text-sm text-gray-700 dark:text-gray-300">
                    Show deleted modules
                  </span>
                </label>
              </div>
            )}
          </div>

          {/* Filter summary */}
          {activeFilterCount > 0 && (
            <div className="mt-4 pt-4 border-t border-gray-200 dark:border-gray-600">
              <div className="flex items-center justify-between">
                <div className="text-sm text-gray-600 dark:text-gray-400">
                  {activeFilterCount} filter{activeFilterCount !== 1 ? 's' : ''} applied
                </div>
                <button
                  onClick={onClearFilters}
                  disabled={loading}
                  className="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 disabled:opacity-50"
                >
                  Clear all filters
                </button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default ModuleFilters;
