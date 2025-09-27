import { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import moduleService from '../services/moduleService';
import {
  Module,
  ModuleListResponse,
  ModuleFilters,
  ModuleListState,
  ApiError,
} from '../types/module';

export interface UseModulesOptions {
  autoFetch?: boolean;
  initialFilters?: ModuleFilters;
}

export interface UseModulesReturn {
  modules: Module[];
  loading: boolean;
  error: string | null;
  pagination: ModuleListResponse['pagination'] | null;
  filters: ModuleFilters;
  refetch: () => Promise<void>;
  setFilters: (filters: ModuleFilters) => void;
  updateFilter: (key: keyof ModuleFilters, value: any) => void;
  clearFilters: () => void;
  deleteModule: (id: number) => Promise<void>;
  refreshModule: (id: number) => Promise<void>;
}

const DEFAULT_FILTERS: ModuleFilters = {
  page: 1,
  page_size: 10,
  active: undefined,
  author: '',
  show_deleted: false,
  published_only: false,
  search: '',
};

export function useModules(options: UseModulesOptions = {}): UseModulesReturn {
  const { autoFetch = true, initialFilters = {} } = options;

  const [state, setState] = useState<ModuleListState>({
    modules: [],
    loading: false,
    error: null,
    pagination: null,
    filters: { ...DEFAULT_FILTERS, ...initialFilters },
  });

  // Use ref to maintain stable reference to current filters
  const filtersRef = useRef(state.filters);

  // Development safety: Track fetch calls to detect infinite loops
  const fetchCountRef = useRef(0);
  const lastFetchTimeRef = useRef(0);

  // Update ref when filters change
  filtersRef.current = state.filters;

  // Fetch modules based on current filters (using ref to break circular dependency)
  const fetchModules = useCallback(async () => {
    // Development safety: Detect potential infinite loops
    if (process.env.NODE_ENV === 'development') {
      const now = Date.now();
      const timeSinceLastFetch = now - lastFetchTimeRef.current;

      if (timeSinceLastFetch < 100) {
        // Less than 100ms since last fetch
        fetchCountRef.current += 1;
        if (fetchCountRef.current > 10) {
          console.error(
            '🚨 INFINITE LOOP DETECTED: useModules fetchModules called more than 10 times in quick succession!'
          );
          console.error('Current filters:', filtersRef.current);
          return; // Prevent further execution
        }
      } else {
        fetchCountRef.current = 1; // Reset counter if enough time has passed
      }

      lastFetchTimeRef.current = now;
    }

    setState((prev) => ({ ...prev, loading: true, error: null }));

    try {
      const response = await moduleService.listModules(filtersRef.current);
      setState((prev) => ({
        ...prev,
        modules: response.data,
        pagination: response.pagination,
        loading: false,
      }));
    } catch (error: any) {
      const apiError = error as ApiError;
      setState((prev) => ({
        ...prev,
        error: apiError.message || 'Failed to fetch modules',
        loading: false,
      }));
    }
  }, []); // No dependencies - breaks circular dependency

  // Set filters and trigger refetch
  const setFilters = useCallback((newFilters: ModuleFilters) => {
    setState((prev) => ({
      ...prev,
      filters: { ...prev.filters, ...newFilters, page: 1 }, // Reset to page 1 when filters change
    }));
  }, []);

  // Update single filter
  const updateFilter = useCallback((key: keyof ModuleFilters, value: any) => {
    setState((prev) => ({
      ...prev,
      filters: { ...prev.filters, [key]: value, page: 1 }, // Reset to page 1 when filter changes
    }));
  }, []);

  // Clear all filters
  const clearFilters = useCallback(() => {
    setState((prev) => ({
      ...prev,
      filters: DEFAULT_FILTERS,
    }));
  }, []);

  // Refetch with current filters
  const refetch = useCallback(async () => {
    await fetchModules();
  }, [fetchModules]);

  // Delete module and update local state
  const deleteModule = useCallback(
    async (id: number) => {
      try {
        await moduleService.deleteModule(id);

        // If showing deleted modules, refetch to update the list
        if (filtersRef.current.show_deleted) {
          await fetchModules();
        } else {
          // Otherwise, remove from local state
          setState((prev) => ({
            ...prev,
            modules: prev.modules.filter((module) => module.id !== id),
          }));
        }
      } catch (error: any) {
        const apiError = error as ApiError;
        setState((prev) => ({
          ...prev,
          error: apiError.message || 'Failed to delete module',
        }));
        throw error; // Re-throw for component error handling
      }
    },
    [fetchModules] // Keep fetchModules dependency but it's now stable
  );

  // Refresh single module
  const refreshModule = useCallback(async (id: number) => {
    try {
      const updatedModule = await moduleService.getModule(id);
      setState((prev) => ({
        ...prev,
        modules: prev.modules.map((module) => (module.id === id ? updatedModule : module)),
      }));
    } catch (error: any) {
      // If module not found (404), remove from list
      if (error.status === 404) {
        setState((prev) => ({
          ...prev,
          modules: prev.modules.filter((module) => module.id !== id),
        }));
      } else {
        const apiError = error as ApiError;
        setState((prev) => ({
          ...prev,
          error: apiError.message || 'Failed to refresh module',
        }));
      }
    }
  }, []);

  // Auto-fetch on mount and when filters change
  useEffect(() => {
    if (autoFetch) {
      fetchModules();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [autoFetch, state.filters]); // Depend on filters directly, not fetchModules (intentionally breaking circular dependency)

  // Memoized return value to prevent unnecessary re-renders
  const returnValue = useMemo(
    () => ({
      modules: state.modules,
      loading: state.loading,
      error: state.error,
      pagination: state.pagination,
      filters: state.filters,
      refetch,
      setFilters,
      updateFilter,
      clearFilters,
      deleteModule,
      refreshModule,
    }),
    [
      state.modules,
      state.loading,
      state.error,
      state.pagination,
      state.filters,
      refetch,
      setFilters,
      updateFilter,
      clearFilters,
      deleteModule,
      refreshModule,
    ]
  );

  return returnValue;
}

// Helper hook for searching modules
export function useModuleSearch(modules: Module[], searchTerm: string) {
  return useMemo(() => {
    if (!searchTerm.trim()) return modules;

    const lowerSearchTerm = searchTerm.toLowerCase();
    return modules.filter(
      (module) =>
        module.name.toLowerCase().includes(lowerSearchTerm) ||
        module.description.toLowerCase().includes(lowerSearchTerm) ||
        module.author.toLowerCase().includes(lowerSearchTerm)
    );
  }, [modules, searchTerm]);
}

// Helper hook for sorting modules
export function useModuleSort(modules: Module[], sortBy: string, sortDirection: 'asc' | 'desc') {
  return useMemo(() => {
    const sortedModules = [...modules].sort((a, b) => {
      let aValue: any;
      let bValue: any;

      switch (sortBy) {
        case 'name':
          aValue = a.name.toLowerCase();
          bValue = b.name.toLowerCase();
          break;
        case 'author':
          aValue = a.author.toLowerCase();
          bValue = b.author.toLowerCase();
          break;
        case 'created_at':
          aValue = new Date(a.created_at);
          bValue = new Date(b.created_at);
          break;
        case 'updated_at':
          aValue = new Date(a.updated_at);
          bValue = new Date(b.updated_at);
          break;
        default:
          return 0;
      }

      if (aValue < bValue) return sortDirection === 'asc' ? -1 : 1;
      if (aValue > bValue) return sortDirection === 'asc' ? 1 : -1;
      return 0;
    });

    return sortedModules;
  }, [modules, sortBy, sortDirection]);
}
