// Service management type definitions
import { ApiError, PaginationInfo, SortOption } from './container';

// Re-export imported types for use in service-related modules
export type { ApiError, PaginationInfo, SortOption };

export interface Service {
  id: number;
  name: string;
  description: string;
  active: boolean;
  user_id: number;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
  containers?: ServiceContainer[];
}

export interface ServiceContainer {
  id: number;
  service_id: number;
  container_id: number;
  container_version: string;
  variables?: Record<string, any>;
  order: number;
  created_at: string;
  updated_at: string;
  container?: {
    id: number;
    name: string;
    description: string;
    author: string;
    repository: string;
    active: boolean;
  };
}

// API Request Types
export interface CreateServiceRequest {
  name: string;
  description?: string;
}

export interface UpdateServiceRequest {
  description?: string;
  active?: boolean;
}

export interface AddContainerToServiceRequest {
  container_id: number;
  container_version: string;
  variables?: Record<string, any>;
  order?: number;
}

export interface UpdateServiceContainerRequest {
  container_version?: string;
  variables?: Record<string, any>;
  order?: number;
}

export interface ValidateServiceRequest {
  // Service validation parameters
}

export interface BuildServiceRequest {
  // Service build parameters
}

// API Response Types
export interface ServiceListResponse {
  data: Service[];
  pagination: PaginationInfo;
}

export interface ServiceContainerListResponse {
  data: ServiceContainer[];
}

// PaginationInfo is imported from './container'

// Query Parameters
export interface ServiceFilters {
  page?: number;
  page_size?: number;
  active?: boolean;
  search?: string;
}

// UI State Types
export interface ServiceListState {
  services: Service[];
  loading: boolean;
  initialLoading: boolean;
  isRefreshing: boolean;
  error: string | null;
  pagination: PaginationInfo | null;
  filters: ServiceFilters;
}

export interface ServiceContainerState {
  containers: ServiceContainer[];
  loading: boolean;
  error: string | null;
}

// Form State Types
export interface ServiceFormState {
  name: string;
  description: string;
  active: boolean;
}

export interface ServiceContainerFormState {
  container_id: number;
  container_version: string;
  variables: Record<string, any>;
  order: number;
}

// Container Selection Types
export interface ContainerOption {
  id: number;
  name: string;
  description: string;
  author: string;
  versions: ContainerVersionOption[];
}

export interface ContainerVersionOption {
  version: string;
  published: boolean;
  created_at: string;
}

// ApiError is imported from './container'

// Status Enums
export enum ServiceStatus {
  Active = 'active',
  Inactive = 'inactive',
  Deleted = 'deleted',
}

// SortOption is imported from './container'

export const SERVICE_SORT_OPTIONS: SortOption[] = [
  { key: 'name', label: 'Name', direction: 'asc' },
  { key: 'created_at', label: 'Created Date', direction: 'desc' },
  { key: 'updated_at', label: 'Updated Date', direction: 'desc' },
];

// Container selector types
export interface ContainerSelectorProps {
  onSelectionChange: (selectedContainers: ServiceContainerFormState[]) => void;
  initialSelection?: ServiceContainerFormState[];
  disabled?: boolean;
}

export interface ContainerSearchFilters {
  search?: string;
  published_only?: boolean;
  active?: boolean;
}