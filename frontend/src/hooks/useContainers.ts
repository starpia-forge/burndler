import { useState, useEffect, useCallback, useMemo, useRef } from 'react';
import containerService from '../services/containerService';
import {
  Container,
  ContainerListResponse,
  ContainerFilters,
  ContainerListState,
  ApiError,
} from '../types/container';

export interface UseContainersOptions {
  autoFetch?: boolean;
  initialFilters?: ContainerFilters;
}

export interface UseContainersReturn {
  containers: Container[];
  loading: boolean;
  initialLoading: boolean;
  isRefreshing: boolean;
  error: string | null;
  pagination: ContainerListResponse['pagination'] | null;
  filters: ContainerFilters;
  refetch: () => Promise<void>;
  setFilters: (filters: ContainerFilters) => void;
  updateFilter: (key: keyof ContainerFilters, value: any) => void;
  clearFilters: () => void;
  deleteContainer: (id: number) => Promise<void>;
  refreshContainer: (id: number) => Promise<void>;
}

const DEFAULT_FILTERS: ContainerFilters = {
  page: 1,
  page_size: 10,
  active: undefined,
  author: '',
  show_deleted: false,
  published_only: false,
  search: '',
};

export function useContainers(options: UseContainersOptions = {}): UseContainersReturn {
  const { autoFetch = true, initialFilters = {} } = options;

  const [state, setState] = useState<ContainerListState>({
    containers: [],
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

  // Fetch containers based on current filters (using ref to break circular dependency)
  const fetchContainers = useCallback(async () => {
    // Development safety: Detect potential infinite loops
    if (process.env.NODE_ENV === 'development') {
      const now = Date.now();
      const timeSinceLastFetch = now - lastFetchTimeRef.current;

      if (timeSinceLastFetch < 100) {
        // Less than 100ms since last fetch
        fetchCountRef.current += 1;
        if (fetchCountRef.current > 10) {
          console.error(
            '🚨 INFINITE LOOP DETECTED: useContainers fetchContainers called more than 10 times in quick succession!'
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
      const response = await containerService.listContainers(filtersRef.current);
      setState((prev) => ({
        ...prev,
        containers: response.data,
        pagination: response.pagination,
        loading: false,
        initialLoading: false,
        isRefreshing: false,
      }));
    } catch (error: any) {
      const apiError = error as ApiError;
      setState((prev) => ({
        ...prev,
        error: apiError.message || 'Failed to fetch containers',
        loading: false,
        initialLoading: false,
        isRefreshing: false,
      }));
    }
  }, []); // No dependencies - breaks circular dependency

  // Set filters and trigger refetch
  const setFilters = useCallback((newFilters: ContainerFilters) => {
    setState((prev) => ({
      ...prev,
      filters: { ...prev.filters, ...newFilters, page: 1 }, // Reset to page 1 when filters change
    }));
  }, []);

  // Update single filter
  const updateFilter = useCallback((key: keyof ContainerFilters, value: any) => {
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
    await fetchContainers();
  }, [fetchContainers]);

  // Delete container and update local state
  const deleteContainer = useCallback(
    async (id: number) => {
      try {
        await containerService.deleteContainer(id);

        // If showing deleted containers, refetch to update the list
        if (filtersRef.current.show_deleted) {
          await fetchContainers();
        } else {
          // Otherwise, remove from local state
          setState((prev) => ({
            ...prev,
            containers: prev.containers.filter((container) => container.id !== id),
          }));
        }
      } catch (error: any) {
        const apiError = error as ApiError;
        setState((prev) => ({
          ...prev,
          error: apiError.message || 'Failed to delete container',
        }));
        throw error; // Re-throw for component error handling
      }
    },
    [fetchContainers] // Keep fetchContainers dependency but it's now stable
  );

  // Refresh single container
  const refreshContainer = useCallback(async (id: number) => {
    try {
      const updatedContainer = await containerService.getContainer(id);
      setState((prev) => ({
        ...prev,
        containers: prev.containers.map((container) =>
          container.id === id ? updatedContainer : container
        ),
      }));
    } catch (error: any) {
      // If container not found (404), remove from list
      if (error.status === 404) {
        setState((prev) => ({
          ...prev,
          containers: prev.containers.filter((container) => container.id !== id),
        }));
      } else {
        const apiError = error as ApiError;
        setState((prev) => ({
          ...prev,
          error: apiError.message || 'Failed to refresh container',
        }));
      }
    }
  }, []);

  // Auto-fetch on mount and when filters change
  useEffect(() => {
    if (autoFetch) {
      fetchContainers();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [autoFetch, state.filters]); // Depend on filters directly, not fetchContainers (intentionally breaking circular dependency)

  // Memoized return value to prevent unnecessary re-renders
  const returnValue = useMemo(
    () => ({
      containers: state.containers,
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
      deleteContainer,
      refreshContainer,
    }),
    [
      state.containers,
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
      deleteContainer,
      refreshContainer,
    ]
  );

  return returnValue;
}

// Helper hook for searching containers
export function useContainerSearch(containers: Container[], searchTerm: string) {
  return useMemo(() => {
    if (!searchTerm.trim()) return containers;

    const lowerSearchTerm = searchTerm.toLowerCase();
    return containers.filter(
      (container) =>
        container.name.toLowerCase().includes(lowerSearchTerm) ||
        container.description.toLowerCase().includes(lowerSearchTerm) ||
        container.author.toLowerCase().includes(lowerSearchTerm)
    );
  }, [containers, searchTerm]);
}

// Helper hook for sorting containers
export function useContainerSort(
  containers: Container[],
  sortBy: string,
  sortDirection: 'asc' | 'desc'
) {
  return useMemo(() => {
    const sortedContainers = [...containers].sort((a, b) => {
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

    return sortedContainers;
  }, [containers, sortBy, sortDirection]);
}
