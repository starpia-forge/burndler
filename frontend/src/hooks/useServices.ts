import { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import serviceService from '../services/serviceService';
import {
  Service,
  ServiceListResponse,
  ServiceFilters,
  ServiceListState,
  ApiError,
} from '../types/service';

export interface UseServicesOptions {
  autoFetch?: boolean;
  initialFilters?: ServiceFilters;
}

export interface UseServicesReturn {
  services: Service[];
  loading: boolean;
  initialLoading: boolean;
  isRefreshing: boolean;
  error: string | null;
  pagination: ServiceListResponse['pagination'] | null;
  filters: ServiceFilters;
  refetch: () => Promise<void>;
  setFilters: (filters: ServiceFilters) => void;
  updateFilter: (key: keyof ServiceFilters, value: any) => void;
  clearFilters: () => void;
  deleteService: (id: number) => Promise<void>;
  refreshService: (id: number) => Promise<void>;
}

const DEFAULT_FILTERS: ServiceFilters = {
  page: 1,
  page_size: 10,
  active: undefined,
  search: '',
};

export function useServices(options: UseServicesOptions = {}): UseServicesReturn {
  const { autoFetch = true, initialFilters = {} } = options;

  const [state, setState] = useState<ServiceListState>({
    services: [],
    loading: false,
    initialLoading: true,
    isRefreshing: false,
    error: null,
    pagination: null,
    filters: { ...DEFAULT_FILTERS, ...initialFilters },
  });

  // Use ref to maintain stable reference to current filters
  const filtersRef = useRef(state.filters);

  // Development safety: Track fetch calls to detect infinite loops
  const fetchCountRef = useRef(0);
  const lastFetchTimeRef = useRef(0);

  // Track if this is the first fetch to determine loading state
  const isFirstFetchRef = useRef(true);

  // Update ref when filters change
  filtersRef.current = state.filters;

  // Fetch services based on current filters (using ref to break circular dependency)
  const fetchServices = useCallback(async () => {
    // Development safety: Detect potential infinite loops
    if (process.env.NODE_ENV === 'development') {
      const now = Date.now();
      const timeSinceLastFetch = now - lastFetchTimeRef.current;

      if (timeSinceLastFetch < 100) {
        // Less than 100ms since last fetch
        fetchCountRef.current += 1;
        if (fetchCountRef.current > 10) {
          console.error(
            '🚨 INFINITE LOOP DETECTED: useServices fetchServices called more than 10 times in quick succession!'
          );
          console.error('Current filters:', filtersRef.current);
          return; // Prevent further execution
        }
      } else {
        fetchCountRef.current = 1; // Reset counter if enough time has passed
      }

      lastFetchTimeRef.current = now;
    }

    // Determine if this is initial loading or refresh using ref
    const isInitialLoad = isFirstFetchRef.current;
    isFirstFetchRef.current = false;

    setState((prev) => ({
      ...prev,
      loading: true,
      initialLoading: isInitialLoad,
      isRefreshing: !isInitialLoad,
      error: null,
    }));

    try {
      const response = await serviceService.listServices(filtersRef.current);
      setState((prev) => ({
        ...prev,
        services: response.data,
        pagination: response.pagination,
        loading: false,
        initialLoading: false,
        isRefreshing: false,
      }));
    } catch (error: any) {
      const apiError = error as ApiError;
      setState((prev) => ({
        ...prev,
        error: apiError.message || 'Failed to fetch services',
        loading: false,
        initialLoading: false,
        isRefreshing: false,
      }));
    }
  }, []); // No dependencies - breaks circular dependency

  // Set filters and trigger refetch
  const setFilters = useCallback((newFilters: ServiceFilters) => {
    setState((prev) => ({
      ...prev,
      filters: { ...prev.filters, ...newFilters, page: 1 }, // Reset to page 1 when filters change
    }));
  }, []);

  // Update single filter
  const updateFilter = useCallback((key: keyof ServiceFilters, value: any) => {
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
    await fetchServices();
  }, [fetchServices]);

  // Delete service and update local state
  const deleteService = useCallback(
    async (id: number) => {
      try {
        await serviceService.deleteService(id);

        // Remove from local state
        setState((prev) => ({
          ...prev,
          services: prev.services.filter((service) => service.id !== id),
        }));
      } catch (error: any) {
        const apiError = error as ApiError;
        setState((prev) => ({
          ...prev,
          error: apiError.message || 'Failed to delete service',
        }));
        throw error; // Re-throw for component error handling
      }
    },
    []
  );

  // Refresh single service
  const refreshService = useCallback(async (id: number) => {
    try {
      const updatedService = await serviceService.getService(id);
      setState((prev) => ({
        ...prev,
        services: prev.services.map((service) =>
          service.id === id ? updatedService : service
        ),
      }));
    } catch (error: any) {
      // If service not found (404), remove from list
      if (error.status === 404) {
        setState((prev) => ({
          ...prev,
          services: prev.services.filter((service) => service.id !== id),
        }));
      } else {
        const apiError = error as ApiError;
        setState((prev) => ({
          ...prev,
          error: apiError.message || 'Failed to refresh service',
        }));
      }
    }
  }, []);

  // Auto-fetch on mount and when filters change
  useEffect(() => {
    if (autoFetch) {
      fetchServices();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [autoFetch, state.filters]); // Depend on filters directly, not fetchServices (intentionally breaking circular dependency)

  // Memoized return value to prevent unnecessary re-renders
  const returnValue = useMemo(
    () => ({
      services: state.services,
      loading: state.loading,
      initialLoading: state.initialLoading,
      isRefreshing: state.isRefreshing,
      error: state.error,
      pagination: state.pagination,
      filters: state.filters,
      refetch,
      setFilters,
      updateFilter,
      clearFilters,
      deleteService,
      refreshService,
    }),
    [
      state.services,
      state.loading,
      state.initialLoading,
      state.isRefreshing,
      state.error,
      state.pagination,
      state.filters,
      refetch,
      setFilters,
      updateFilter,
      clearFilters,
      deleteService,
      refreshService,
    ]
  );

  return returnValue;
}

// Helper hook for searching services
export function useServiceSearch(services: Service[], searchTerm: string) {
  return useMemo(() => {
    if (!searchTerm.trim()) return services;

    const lowerSearchTerm = searchTerm.toLowerCase();
    return services.filter(
      (service) =>
        service.name.toLowerCase().includes(lowerSearchTerm) ||
        service.description.toLowerCase().includes(lowerSearchTerm)
    );
  }, [services, searchTerm]);
}

// Helper hook for sorting services
export function useServiceSort(
  services: Service[],
  sortBy: string,
  sortDirection: 'asc' | 'desc'
) {
  return useMemo(() => {
    const sortedServices = [...services].sort((a, b) => {
      let aValue: any;
      let bValue: any;

      switch (sortBy) {
        case 'name':
          aValue = a.name.toLowerCase();
          bValue = b.name.toLowerCase();
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

    return sortedServices;
  }, [services, sortBy, sortDirection]);
}